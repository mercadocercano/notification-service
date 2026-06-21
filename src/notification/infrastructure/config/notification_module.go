package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"notification-service/pkg/validator"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/infrastructure/controller"
	"notification-service/src/notification/infrastructure/dedup"
	"notification-service/src/notification/infrastructure/email"
	notificationevent "notification-service/src/notification/infrastructure/event"
	notificationlog "notification-service/src/notification/infrastructure/logging"
	"notification-service/src/notification/infrastructure/queue"
	"notification-service/src/notification/infrastructure/repository"
	"notification-service/src/notification/infrastructure/template"
	"notification-service/src/notification/infrastructure/worker"
	"notification-service/src/notification/ports/output"
	"notification-service/src/shared/config"

	notificationservice "notification-service"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/postgres"
	sharedmigrate "github.com/hornosg/go-shared/migrate"
	"github.com/mercadocercano/eventbus"
	"github.com/redis/go-redis/v9"
)

// SetupNotificationModule configura el módulo de notificaciones
func SetupNotificationModule(router *gin.RouterGroup, cfg *config.Config) {
	// Inicializar conexión a base de datos
	db, err := initDatabase(cfg)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		log.Printf("Template repository will be nil")
	}

	// Migraciones versionadas in-app (ADR-001) — reemplaza el migrador casero
	// (MigrationUseCase + endpoint /api/v1/migrations), fail-fast antes de servir.
	if db != nil {
		if err := sharedmigrate.RunMigrations(db, notificationservice.MigrationsFS, cfg.Database.Name); err != nil {
			log.Fatalf("Error running migrations: %v", err)
		}
	}

	// Inicializar repositorios
	var templateRepo domain.TemplateRepository
	var notificationRepo domain.NotificationRepository

	if db != nil {
		templateRepo = repository.NewPostgresTemplateRepository(db)
		notificationRepo = repository.NewPostgresNotificationRepository(db)
	}

	// Inicializar servicios
	templateService := template.NewTemplateService(templateRepo)
	emailSender := email.NewResendClient(cfg.Resend.APIKey, cfg.Resend.FromEmail, templateService)
	emailValidator := validator.NewEmailValidator()

	// Inicializar cola SQS si está habilitada
	var queueService output.Queue
	var sqsWorker *worker.SQSWorker

	if cfg.SQS.Enabled {
		log.Printf("Initializing SQS queue with URL: %s, Region: %s", cfg.SQS.QueueURL, cfg.SQS.Region)

		sqsConfig := queue.SQSConfig{
			QueueURL: cfg.SQS.QueueURL,
			Region:   cfg.SQS.Region,
		}

		queueService, err = queue.NewSQSQueue(sqsConfig)
		if err != nil {
			log.Printf("Warning: Could not initialize SQS queue: %v", err)
			log.Printf("Queue will be nil, async notifications will not work")
			queueService = nil
		} else {
			log.Printf("SQS queue initialized successfully")

			// Inicializar y arrancar el worker SQS
			sqsWorker = worker.NewSQSWorker(queueService, emailSender, notificationRepo)

			// Crear un contexto para el worker (en producción podrías querer manejarlo mejor)
			ctx := context.Background()
			sqsWorker.Start(ctx)

			log.Printf("SQS worker started successfully")
		}
	} else {
		log.Printf("SQS is disabled, async notifications will not work")
	}

	// Inicializar logger canónico de dominio (ADR-001)
	eventLogger := notificationlog.NewNotificationLogger("notification-service")

	// Deduplicador Redis (fast-path de idempotencia, patrón Xu). Nil-safe: si no hay
	// REDIS_HOST, el use case cae al backstop de DB (ExistsByDedupKey + UNIQUE index).
	var deduplicator output.Deduplicator
	if cfg.Redis.Host != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		deduplicator = dedup.NewRedisDeduplicator(redisClient, time.Hour)
		log.Printf("Redis deduplicator initialized (%s:%d)", cfg.Redis.Host, cfg.Redis.Port)
	} else {
		log.Printf("REDIS_HOST not set, dedup falls back to DB backstop only")
	}

	// Inicializar casos de uso
	sendNotificationUseCase := usecase.NewSendNotificationUseCase(
		notificationRepo,
		templateRepo,
		emailSender,
		queueService, // Ahora puede ser SQS o nil
		emailValidator,
	).WithEventLogger(eventLogger).WithDeduplicator(deduplicator)

	getNotificationUseCase := usecase.NewGetNotificationUseCase(
		notificationRepo,
	)

	listNotificationsUseCase := usecase.NewListNotificationsUseCase(
		notificationRepo,
	)

	// Inicializar y registrar handlers
	notificationHandler := controller.NewNotificationHandler(
		sendNotificationUseCase,
		getNotificationUseCase,
		listNotificationsUseCase,
	)

	// Registrar rutas de notificaciones
	notificationHandler.RegisterRoutes(router)

	// Agregar endpoint para monitoreo de la cola SQS (opcional)
	if cfg.SQS.Enabled && sqsWorker != nil {
		setupQueueMonitoringRoutes(router, sqsWorker)
	}

	// Worker consumer del EventBus (ingestión event-driven, ADR/Plan F1). Opt-in por
	// EVENTBUS_ENABLED para no colgar el pod antes de cablear los secrets EVENTBUS_DB_*.
	setupEventWorker(cfg, sendNotificationUseCase, eventLogger)
}

