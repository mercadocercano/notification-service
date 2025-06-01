package domain

import (
	"context"
	"time"
)

type NotificationType string
type NotificationStatus string
type NotificationAction string

const (
	EmailNotification NotificationType = "email"
	SMSNotification   NotificationType = "sms"

	StatusPending  NotificationStatus = "pending"
	StatusSent     NotificationStatus = "sent"
	StatusFailed   NotificationStatus = "failed"
	StatusRetrying NotificationStatus = "retrying"
	StatusQueued   NotificationStatus = "queued"

	// Acciones de notificación
	ActionWelcome              NotificationAction = "WELCOME"
	ActionEmailVerification    NotificationAction = "EMAIL_VERIFICATION"
	ActionPasswordReset        NotificationAction = "PASSWORD_RESET"
	ActionOrderConfirmation    NotificationAction = "ORDER_CONFIRMATION"
	ActionShippingNotification NotificationAction = "SHIPPING_NOTIFICATION"
	ActionOrderCancellation    NotificationAction = "ORDER_CANCELLATION"
	ActionPaymentReminder      NotificationAction = "PAYMENT_REMINDER"
)

type Notification struct {
	ID         string
	Type       NotificationType
	Action     NotificationAction
	TemplateID string
	Recipient  string
	Data       map[string]interface{}
	Status     NotificationStatus
	RetryCount int
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ValidActions devuelve las acciones válidas
func ValidActions() []NotificationAction {
	return []NotificationAction{
		ActionWelcome,
		ActionEmailVerification,
		ActionPasswordReset,
		ActionOrderConfirmation,
		ActionShippingNotification,
		ActionOrderCancellation,
		ActionPaymentReminder,
	}
}

// IsValidAction verifica si una acción es válida
func IsValidAction(action NotificationAction) bool {
	for _, validAction := range ValidActions() {
		if action == validAction {
			return true
		}
	}
	return false
}

type NotificationRepository interface {
	Save(ctx context.Context, notification *Notification) error
	FindByID(ctx context.Context, id string) (*Notification, error)
	Update(ctx context.Context, notification *Notification) error
	UpdateStatus(ctx context.Context, id string, status NotificationStatus, error string) error
	FindPendingNotifications(ctx context.Context) ([]*Notification, error)
}
