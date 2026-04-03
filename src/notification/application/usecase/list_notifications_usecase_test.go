package usecase_test

import (
	"context"
	"errors"
	"testing"

	"notification-service/src/notification/application/request"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/domain/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListNotificationsUseCase_Execute_WithValidRequest_ReturnsNotifications(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()
	notifMother := testutil.NewNotificationMother()

	notifications := []*domain.Notification{
		notifMother.Sent(),
		notifMother.WithStatus(domain.StatusPending),
	}

	notifRepo.On("FindByFilters", mock.Anything, mock.AnythingOfType("domain.NotificationFilters")).
		Return(notifications, nil)

	uc := usecase.NewListNotificationsUseCase(notifRepo)
	req := &request.ListNotificationsRequest{
		Page:  1,
		Limit: 20,
	}

	// Act
	result, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Notifications, 2)
	assert.Equal(t, "notif-001", result.Notifications[0].ID)
}

func TestListNotificationsUseCase_Execute_WithTypeFilter_PassesFilterToRepo(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()

	notifRepo.On("FindByFilters", mock.Anything, mock.MatchedBy(func(f domain.NotificationFilters) bool {
		return f.Type != nil && *f.Type == domain.EmailNotification
	})).Return([]*domain.Notification{}, nil)

	uc := usecase.NewListNotificationsUseCase(notifRepo)
	req := &request.ListNotificationsRequest{
		Type:  "email",
		Page:  1,
		Limit: 20,
	}

	// Act
	result, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result.Notifications)
	assert.Equal(t, "email", result.Filters.Type)
	notifRepo.AssertExpectations(t)
}

func TestListNotificationsUseCase_Execute_WithInvalidType_ReturnsError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	uc := usecase.NewListNotificationsUseCase(notifRepo)

	req := &request.ListNotificationsRequest{
		Type: "push",
	}

	result, err := uc.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "solicitud inválida")
}

func TestListNotificationsUseCase_Execute_WithInvalidAction_ReturnsError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	uc := usecase.NewListNotificationsUseCase(notifRepo)

	req := &request.ListNotificationsRequest{
		Action: "NONEXISTENT",
	}

	result, err := uc.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestListNotificationsUseCase_Execute_WithInvalidStatus_ReturnsError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	uc := usecase.NewListNotificationsUseCase(notifRepo)

	req := &request.ListNotificationsRequest{
		Status: "invalid",
	}

	result, err := uc.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestListNotificationsUseCase_Execute_WhenRepoFails_ReturnsError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()

	notifRepo.On("FindByFilters", mock.Anything, mock.AnythingOfType("domain.NotificationFilters")).
		Return(nil, errors.New("db connection failed"))

	uc := usecase.NewListNotificationsUseCase(notifRepo)
	req := &request.ListNotificationsRequest{
		Page:  1,
		Limit: 20,
	}

	result, err := uc.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error buscando notificaciones")
}

func TestListNotificationsUseCase_Execute_WithNilRepo_ReturnsError(t *testing.T) {
	uc := usecase.NewListNotificationsUseCase(nil)
	req := &request.ListNotificationsRequest{
		Page:  1,
		Limit: 20,
	}

	result, err := uc.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "base de datos no disponible")
}

func TestListNotificationsUseCase_Execute_ReturnsPaginationMetadata(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	notifMother := testutil.NewNotificationMother()

	// Retornar exactamente el limite para que HasMore sea true
	notifications := make([]*domain.Notification, 10)
	for i := range notifications {
		notifications[i] = notifMother.Default()
	}

	notifRepo.On("FindByFilters", mock.Anything, mock.AnythingOfType("domain.NotificationFilters")).
		Return(notifications, nil)

	uc := usecase.NewListNotificationsUseCase(notifRepo)
	req := &request.ListNotificationsRequest{
		Page:  2,
		Limit: 10,
	}

	result, err := uc.Execute(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 10, result.Pagination.Total)
	assert.True(t, result.Pagination.HasMore)
}

func TestListNotificationsUseCase_Execute_WithFewerResultsThanLimit_HasMoreIsFalse(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	notifMother := testutil.NewNotificationMother()

	notifications := []*domain.Notification{
		notifMother.Default(),
	}

	notifRepo.On("FindByFilters", mock.Anything, mock.AnythingOfType("domain.NotificationFilters")).
		Return(notifications, nil)

	uc := usecase.NewListNotificationsUseCase(notifRepo)
	req := &request.ListNotificationsRequest{
		Page:  1,
		Limit: 20,
	}

	result, err := uc.Execute(context.Background(), req)

	require.NoError(t, err)
	assert.False(t, result.Pagination.HasMore)
}

func TestToNotificationDTO_MapsAllFields(t *testing.T) {
	notifMother := testutil.NewNotificationMother()
	notification := notifMother.Default()

	dto := usecase.ToNotificationDTO(notification)

	assert.Equal(t, notification.ID, dto.ID)
	assert.Equal(t, string(notification.Type), dto.Type)
	assert.Equal(t, string(notification.Action), dto.Action)
	assert.Equal(t, notification.Recipient, dto.Recipient)
	assert.Equal(t, string(notification.Status), dto.Status)
	assert.Equal(t, notification.Data, dto.Data)
	assert.NotEmpty(t, dto.CreatedAt)
	assert.NotEmpty(t, dto.UpdatedAt)
}

func TestListNotificationsUseCase_Execute_WithFiltersSummary_ReturnsAppliedFilters(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()

	notifRepo.On("FindByFilters", mock.Anything, mock.AnythingOfType("domain.NotificationFilters")).
		Return([]*domain.Notification{}, nil)

	uc := usecase.NewListNotificationsUseCase(notifRepo)
	req := &request.ListNotificationsRequest{
		Type:      "email",
		Action:    "WELCOME",
		Recipient: "test@example.com",
		Status:    "sent",
		Page:      1,
		Limit:     20,
	}

	result, err := uc.Execute(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "email", result.Filters.Type)
	assert.Equal(t, "WELCOME", result.Filters.Action)
	assert.Equal(t, "test@example.com", result.Filters.Recipient)
	assert.Equal(t, "sent", result.Filters.Status)
}
