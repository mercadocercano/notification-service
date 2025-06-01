package controller

import (
	"net/http"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/shared/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MigrationHandler struct {
	migrationUseCase *usecase.MigrationUseCase
}

func NewMigrationHandler(migrationUseCase *usecase.MigrationUseCase) *MigrationHandler {
	return &MigrationHandler{
		migrationUseCase: migrationUseCase,
	}
}

// RunMigrations godoc
// @Summary Run database migrations
// @Description Execute pending database migrations
// @Tags migrations
// @Accept json
// @Produce json
// @Success 200 {object} usecase.MigrationResult
// @Failure 500 {object} usecase.MigrationResult
// @Router /migrations [post]
func (handler *MigrationHandler) RunMigrations(ctx *gin.Context) {
	log := logger.GetLogger()

	log.Info("Migrations endpoint called")

	// Ejecutar migraciones
	result := handler.migrationUseCase.RunMigrations(ctx.Request.Context())

	// Determinar código de respuesta HTTP
	httpStatus := http.StatusOK
	if !result.Success {
		httpStatus = http.StatusInternalServerError
		log.Error("Migration failed", zap.String("error", result.Error))
	} else {
		log.Info("Migration completed successfully",
			zap.String("message", result.Message),
			zap.Strings("executed", result.ExecutedMigrations))
	}

	ctx.JSON(httpStatus, result)
}

// GetMigrationStatus godoc
// @Summary Get migration status
// @Description Get the current status of database migrations
// @Tags migrations
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /migrations/status [get]
func (handler *MigrationHandler) GetMigrationStatus(ctx *gin.Context) {
	log := logger.GetLogger()

	log.Info("Migration status endpoint called")

	// Por ahora retornamos información básica
	// En el futuro se podría extender para mostrar estado detallado
	status := map[string]interface{}{
		"status":  "ready",
		"message": "Use POST /migrations to run pending migrations",
		"endpoints": map[string]string{
			"run_migrations": "POST /api/v1/migrations",
			"status":         "GET /api/v1/migrations/status",
		},
	}

	ctx.JSON(http.StatusOK, status)
}

// RegisterRoutes registra las rutas del módulo migrations
func (handler *MigrationHandler) RegisterRoutes(router *gin.RouterGroup) {
	migrationsGroup := router.Group("/migrations")
	{
		migrationsGroup.POST("", handler.RunMigrations)
		migrationsGroup.GET("/status", handler.GetMigrationStatus)
	}
}
