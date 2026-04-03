package response_test

import (
	"net/http"
	"testing"
	"time"

	"notification-service/src/notification/application/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSendNotificationSuccess_ReturnsSuccessResult(t *testing.T) {
	data := &response.SendNotificationResponse{
		ID:        "notif-001",
		Message:   "Sent",
		Status:    "sent",
		CreatedAt: time.Now(),
	}

	result := response.NewSendNotificationSuccess(data)

	assert.True(t, result.Success)
	assert.Nil(t, result.Error)
	assert.Equal(t, http.StatusOK, result.HTTPStatus)
	require.NotNil(t, result.Data)
	assert.Equal(t, "notif-001", result.Data.ID)
}

func TestNewSendNotificationError_ReturnsErrorResult(t *testing.T) {
	result := response.NewSendNotificationError(
		"TEST_ERROR",
		"Error de prueba",
		"detalle extra",
		http.StatusBadRequest,
	)

	assert.False(t, result.Success)
	assert.Nil(t, result.Data)
	assert.Equal(t, http.StatusBadRequest, result.HTTPStatus)
	require.NotNil(t, result.Error)
	assert.Equal(t, "TEST_ERROR", result.Error.Code)
	assert.Equal(t, "Error de prueba", result.Error.Message)
	assert.Equal(t, "detalle extra", result.Error.Details)
}

func TestNewInvalidEmailError_ReturnsCorrectError(t *testing.T) {
	result := response.NewInvalidEmailError()

	assert.False(t, result.Success)
	assert.Equal(t, http.StatusBadRequest, result.HTTPStatus)
	assert.Equal(t, "INVALID_EMAIL", result.Error.Code)
}

func TestNewTemplateNotFoundError_ReturnsCorrectError(t *testing.T) {
	result := response.NewTemplateNotFoundError()

	assert.False(t, result.Success)
	assert.Equal(t, http.StatusBadRequest, result.HTTPStatus)
	assert.Equal(t, "TEMPLATE_NOT_FOUND", result.Error.Code)
}

func TestNewInvalidNotificationTypeError_ReturnsCorrectError(t *testing.T) {
	result := response.NewInvalidNotificationTypeError()

	assert.False(t, result.Success)
	assert.Equal(t, http.StatusBadRequest, result.HTTPStatus)
	assert.Equal(t, "INVALID_NOTIFICATION_TYPE", result.Error.Code)
}

func TestNewInternalServerError_ReturnsCorrectError(t *testing.T) {
	result := response.NewInternalServerError("db connection failed")

	assert.False(t, result.Success)
	assert.Equal(t, http.StatusInternalServerError, result.HTTPStatus)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Error.Code)
	assert.Equal(t, "db connection failed", result.Error.Details)
}

func TestSendNotificationResult_ToMiddlewareError_WithError_ReturnsMappedError(t *testing.T) {
	result := response.NewSendNotificationError(
		"CUSTOM_CODE",
		"mensaje custom",
		"detalle",
		http.StatusConflict,
	)

	middlewareErr := result.ToMiddlewareError()

	assert.Equal(t, "CUSTOM_CODE", middlewareErr.Code)
	assert.Equal(t, "mensaje custom", middlewareErr.Message)
	assert.Equal(t, "detalle", middlewareErr.Details)
	assert.Equal(t, http.StatusConflict, middlewareErr.HTTPStatus)
}

func TestSendNotificationResult_ToMiddlewareError_WithSuccess_ReturnsEmptyError(t *testing.T) {
	data := &response.SendNotificationResponse{
		ID:     "notif-001",
		Status: "sent",
	}
	result := response.NewSendNotificationSuccess(data)

	middlewareErr := result.ToMiddlewareError()

	assert.Empty(t, middlewareErr.Code)
	assert.Empty(t, middlewareErr.Message)
}
