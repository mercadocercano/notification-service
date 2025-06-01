package request

import (
	"errors"
	"notification-service/src/notification/domain"
)

type SendNotificationRequest struct {
	NotificationID string                 `json:"notification_id,omitempty"` // Solo para SQS
	Type           string                 `json:"type" binding:"required"`
	Action         string                 `json:"action" binding:"required"`
	Recipient      string                 `json:"recipient" binding:"required"`
	Data           map[string]interface{} `json:"data"`
	Async          bool                   `json:"async"`
}

// Validate realiza validaciones adicionales de negocio
func (r *SendNotificationRequest) Validate() error {
	// Validar que la acción sea válida
	action := domain.NotificationAction(r.Action)
	if !domain.IsValidAction(action) {
		return errors.New("acción no válida")
	}

	return nil
}
