package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"notification-service/src/notification/application/request"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/shared/logger"
	"notification-service/src/shared/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	sendNotificationUseCase *usecase.SendNotificationUseCase
	getNotificationUseCase  *usecase.GetNotificationUseCase
}

func NewNotificationHandler(
	sendNotificationUseCase *usecase.SendNotificationUseCase,
	getNotificationUseCase *usecase.GetNotificationUseCase,
) *NotificationHandler {
	return &NotificationHandler{
		sendNotificationUseCase: sendNotificationUseCase,
		getNotificationUseCase:  getNotificationUseCase,
	}
}

// SendNotification godoc
// @Summary Send notification
// @Description Send a notification via email or other channels
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body request.SendNotificationRequest true "Send notification request"
// @Success 200 {object} response.SendNotificationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications [post]
func (handler *NotificationHandler) SendNotification(ctx *gin.Context) {
	log := logger.GetLogger()

	// Log del body raw
	body, _ := ctx.GetRawData()
	log.Info("Raw request body", zap.String("body", string(body)))

	// Rebuild el context para que ShouldBindJSON pueda leer el body
	ctx.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	var request request.SendNotificationRequest

	// Validación de binding JSON
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// Log detallado del error
		log.Error("Binding error details",
			zap.Error(err),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.String("error_string", err.Error()))
		middleware.AbortWithBusinessError(ctx, middleware.ErrInvalidRequestFormat)
		return
	}

	log.Info("Request bound successfully",
		zap.String("type", request.Type),
		zap.String("action", request.Action),
		zap.String("recipient", request.Recipient))

	// Ejecutar caso de uso
	result := handler.sendNotificationUseCase.Execute(ctx.Request.Context(), &request)

	// Si hay error en el resultado, usar el middleware
	if !result.Success {
		middleware.AbortWithBusinessError(ctx, result.ToMiddlewareError())
		return
	}

	// Respuesta exitosa
	ctx.JSON(http.StatusOK, result.Data)
}

// GetNotificationStatus godoc
// @Summary Get notification status
// @Description Get the status of a specific notification by ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.GetNotificationResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /notifications/{id} [get]
func (handler *NotificationHandler) GetNotificationStatus(ctx *gin.Context) {
	notificationID := ctx.Param("id")
	if notificationID == "" {
		middleware.AbortWithBusinessError(ctx, middleware.BusinessError{
			Code:       "MISSING_NOTIFICATION_ID",
			Message:    "ID de notificación requerido",
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	response, err := handler.getNotificationUseCase.Execute(ctx.Request.Context(), notificationID)
	if err != nil {
		switch err {
		case usecase.ErrNotificationNotFound:
			middleware.AbortWithBusinessError(ctx, middleware.ErrNotificationNotFound)
		default:
			middleware.AbortWithError(ctx, err)
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// RegisterRoutes registra las rutas del módulo notifications
func (handler *NotificationHandler) RegisterRoutes(router *gin.RouterGroup) {
	notificationsGroup := router.Group("/notifications")
	{
		notificationsGroup.POST("", handler.SendNotification)
		notificationsGroup.GET("/:id", handler.GetNotificationStatus)
	}
}
