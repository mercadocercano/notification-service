package repository

import (
	"context"
	"database/sql"
	"fmt"
	"notification-service/src/notification/domain"
	"notification-service/src/shared/logger"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type postgresTemplateRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresTemplateRepository(db *sql.DB) domain.TemplateRepository {
	return &postgresTemplateRepository{
		db:     db,
		logger: logger.GetLogger(),
	}
}

func (r *postgresTemplateRepository) FindByID(ctx context.Context, id string) (*domain.Template, error) {
	query := `
		SELECT id, name, subject, file_path, action, type, version, is_active, created_at, updated_at
		FROM templates 
		WHERE id = $1 AND is_active = true
	`

	var template domain.Template
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Subject,
		&template.FilePath,
		&template.Action,
		&template.Type,
		&template.Version,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Template not found", zap.String("id", id))
			return nil, fmt.Errorf("template not found: %s", id)
		}
		r.logger.Error("Error finding template by ID", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &template, nil
}

func (r *postgresTemplateRepository) FindByName(ctx context.Context, name string) (*domain.Template, error) {
	query := `
		SELECT id, name, subject, file_path, action, type, version, is_active, created_at, updated_at
		FROM templates 
		WHERE name = $1 AND is_active = true
	`

	var template domain.Template
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Subject,
		&template.FilePath,
		&template.Action,
		&template.Type,
		&template.Version,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Template not found", zap.String("name", name))
			return nil, fmt.Errorf("template not found: %s", name)
		}
		r.logger.Error("Error finding template by name", zap.String("name", name), zap.Error(err))
		return nil, err
	}

	return &template, nil
}

func (r *postgresTemplateRepository) FindByAction(ctx context.Context, action domain.NotificationAction, notificationType domain.NotificationType) (*domain.Template, error) {
	r.logger.Debug("Finding template by action",
		zap.String("action", string(action)),
		zap.String("type", string(notificationType)))

	query := `
		SELECT id, name, subject, file_path, action, type, version, is_active, created_at, updated_at
		FROM templates 
		WHERE action = $1 AND type = $2 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	var template domain.Template
	err := r.db.QueryRowContext(ctx, query, action, notificationType).Scan(
		&template.ID,
		&template.Name,
		&template.Subject,
		&template.FilePath,
		&template.Action,
		&template.Type,
		&template.Version,
		&template.IsActive,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Template not found for action",
				zap.String("action", string(action)),
				zap.String("type", string(notificationType)))
			return nil, fmt.Errorf("template not found for action: %s and type: %s", action, notificationType)
		}
		r.logger.Error("Error finding template by action",
			zap.String("action", string(action)),
			zap.String("type", string(notificationType)),
			zap.Error(err))
		return nil, err
	}

	r.logger.Debug("Template found by action",
		zap.String("template_id", template.ID),
		zap.String("template_name", template.Name))

	return &template, nil
}

func (r *postgresTemplateRepository) Save(ctx context.Context, template *domain.Template) error {
	query := `
		INSERT INTO templates (id, name, subject, file_path, action, type, version, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Subject,
		template.FilePath,
		template.Action,
		template.Type,
		template.Version,
		template.IsActive,
		template.CreatedAt,
		template.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Error saving template", zap.String("id", template.ID), zap.Error(err))
		return err
	}

	r.logger.Info("Template saved successfully", zap.String("id", template.ID))
	return nil
}

func (r *postgresTemplateRepository) Update(ctx context.Context, template *domain.Template) error {
	query := `
		UPDATE templates 
		SET name = $2, subject = $3, file_path = $4, action = $5, type = $6, version = $7, is_active = $8, updated_at = $9
		WHERE id = $1
	`

	template.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		template.ID,
		template.Name,
		template.Subject,
		template.FilePath,
		template.Action,
		template.Type,
		template.Version,
		template.IsActive,
		template.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Error updating template", zap.String("id", template.ID), zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	r.logger.Info("Template updated successfully", zap.String("id", template.ID))
	return nil
}
