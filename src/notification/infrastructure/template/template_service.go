package template

import (
	"context"
	"fmt"
	"notification-service/src/notification/domain"
	"notification-service/src/shared/config"
	"notification-service/src/shared/logger"
	"time"

	"go.uber.org/zap"
)

type templateService struct {
	logger           *zap.Logger
	templateRepo     domain.TemplateRepository
	fallbackToStatic bool // Fallback a mapeo estático si no hay DB
	config           *config.Config
}

func NewTemplateService(templateRepo domain.TemplateRepository) domain.TemplateService {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Si no puede cargar config, usar valores por defecto
		cfg = &config.Config{
			Contact: config.ContactConfig{
				Email: "contacto@mercadocercano.com",
			},
		}
	}

	return &templateService{
		logger:           logger.GetLogger(),
		templateRepo:     templateRepo,
		config:           cfg,
		fallbackToStatic: templateRepo == nil,
	}
}

func (s *templateService) RenderTemplateByAction(action domain.NotificationAction, notificationType domain.NotificationType, data map[string]interface{}) (subject string, html string, err error) {
	s.logger.Debug("Rendering template by action",
		zap.String("action", string(action)),
		zap.String("type", string(notificationType)),
		zap.Any("data_keys", getMapKeys(data)))

	// Agregar variables automáticas a los datos
	if data == nil {
		data = make(map[string]interface{})
	}
	data["current_year"] = time.Now().Year()
	data["contact_email"] = s.config.Contact.Email

	// Crear contexto para la operación de base de datos
	ctx := context.Background()

	// Buscar template por acción
	template, err := s.GetTemplateByAction(ctx, action, notificationType)
	if err != nil {
		s.logger.Error("Failed to get template by action",
			zap.String("action", string(action)),
			zap.String("type", string(notificationType)),
			zap.Error(err))
		return "", "", err
	}

	// Renderizar el HTML
	html, err = template.RenderHTML(data)
	if err != nil {
		s.logger.Error("Failed to render template",
			zap.String("template_id", template.ID),
			zap.String("action", string(action)),
			zap.Error(err))
		return "", "", fmt.Errorf("error rendering template for action %s: %w", action, err)
	}

	subject = template.Subject

	s.logger.Debug("Template rendered successfully by action",
		zap.String("action", string(action)),
		zap.String("template_id", template.ID),
		zap.String("subject", subject))

	return subject, html, nil
}

func (s *templateService) RenderTemplate(templateID string, data map[string]interface{}) (subject string, html string, err error) {
	s.logger.Debug("Rendering template",
		zap.String("template_id", templateID),
		zap.Any("data_keys", getMapKeys(data)))

	// Agregar variables automáticas a los datos
	if data == nil {
		data = make(map[string]interface{})
	}
	data["current_year"] = time.Now().Year()
	data["contact_email"] = s.config.Contact.Email

	// Crear el template con el path (método legacy)
	template := &domain.Template{
		ID:       templateID,
		Name:     templateID,
		Subject:  domain.GetTemplateSubject(templateID),
		FilePath: domain.GetTemplatePath(templateID),
	}

	// Renderizar el HTML
	html, err = template.RenderHTML(data)
	if err != nil {
		s.logger.Error("Failed to render template",
			zap.String("template_id", templateID),
			zap.Error(err))
		return "", "", fmt.Errorf("error rendering template %s: %w", templateID, err)
	}

	subject = template.Subject

	s.logger.Debug("Template rendered successfully",
		zap.String("template_id", templateID),
		zap.String("subject", subject))

	return subject, html, nil
}

func (s *templateService) GetTemplate(templateID string) (*domain.Template, error) {
	s.logger.Debug("Getting template", zap.String("template_id", templateID))

	template := &domain.Template{
		ID:       templateID,
		Name:     templateID,
		Subject:  domain.GetTemplateSubject(templateID),
		FilePath: domain.GetTemplatePath(templateID),
		Type:     domain.EmailNotification,
		IsActive: true,
	}

	return template, nil
}

func (s *templateService) GetTemplateByAction(ctx context.Context, action domain.NotificationAction, notificationType domain.NotificationType) (*domain.Template, error) {
	s.logger.Debug("Getting template by action",
		zap.String("action", string(action)),
		zap.String("type", string(notificationType)))

	// Si tenemos repository, usarlo
	if !s.fallbackToStatic && s.templateRepo != nil {
		template, err := s.templateRepo.FindByAction(ctx, action, notificationType)
		if err != nil {
			s.logger.Warn("Template not found in database, falling back to static mapping",
				zap.String("action", string(action)),
				zap.Error(err))
			// Fallback a mapeo estático
			return s.getTemplateByActionStatic(action, notificationType)
		}
		return template, nil
	}

	// Fallback a mapeo estático
	return s.getTemplateByActionStatic(action, notificationType)
}

// getTemplateByActionStatic mapeo estático como fallback
func (s *templateService) getTemplateByActionStatic(action domain.NotificationAction, notificationType domain.NotificationType) (*domain.Template, error) {
	mapping := domain.DefaultTemplateMapping()
	templateID, exists := mapping[action]
	if !exists {
		return nil, fmt.Errorf("no template found for action: %s", action)
	}

	template := &domain.Template{
		ID:       templateID,
		Name:     templateID,
		Subject:  domain.GetTemplateSubject(templateID),
		FilePath: domain.GetTemplatePath(templateID),
		Action:   action,
		Type:     notificationType,
		IsActive: true,
	}

	s.logger.Debug("Using static template mapping",
		zap.String("action", string(action)),
		zap.String("template_id", templateID))

	return template, nil
}

// getMapKeys extrae las claves de un map para logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