// setupEventWorker conecta a la DB del EventBus, registra el handler de notificaciones y
// arranca el worker (patrón ledger-service). Best-effort: si la conexión falla, loguea y
// sigue — el path HTTP sync (OTP/reset) no depende del EventBus.
func setupEventWorker(cfg *config.Config, sender notificationevent.NotificationSender, eventLogger *notificationlog.NotificationLogger) {
	if !cfg.EventBus.Enabled {
		log.Printf("EventBus consumer disabled (EVENTBUS_ENABLED != true), event-driven notifications inactive")
		return
	}

	eventbusDB, err := postgres.Connect(postgres.Config{
		Host:     cfg.EventBus.Host,
		Port:     cfg.EventBus.Port,
		User:     cfg.EventBus.User,
		Password: cfg.EventBus.Password,
		DBName:   cfg.EventBus.Name,
		SSLMode:  cfg.EventBus.SSLMode,
	})
	if err != nil {
		log.Printf("Warning: could not connect to EventBus DB, event-driven notifications disabled: %v", err)
		return
	}

	postgres.StartPoolMonitor(context.Background(), eventbusDB, postgres.MonitorOptions{
		Service: "notification-service",
		DBName:  cfg.EventBus.Name,
	})

	infraLogger := eventbus.NewLogger(eventbus.LevelInfo)
	eventStore := eventbus.NewSQLEventStore(eventbusDB, infraLogger)
	processUseCase := eventbus.NewProcessEventUseCase(eventStore, infraLogger)

	eventWorker := eventbus.NewEventWorker(
		processUseCase,
		infraLogger,
		10,            // batch size
		5*time.Second, // poll interval
	)

	handler := notificationevent.NewNotificationEventHandler(sender, eventLogger)
	if err := eventWorker.RegisterHandler(handler); err != nil {
		log.Printf("Warning: could not register notification event handler: %v", err)
		return
	}

	if err := eventWorker.Start(context.Background()); err != nil {
		log.Printf("Warning: could not start EventBus worker: %v", err)
		return
	}
	log.Printf("EventBus consumer started (consumer=%s, batch=10, poll=5s)", handler.ConsumerName())
}

// setupQueueMonitoringRoutes configura rutas para monitorear la cola SQS
func setupQueueMonitoringRoutes(router *gin.RouterGroup, sqsWorker *worker.SQSWorker) {
	router.GET("/queue/status", func(c *gin.Context) {
		ctx := c.Request.Context()

		size, err := sqsWorker.GetQueueSize(ctx)
		if err != nil {
			c.JSON(500, gin.H{
				"error":   "Failed to get queue size",
				"details": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"queue_size":     size,
			"worker_running": sqsWorker.IsRunning(),
		})
	})
}

// initDatabase inicializa la conexión a PostgreSQL
func initDatabase(cfg *config.Config) (*sql.DB, error) {
	db, err := postgres.Connect(postgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		return nil, err
	}

	postgres.StartPoolMonitor(context.Background(), db, postgres.MonitorOptions{
		Service: "notification-service",
		DBName:  cfg.Database.Name,
	})

	log.Printf("Successfully connected to PostgreSQL database")
	return db, nil
}
