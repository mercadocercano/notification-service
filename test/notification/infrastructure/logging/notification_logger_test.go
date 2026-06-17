package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"notification-service/src/notification/domain/port"
	notificationlog "notification-service/src/notification/infrastructure/logging"

	"github.com/stretchr/testify/assert"
)

// parseLine verifica que haya exactamente UNA línea JSON por evento (ADR-001).
func parseLine(t *testing.T, b []byte) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(b), []byte("\n"))
	assert.Len(t, lines, 1, "debe ser exactamente una línea por evento")
	var m map[string]any
	assert.NoError(t, json.Unmarshal(lines[0], &m))
	return m
}

// TestNotificationLogger_Sent_EnvelopeAndInfoLevel verifica el envelope canónico en envío exitoso.
func TestNotificationLogger_Sent_EnvelopeAndInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := notificationlog.NewNotificationLoggerWithWriter("notification-test", &buf)

	logger.Log(port.NotificationEvent{
		Event:            "notification.sent",
		TenantID:         "t-123",
		NotificationID:   "n-456",
		NotificationType: "email",
		Action:           "WELCOME",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "notification.sent", line["event"])
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "notification-test", line["service"])
	assert.NotEmpty(t, line["ts"], "ts (RFC3339 UTC) siempre presente")
	assert.Equal(t, "t-123", line["tenant_id"])
	assert.Equal(t, "n-456", line["notification_id"])
	assert.Equal(t, "email", line["notification_type"])
	assert.Equal(t, "WELCOME", line["action"])
	// PRIVACIDAD: nunca debe haber destinatario ni contenido
	_, hasRecipient := line["recipient"]
	assert.False(t, hasRecipient, "el destinatario nunca debe loguearse (privacidad)")
}

// TestNotificationLogger_SendFailed_WarnLevel verifica nivel warn y omisión de campos vacíos.
func TestNotificationLogger_SendFailed_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := notificationlog.NewNotificationLoggerWithWriter("notification-test", &buf)

	logger.Log(port.NotificationEvent{
		Event:            "notification.send_failed",
		TenantID:         "t-123",
		NotificationID:   "n-789",
		NotificationType: "email",
		Action:           "PASSWORD_RESET",
		Reason:           "SMTP connection refused",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "notification.send_failed", line["event"])
	assert.Equal(t, "SMTP connection refused", line["reason"])
	// omitempty: user_id vacío no debe aparecer
	_, hasUserID := line["user_id"]
	assert.False(t, hasUserID, "user_id vacío debe omitirse")
}

// TestNotificationLogger_SaveFailed_ErrorLevel verifica nivel error para fallos de persistencia.
func TestNotificationLogger_SaveFailed_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := notificationlog.NewNotificationLoggerWithWriter("notification-test", &buf)

	logger.Log(port.NotificationEvent{
		Event:            "notification.save_failed",
		TenantID:         "t-999",
		NotificationID:   "n-001",
		NotificationType: "email",
		Action:           "ORDER_CONFIRMATION",
		Reason:           "db: connection refused",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "error", line["level"])
	assert.Equal(t, "notification.save_failed", line["event"])
}

// TestNotificationLogger_Queued_InfoLevel verifica nivel info para envío asíncrono exitoso.
func TestNotificationLogger_Queued_InfoLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := notificationlog.NewNotificationLoggerWithWriter("notification-test", &buf)

	logger.Log(port.NotificationEvent{
		Event:            "notification.queued",
		TenantID:         "t-123",
		NotificationID:   "n-111",
		NotificationType: "email",
		Action:           "SHIPPING_NOTIFICATION",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "notification.queued", line["event"])
	// reason debe omitirse si está vacío
	_, hasReason := line["reason"]
	assert.False(t, hasReason, "reason vacío debe omitirse")
}

// TestNotificationLogger_ValidationFailed_WarnLevel verifica nivel warn para validaciones fallidas.
func TestNotificationLogger_ValidationFailed_WarnLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := notificationlog.NewNotificationLoggerWithWriter("notification-test", &buf)

	logger.Log(port.NotificationEvent{
		Event:            "notification.validation_failed",
		NotificationType: "sms",
		Action:           "WELCOME",
		Reason:           "INVALID_NOTIFICATION_TYPE",
	})

	line := parseLine(t, buf.Bytes())
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "notification.validation_failed", line["event"])
	// tenant_id y notification_id no están seteados — deben omitirse
	_, hasTenant := line["tenant_id"]
	assert.False(t, hasTenant, "tenant_id vacío debe omitirse")
}
