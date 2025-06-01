package config

import (
	"database/sql"
	"log"
	"notification-service/pkg/validator"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/infrastructure/controller"
	"notification-service/src/notification/infrastructure/email"
	"notification-service/src/notification/infrastructure/repository"
	"notification-service/src/notification/infrastructure/template"
	"notification-service/src/shared/config"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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

	// Inicializar casos de uso
	sendNotificationUseCase := usecase.NewSendNotificationUseCase(
		notificationRepo,
		templateRepo,
		emailSender,
		nil, // queue - por ahora nil
		emailValidator,
	)

	getNotificationUseCase := usecase.NewGetNotificationUseCase(
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
	)

	// Registrar handler de migraciones si la BD está disponible
	if migrationUseCase != nil {
		migrationHandler := controller.NewMigrationHandler(migrationUseCase)
		migrationHandler.RegisterRoutes(router)
	}

	// Registrar rutas de notificaciones
	notificationHandler.RegisterRoutes(router)
}

// initDatabase inicializa la conexión a PostgreSQL
func initDatabase(cfg *config.Config) (*sql.DB, error) {
	// Construir string de conexión
	connStr := "host=" + cfg.Database.Host +
		" port=" + cfg.Database.Port +
		" user=" + cfg.Database.User +
		" password=" + cfg.Database.Password +
		" dbname=" + cfg.Database.Name +
		" sslmode=" + cfg.Database.SSLMode

	// Abrir conexión
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verificar conexión
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	log.Printf("Successfully connected to PostgreSQL database")
	return db, nil
}
