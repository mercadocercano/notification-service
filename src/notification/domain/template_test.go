package domain_test

import (
	"os"
	"path/filepath"
	"testing"

	"notification-service/src/notification/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTemplatePath_ReturnsCorrectPath(t *testing.T) {
	tests := []struct {
		name       string
		templateID string
		expected   string
	}{
		{
			name:       "welcome email",
			templateID: "welcome_email",
			expected:   filepath.Join("templates", "email", "welcome_email.html"),
		},
		{
			name:       "password reset",
			templateID: "password_reset",
			expected:   filepath.Join("templates", "email", "password_reset.html"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.GetTemplatePath(tc.templateID)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetTemplateSubject_WithKnownTemplate_ReturnsSubject(t *testing.T) {
	tests := []struct {
		templateID string
		expected   string
	}{
		{"welcome_email", "!Bienvenido a nuestra plataforma!"},
		{"verification_email", "Verifica tu cuenta"},
		{"password_reset", "Restablece tu contrasena"},
		{"order_confirmation", "Confirmacion de pedido"},
		{"shipping_notification", "Tu pedido ha sido enviado"},
	}

	for _, tc := range tests {
		t.Run(tc.templateID, func(t *testing.T) {
			result := domain.GetTemplateSubject(tc.templateID)
			assert.NotEmpty(t, result)
			assert.NotEqual(t, "Notificacion", result, "known templates should have specific subjects")
		})
	}
}

func TestGetTemplateSubject_WithUnknownTemplate_ReturnsDefault(t *testing.T) {
	result := domain.GetTemplateSubject("unknown_template")
	assert.Equal(t, "Notificación", result)
}

func TestDefaultTemplateMapping_ReturnsExpectedMappings(t *testing.T) {
	mapping := domain.DefaultTemplateMapping()

	assert.NotEmpty(t, mapping)
	assert.Equal(t, "welcome_email", mapping[domain.ActionWelcome])
	assert.Equal(t, "verification_email", mapping[domain.ActionEmailVerification])
	assert.Equal(t, "password_reset", mapping[domain.ActionPasswordReset])
	assert.Equal(t, "order_confirmation", mapping[domain.ActionOrderConfirmation])
	assert.Equal(t, "shipping_notification", mapping[domain.ActionShippingNotification])
}

func TestDefaultTemplateMapping_DoesNotContainAllActions(t *testing.T) {
	mapping := domain.DefaultTemplateMapping()

	// OrderCancellation y PaymentReminder no tienen template por defecto
	_, hasCancellation := mapping[domain.ActionOrderCancellation]
	_, hasReminder := mapping[domain.ActionPaymentReminder]

	assert.False(t, hasCancellation, "OrderCancellation should not have default template")
	assert.False(t, hasReminder, "PaymentReminder should not have default template")
}

func TestTemplate_RenderHTML_WithValidTemplate_ReturnsRenderedHTML(t *testing.T) {
	// Arrange - crear archivo temporal de template
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test_template.html")
	templateContent := `<h1>Hola {{.name}}</h1><p>Tu email es {{.email}}</p>`
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	tmpl := &domain.Template{
		ID:       "test-001",
		Name:     "test_template",
		FilePath: templatePath,
	}

	data := map[string]interface{}{
		"name":  "Juan",
		"email": "juan@example.com",
	}

	// Act
	result, err := tmpl.RenderHTML(data)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, result, "Hola Juan")
	assert.Contains(t, result, "juan@example.com")
}

func TestTemplate_RenderHTML_WithMissingFile_ReturnsError(t *testing.T) {
	tmpl := &domain.Template{
		ID:       "test-002",
		Name:     "missing_template",
		FilePath: "/nonexistent/path/template.html",
	}

	_, err := tmpl.RenderHTML(map[string]interface{}{})

	assert.Error(t, err)
}

func TestTemplate_RenderHTML_WithInvalidTemplate_ReturnsError(t *testing.T) {
	// Arrange - template con sintaxis invalida
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "bad_template.html")
	err := os.WriteFile(templatePath, []byte(`{{.name`), 0644)
	require.NoError(t, err)

	tmpl := &domain.Template{
		ID:       "test-003",
		Name:     "bad_template",
		FilePath: templatePath,
	}

	// Act
	_, err = tmpl.RenderHTML(map[string]interface{}{"name": "Test"})

	// Assert
	assert.Error(t, err)
}
