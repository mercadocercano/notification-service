package domain_test

import (
	"testing"

	"notification-service/src/notification/domain"

	"github.com/stretchr/testify/assert"
)

func TestIsValidAction_WithValidAction_ReturnsTrue(t *testing.T) {
	validActions := []domain.NotificationAction{
		domain.ActionWelcome,
		domain.ActionEmailVerification,
		domain.ActionPasswordReset,
		domain.ActionOrderConfirmation,
		domain.ActionShippingNotification,
		domain.ActionOrderCancellation,
		domain.ActionPaymentReminder,
	}

	for _, action := range validActions {
		t.Run(string(action), func(t *testing.T) {
			assert.True(t, domain.IsValidAction(action), "action %s should be valid", action)
		})
	}
}

func TestIsValidAction_WithInvalidAction_ReturnsFalse(t *testing.T) {
	invalidActions := []domain.NotificationAction{
		"INVALID",
		"",
		"random_action",
		"welcome", // lowercase, real is WELCOME
	}

	for _, action := range invalidActions {
		t.Run(string(action), func(t *testing.T) {
			assert.False(t, domain.IsValidAction(action), "action %q should be invalid", action)
		})
	}
}

func TestValidActions_ReturnsAllExpectedActions(t *testing.T) {
	actions := domain.ValidActions()

	assert.Len(t, actions, 7)
	assert.Contains(t, actions, domain.ActionWelcome)
	assert.Contains(t, actions, domain.ActionEmailVerification)
	assert.Contains(t, actions, domain.ActionPasswordReset)
	assert.Contains(t, actions, domain.ActionOrderConfirmation)
	assert.Contains(t, actions, domain.ActionShippingNotification)
	assert.Contains(t, actions, domain.ActionOrderCancellation)
	assert.Contains(t, actions, domain.ActionPaymentReminder)
}

func TestNotificationType_Constants_HaveExpectedValues(t *testing.T) {
	assert.Equal(t, domain.NotificationType("email"), domain.EmailNotification)
	assert.Equal(t, domain.NotificationType("sms"), domain.SMSNotification)
}

func TestNotificationStatus_Constants_HaveExpectedValues(t *testing.T) {
	assert.Equal(t, domain.NotificationStatus("pending"), domain.StatusPending)
	assert.Equal(t, domain.NotificationStatus("sent"), domain.StatusSent)
	assert.Equal(t, domain.NotificationStatus("failed"), domain.StatusFailed)
	assert.Equal(t, domain.NotificationStatus("retrying"), domain.StatusRetrying)
	assert.Equal(t, domain.NotificationStatus("queued"), domain.StatusQueued)
}
