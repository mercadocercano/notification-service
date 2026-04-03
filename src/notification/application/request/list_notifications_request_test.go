package request_test

import (
	"testing"

	"notification-service/src/notification/application/request"
	"notification-service/src/notification/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNotificationsRequest_Validate_WithValidRequest_ReturnsNil(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Type:   "email",
		Status: "sent",
		Page:   1,
		Limit:  20,
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestListNotificationsRequest_Validate_WithEmptyFilters_ReturnsNil(t *testing.T) {
	req := &request.ListNotificationsRequest{}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestListNotificationsRequest_Validate_WithInvalidType_ReturnsError(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Type: "push",
	}

	err := req.Validate()

	require.Error(t, err)
	var validationErr request.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "type", validationErr.Field)
}

func TestListNotificationsRequest_Validate_WithInvalidStatus_ReturnsError(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Status: "invalid_status",
	}

	err := req.Validate()

	require.Error(t, err)
	var validationErr request.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "status", validationErr.Field)
}

func TestListNotificationsRequest_Validate_WithInvalidAction_ReturnsError(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Action: "NONEXISTENT",
	}

	err := req.Validate()

	require.Error(t, err)
	var validationErr request.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "action", validationErr.Field)
}

func TestListNotificationsRequest_Validate_WithNegativePage_ReturnsError(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Page: -1,
	}

	err := req.Validate()

	require.Error(t, err)
	var validationErr request.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "page", validationErr.Field)
}

func TestListNotificationsRequest_Validate_WithLimitOver100_ReturnsError(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Limit: 101,
	}

	err := req.Validate()

	require.Error(t, err)
	var validationErr request.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "limit", validationErr.Field)
}

func TestListNotificationsRequest_Validate_WithAllValidStatuses_ReturnsNil(t *testing.T) {
	validStatuses := []string{"pending", "sent", "failed", "retrying", "queued"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			req := &request.ListNotificationsRequest{Status: status}
			err := req.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestListNotificationsRequest_ToFilters_WithAllFields_MapsCorrectly(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Type:      "email",
		Action:    "WELCOME",
		Recipient: "user@example.com",
		Status:    "sent",
		Page:      2,
		Limit:     25,
	}

	filters := req.ToFilters()

	require.NotNil(t, filters.Type)
	assert.Equal(t, domain.EmailNotification, *filters.Type)

	require.NotNil(t, filters.Action)
	assert.Equal(t, domain.ActionWelcome, *filters.Action)

	require.NotNil(t, filters.Recipient)
	assert.Equal(t, "user@example.com", *filters.Recipient)

	require.NotNil(t, filters.Status)
	assert.Equal(t, domain.StatusSent, *filters.Status)

	assert.Equal(t, 25, filters.Limit)
	assert.Equal(t, 25, filters.Offset) // (2-1) * 25
}

func TestListNotificationsRequest_ToFilters_WithEmptyFields_ReturnsNilPointers(t *testing.T) {
	req := &request.ListNotificationsRequest{}

	filters := req.ToFilters()

	assert.Nil(t, filters.Type)
	assert.Nil(t, filters.Action)
	assert.Nil(t, filters.Recipient)
	assert.Nil(t, filters.Status)
}

func TestListNotificationsRequest_ToFilters_WithZeroPage_DefaultsToPage1(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Page:  0,
		Limit: 10,
	}

	filters := req.ToFilters()

	assert.Equal(t, 0, filters.Offset) // (1-1) * 10 = 0
	assert.Equal(t, 10, filters.Limit)
}

func TestListNotificationsRequest_ToFilters_WithZeroLimit_DefaultsTo50(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Page:  1,
		Limit: 0,
	}

	filters := req.ToFilters()

	assert.Equal(t, 50, filters.Limit)
	assert.Equal(t, 0, filters.Offset)
}

func TestListNotificationsRequest_ToFilters_WithLimitOver100_DefaultsTo50(t *testing.T) {
	req := &request.ListNotificationsRequest{
		Page:  1,
		Limit: 200,
	}

	filters := req.ToFilters()

	assert.Equal(t, 50, filters.Limit)
}

func TestValidationError_Error_ReturnsFormattedMessage(t *testing.T) {
	err := request.NewValidationError("email", "formato invalido")

	assert.Equal(t, "email: formato invalido", err.Error())
	assert.Equal(t, "email", err.Field)
	assert.Equal(t, "formato invalido", err.Message)
}
