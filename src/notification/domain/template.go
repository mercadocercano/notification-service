package domain

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

type Template struct {
	ID        string
	Name      string
	Subject   string
	FilePath  string             // Path al archivo HTML del template
	Action    NotificationAction // Acción asociada al template
	Type      NotificationType
	Variables []string
	Version   int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TemplateRepository interface {
	FindByID(ctx context.Context, id string) (*Template, error)
	FindByName(ctx context.Context, name string) (*Template, error)
	FindByAction(ctx context.Context, action NotificationAction, notificationType NotificationType) (*Template, error)
	Save(ctx context.Context, template *Template) error
	Update(ctx context.Context, template *Template) error
}

type TemplateService interface {
	RenderTemplateByAction(action NotificationAction, notificationType NotificationType, data map[string]interface{}) (subject string, html string, err error)
	RenderTemplate(templateID string, data map[string]interface{}) (subject string, html string, err error)
	GetTemplate(templateID string) (*Template, error)
	GetTemplateByAction(ctx context.Context, action NotificationAction, notificationType NotificationType) (*Template, error)
}

// RenderHTML renderiza el template HTML con los datos proporcionados
func (t *Template) RenderHTML(data map[string]interface{}) (string, error) {
	// Leer el archivo del template
	content, err := os.ReadFile(t.FilePath)
	if err != nil {
		return "", err
	}

	// Crear el template de Go
	tmpl, err := template.New(t.Name).Parse(string(content))
	if err != nil {
		return "", err
	}

	// Renderizar con los datos
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetTemplatePath devuelve el path completo del template basado en el ID
func GetTemplatePath(templateID string) string {
	return filepath.Join("templates", "email", templateID+".html")
}

// GetTemplateSubject devuelve el subject basado en el template ID
func GetTemplateSubject(templateID string) string {
	subjects := map[string]string{
		"welcome_email":         "¡Bienvenido a nuestra plataforma!",
		"verification_email":    "Verifica tu cuenta",
		"password_reset":        "Restablece tu contraseña",
		"order_confirmation":    "Confirmación de pedido",
		"shipping_notification": "Tu pedido ha sido enviado",
	}

	if subject, exists := subjects[templateID]; exists {
		return subject
	}
	return "Notificación"
}

// DefaultTemplateMapping mapea acciones a template IDs por defecto
// Esto será reemplazado por la base de datos posteriormente
func DefaultTemplateMapping() map[NotificationAction]string {
	return map[NotificationAction]string{
		ActionWelcome:              "welcome_email",
		ActionEmailVerification:    "verification_email",
		ActionPasswordReset:        "password_reset",
		ActionOrderConfirmation:    "order_confirmation",
		ActionShippingNotification: "shipping_notification",
	}
}
