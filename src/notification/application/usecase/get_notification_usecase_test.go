package usecase_test

import (
	"context"
	"errors"
	"testing"

	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetNotificationUseCase_Execute_WithExistingID_ReturnsNotification(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()
	notifMother := testutil.NewNotificationMother()
	notification := notifMother.Sent()

	notifRepo.On("FindByID", mock.Anything, "notif-001").
		Return(notification, nil)

	uc := usecase.NewGetNotificationUseCase(notifRepo)

	// Act
	result, err := uc.Execute(context.Background(), "notif-001")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "notif-001", result.ID)
	assert.Equal(t, "email", result.Type)
	assert.Equal(t, "user@example.com", result.Recipient)
	assert.Equal(t, "sent", result.Status)
	assert.Equal(t, notification.Data, result.Data)
}

func TestGetNotificationUseCase_Execute_WithNonExistingID_ReturnsError(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()
	notifRepo.On("FindByID", mock.Anything, "nonexistent").
		Return(nil, errors.New("not found"))

	uc := usecase.NewGetNotificationUseCase(notifRepo)

	// Act
	result, err := uc.Execute(context.Background(), "nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, usecase.ErrNotificationNotFound, err)
}

func TestGetNotificationUseCase_Execute_WithNilRepo_ReturnsError(t *testing.T) {
	uc := usecase.NewGetNotificationUseCase(nil)

	result, err := uc.Execute(context.Background(), "any-id")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, usecase.ErrNotificationNotFound, err)
}

func TestGetNotificationUseCase_Execute_MapsAllFieldsCorrectly(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()
	notifMother := testutil.NewNotificationMother()
	notification := notifMother.Default()

	notifRepo.On("FindByID", mock.Anything, notification.ID).
		Return(notification, nil)

	uc := usecase.NewGetNotificationUseCase(notifRepo)

	// Act
	result, err := uc.Execute(context.Background(), notification.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, notification.ID, result.ID)
	assert.Equal(t, string(notification.Type), result.Type)
	assert.Equal(t, notification.Recipient, result.Recipient)
	assert.Equal(t, string(notification.Status), result.Status)
	assert.Equal(t, notification.CreatedAt, result.CreatedAt)
	assert.Equal(t, notification.UpdatedAt, result.UpdatedAt)
}
