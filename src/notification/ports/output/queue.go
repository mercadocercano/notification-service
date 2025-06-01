package output

import (
	"context"
	"notification-service/src/notification/domain"
)

type Queue interface {
	Enqueue(ctx context.Context, notification *domain.Notification) error
	Dequeue(ctx context.Context) (*domain.Notification, error)
	Size(ctx context.Context) (int, error)
} 