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
	Namespace  string // proyecto (IDP). Default 'mc'. Scope de nivel superior.
	TenantID   string // tenant dentro del proyecto (puede ser vacío para notifs de plataforma).
	Type       NotificationType
	Action     NotificationAction
	TemplateID string
	Recipient  string
	Data       map[string]interface{}
	Status     NotificationStatus
	RetryCount int
	Error      string
	DedupKey   string // idempotencia: event_id (eventos) o Idempotency-Key/hash (API sync). Vacío = sin dedup.
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
	FindByFilters(ctx context.Context, filters NotificationFilters) ([]*Notification, error)
	// ExistsByDedupKey es el backstop de idempotencia en DB (respaldo del UNIQUE index y
	// del nonce de Redis). Scopeado por (namespace, tenant_id). dedupKey vacío → siempre false.
	ExistsByDedupKey(ctx context.Context, namespace, tenantID, dedupKey string) (bool, error)
}

// NotificationFilters para filtrar notificaciones
type NotificationFilters struct {
	Type      *NotificationType   `json:"type,omitempty"`
	Action    *NotificationAction `json:"action,omitempty"`
	Recipient *string             `json:"recipient,omitempty"`
	Status    *NotificationStatus `json:"status,omitempty"`
	Limit     int                 `json:"limit,omitempty"`
	Offset    int                 `json:"offset,omitempty"`
}
