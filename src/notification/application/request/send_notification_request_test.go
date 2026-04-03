package request_test

import (
	"testing"

	"notification-service/src/notification/application/request"

	"github.com/stretchr/testify/assert"
)

func TestSendNotificationRequest_Validate_WithValidAction_ReturnsNil(t *testing.T) {
	validActions := []string{
		"WELCOME",
		"EMAIL_VERIFICATION",
		"PASSWORD_RESET",
		"ORDER_CONFIRMATION",
		"SHIPPING_NOTIFICATION",
		"ORDER_CANCELLATION",
		"PAYMENT_REMINDER",
	}

	for _, action := range validActions {
		t.Run(action, func(t *testing.T) {
			req := &request.SendNotificationRequest{
				Type:      "email",
				Action:    action,
				Recipient: "user@example.com",
			}

			err := req.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestSendNotificationRequest_Validate_WithInvalidAction_ReturnsError(t *testing.T) {
	req := &request.SendNotificationRequest{
		Type:      "email",
		Action:    "INVALID_ACTION",
		Recipient: "user@example.com",
	}

	err := req.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no válida")
}

func TestSendNotificationRequest_Validate_WithEmptyAction_ReturnsError(t *testing.T) {
	req := &request.SendNotificationRequest{
		Type:      "email",
		Action:    "",
		Recipient: "user@example.com",
	}

	err := req.Validate()

	assert.Error(t, err)
}
