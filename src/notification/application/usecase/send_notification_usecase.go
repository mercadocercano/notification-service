package usecase

import (
	"context"
	"time"

	"notification-service/pkg/validator"
	"notification-service/src/notification/application/request"
	"notification-service/src/notification/application/response"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/ports/output"
	"notification-service/src/shared/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

type SendNotificationUseCase struct {
	notificationRepo domain.NotificationRepository
	templateRepo     domain.TemplateRepository
	emailSender      output.EmailSender
	queue            output.Queue
	emailValidator   *validator.EmailValidator
}

func NewSendNotificationUseCase(
	notificationRepo domain.NotificationRepository,
	templateRepo domain.TemplateRepository,
	emailSender output.EmailSender,
	queue output.Queue,
	emailValidator *validator.EmailValidator,
) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		emailSender:      emailSender,
		queue:            queue,
		emailValidator:   emailValidator,
	}
}

func (uc *SendNotificationUseCase) Execute(ctx context.Context, req *request.SendNotificationRequest) *response.SendNotificationResult {

	log.Info("Starting notification processing",
		zap.String("type", req.Type),
		zap.String("recipient", req.Recipient),
		zap.String("action", req.Action),
		zap.Bool("async", req.Async))

	// Validar request
	if validationResult := uc.validateRequest(req); validationResult != nil {
		log.Warn("Request validation failed",
			zap.String("error_code", validationResult.Error.Code),
			zap.String("error_message", validationResult.Error.Message))
		return validationResult
	}

	// 2. Obtener y validar template
	template, templateResult := uc.getTemplate(ctx, req.Action, req.Type)
	if templateResult != nil {
		return templateResult
	}

	// Log apropiado dependiendo de si tenemos template o no
	if template != nil {
		log.Info("Template found and validated",
			zap.String("action", req.Action),
			zap.String("template_id", template.ID),
			zap.String("template_name", template.Name))
	} else {
		log.Info("Template validation skipped - using fallback mechanism",
			zap.String("action", req.Action),
			zap.String("type", req.Type))
	}

	// 3. Crear notificación
	notification := &domain.Notification{
		ID:        uuid.New().String(),
		Type:      domain.NotificationType(req.Type),
		Action:    domain.NotificationAction(req.Action),
		Recipient: req.Recipient,
		Status:    domain.StatusPending,
		Data:      req.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	log.Info("Notification created",
		zap.String("notification_id", notification.ID),
		zap.String("action", string(notification.Action)),
		zap.String("status", string(notification.Status)))

	// 4. Guardar notificación
	if uc.notificationRepo != nil {
		log.Debug("Saving notification to repository")
		if err := uc.notificationRepo.Save(ctx, notification); err != nil {
			log.Error("Failed to save notification",
				zap.String("notification_id", notification.ID),
				zap.Error(err))
			return response.NewInternalServerError("Error al guardar notificación: " + err.Error())
		}
		log.Debug("Notification saved successfully")
	} else {
		log.Warn("Notification repository is nil, skipping save")
	}

	// 5. Procesar notificación
	if req.Async {
		log.Info("Processing notification asynchronously")
		// Envío asíncrono
		if uc.queue != nil {
			log.Debug("Enqueueing notification")
			if err := uc.queue.Enqueue(ctx, notification); err != nil {
				log.Error("Failed to enqueue notification",
					zap.String("notification_id", notification.ID),
					zap.Error(err))
				return response.NewInternalServerError("Error al encolar notificación: " + err.Error())
			}
			log.Debug("Notification enqueued successfully")
		} else {
			log.Warn("Queue is nil, cannot enqueue notification")
		}
		notification.Status = domain.StatusQueued
	} else {
		log.Info("Processing notification synchronously")
		// Envío síncrono
		if err := uc.sendNotification(ctx, notification); err != nil {
			log.Error("Failed to send notification",
				zap.String("notification_id", notification.ID),
				zap.String("action", string(notification.Action)),
				zap.Error(err))
			notification.Status = domain.StatusFailed
			if uc.notificationRepo != nil {
				uc.notificationRepo.Update(ctx, notification)
			}
			return response.NewInternalServerError("Error al enviar notificación: " + err.Error())
		}
		log.Info("Notification sent successfully", zap.String("notification_id", notification.ID))
		notification.Status = domain.StatusSent
	}

	// 6. Actualizar estado
	if uc.notificationRepo != nil {
		log.Debug("Updating notification status",
			zap.String("notification_id", notification.ID),
			zap.String("status", string(notification.Status)))
		if err := uc.notificationRepo.Update(ctx, notification); err != nil {
			log.Error("Failed to update notification status",
				zap.String("notification_id", notification.ID),
				zap.Error(err))
			return response.NewInternalServerError("Error al actualizar notificación: " + err.Error())
		}
		log.Debug("Notification status updated successfully")
	}

	// 7. Devolver resultado exitoso
	responseData := &response.SendNotificationResponse{
		ID:        notification.ID,
		Message:   "Notification sent successfully",
		Status:    string(notification.Status),
		CreatedAt: notification.CreatedAt,
	}

	log.Info("Notification processing completed successfully",
		zap.String("notification_id", notification.ID),
		zap.String("final_status", string(notification.Status)))

	return response.NewSendNotificationSuccess(responseData)
}

func (uc *SendNotificationUseCase) validateRequest(req *request.SendNotificationRequest) *response.SendNotificationResult {

	if req.Type == "" {
		log.Warn("Missing notification type")
		return response.NewInvalidNotificationTypeError()
	}

	if req.Type != "email" {
		log.Warn("Invalid notification type", zap.String("type", req.Type))
		return response.NewInvalidNotificationTypeError()
	}

	if req.Recipient == "" {
		log.Warn("Missing recipient")
		return response.NewInvalidEmailError()
	}

	if !uc.emailValidator.IsValid(req.Recipient) {
		log.Warn("Invalid email format", zap.String("recipient", req.Recipient))
		return response.NewInvalidEmailError()
	}

	// Validar action
	if req.Action == "" {
		log.Warn("Missing action")
		return response.NewInvalidNotificationTypeError()
	}

	action := domain.NotificationAction(req.Action)
	if !domain.IsValidAction(action) {
		log.Warn("Invalid action", zap.String("action", req.Action))
		return response.NewInvalidNotificationTypeError()
	}

	log.Debug("Request validation passed",
		zap.String("type", req.Type),
		zap.String("recipient", req.Recipient),
		zap.String("action", req.Action))

	return nil
}

func (uc *SendNotificationUseCase) getTemplate(ctx context.Context, action string, notificationType string) (*domain.Template, *response.SendNotificationResult) {

	log.Debug("Validating template existence",
		zap.String("action", action),
		zap.String("type", notificationType))

	if uc.templateRepo != nil {
		actionEnum := domain.NotificationAction(action)
		notificationTypeEnum := domain.NotificationType(notificationType)
		template, err := uc.templateRepo.FindByAction(ctx, actionEnum, notificationTypeEnum)
		if err != nil {
			log.Error("Error searching template in database",
				zap.String("action", action),
				zap.Error(err))
			return nil, response.NewInternalServerError("Error al buscar template: " + err.Error())
		}

		if template == nil {
			log.Warn("Template not found for action",
				zap.String("action", action),
				zap.String("type", notificationType))
			return nil, response.NewTemplateNotFoundError()
		}

		return template, nil
	} else {
		log.Warn("Template repository is nil, skipping template validation")
		return nil, nil
	}
}

func (uc *SendNotificationUseCase) sendNotification(ctx context.Context, notification *domain.Notification) error {

	log.Debug("Sending notification",
		zap.String("notification_id", notification.ID),
		zap.String("type", string(notification.Type)),
		zap.String("action", string(notification.Action)),
		zap.String("recipient", notification.Recipient))

	switch notification.Type {
	case domain.EmailNotification:
		return uc.emailSender.SendEmailByAction(ctx, notification.Recipient, notification.Action, notification.Type, notification.Data)
	default:
		log.Warn("Unknown notification type", zap.String("type", string(notification.Type)))
		return nil // Ya validado anteriormente
	}
}
