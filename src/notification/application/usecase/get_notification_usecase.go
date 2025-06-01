package usecase

import (
	"context"
	"errors"

	"notification-service/src/notification/application/response"
	"notification-service/src/notification/domain"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type GetNotificationUseCase struct {
	notificationRepo domain.NotificationRepository
}

func NewGetNotificationUseCase(
	notificationRepo domain.NotificationRepository,
) *GetNotificationUseCase {
	return &GetNotificationUseCase{
		notificationRepo: notificationRepo,
	}
}

func (uc *GetNotificationUseCase) Execute(ctx context.Context, notificationID string) (*response.GetNotificationResponse, error) {
	if uc.notificationRepo == nil {
		return nil, ErrNotificationNotFound
	}

	notification, err := uc.notificationRepo.FindByID(ctx, notificationID)
	if err != nil {
		return nil, ErrNotificationNotFound
	}

	return &response.GetNotificationResponse{
		ID:        notification.ID,
		Type:      string(notification.Type),
		Recipient: notification.Recipient,
		Status:    string(notification.Status),
		Data:      notification.Data,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}, nil
} 