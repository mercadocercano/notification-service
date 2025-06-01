package email

import (
	"context"
	"notification-service/pkg/validator"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/ports/output"
	"notification-service/src/shared/logger"

	"github.com/resendlabs/resend-go"
	"go.uber.org/zap"
)

type resendClient struct {
	client          *resend.Client
	logger          *zap.Logger
	fromEmail       string
	templateService domain.TemplateService
}

func NewResendClient(apiKey string, fromEmail string, templateService domain.TemplateService) output.EmailSender {
	return &resendClient{
		client:          resend.NewClient(apiKey),
		logger:          logger.GetLogger(),
		fromEmail:       fromEmail,
		templateService: templateService,
	}
}

func (client *resendClient) SendEmail(ctx context.Context, to string, templateID string, data map[string]interface{}) error {
	client.logger.Info("Sending email",
		zap.String("to", to),
		zap.String("template_id", templateID))

	// Renderizar el template
	subject, html, err := client.templateService.RenderTemplate(templateID, data)
	if err != nil {
		client.logger.Error("Failed to render template",
			zap.String("template_id", templateID),
			zap.Error(err))
		return err
	}

	// Enviar el email
	return client.sendEmail(ctx, to, subject, html)
}

func (client *resendClient) SendEmailByAction(ctx context.Context, to string, action domain.NotificationAction, notificationType domain.NotificationType, data map[string]interface{}) error {
	client.logger.Info("Sending email by action",
		zap.String("to", to),
		zap.String("action", string(action)),
		zap.String("type", string(notificationType)))

	// Renderizar el template por acción
	subject, html, err := client.templateService.RenderTemplateByAction(action, notificationType, data)
	if err != nil {
		client.logger.Error("Failed to render template by action",
			zap.String("action", string(action)),
			zap.String("type", string(notificationType)),
			zap.Error(err))
		return err
	}

	// Enviar el email
	return client.sendEmail(ctx, to, subject, html)
}

// sendEmail método privado que realiza el envío real
func (client *resendClient) sendEmail(ctx context.Context, to, subject, html string) error {
	// Preparar el email
	params := &resend.SendEmailRequest{
		From:    client.fromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	// Enviar el email
	_, err := client.client.Emails.Send(params)
	if err != nil {
		client.logger.Error("Failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return err
	}

	client.logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject))

	return nil
}

func (client *resendClient) ValidateEmail(email string) bool {
	return validator.IsValidEmail(email)
}
