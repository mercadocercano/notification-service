package input

import (
	"context"
	"notification-service/src/notification/domain"
)

type NotificationService interface {
	SendNotification(ctx context.Context, notification *domain.Notification) error
	SendNotificationAsync(ctx context.Context, notification *domain.Notification) error
	RetryFailedNotifications(ctx context.Context) error
	GetNotificationStatus(ctx context.Context, id string) (*domain.Notification, error)
} 