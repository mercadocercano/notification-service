package config

import (
	"context"
	"database/sql"
	"log"
	"notification-service/pkg/validator"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/infrastructure/controller"
	"notification-service/src/notification/infrastructure/email"
	notificationlog "notification-service/src/notification/infrastructure/logging"
	"notification-service/src/notification/infrastructure/queue"
	"notification-service/src/notification/infrastructure/repository"
	"notification-service/src/notification/infrastructure/template"
	"notification-service/src/notification/infrastructure/worker"
	"notification-service/src/notification/ports/output"
	"notification-service/src/shared/config"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/postgres"
)

// SetupNotificationModule configura el módulo de notificaciones
func SetupNotificationModule(router *gin.RouterGroup, cfg *config.Config) {
	// Inicializar conexión a base de datos
	db, err := initDatabase(cfg)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		log.Printf("Template repository will be nil")
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

	// Inicializar casos de uso
	sendNotificationUseCase := usecase.NewSendNotificationUseCase(
		notificationRepo,
		templateRepo,
		emailSender,
		queueService, // Ahora puede ser SQS o nil
		emailValidator,
	).WithEventLogger(eventLogger)

	getNotificationUseCase := usecase.NewGetNotificationUseCase(
		notificationRepo,
	)

	listNotificationsUseCase := usecase.NewListNotificationsUseCase(
		notificationRepo,
	)

	// Inicializar caso de uso de migraciones
	var migrationUseCase *usecase.MigrationUseCase
	if db != nil {
		migrationUseCase = usecase.NewMigrationUseCase(db)
	}

	// Inicializar y registrar handlers
	notificationHandler := controller.NewNotificationHandler(
		sendNotificationUseCase,
		getNotificationUseCase,
		listNotificationsUseCase,
	)

	// Registrar handler de migraciones si la BD está disponible
	if migrationUseCase != nil {
		migrationHandler := controller.NewMigrationHandler(migrationUseCase)
		migrationHandler.RegisterRoutes(router)
	}

	// Registrar rutas de notificaciones
	notificationHandler.RegisterRoutes(router)

	// Agregar endpoint para monitoreo de la cola SQS (opcional)
	if cfg.SQS.Enabled && sqsWorker != nil {
		setupQueueMonitoringRoutes(router, sqsWorker)
	}
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
