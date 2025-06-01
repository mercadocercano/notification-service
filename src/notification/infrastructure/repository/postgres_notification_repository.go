package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"notification-service/src/notification/domain"
	"notification-service/src/shared/logger"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type postgresNotificationRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresNotificationRepository(db *sql.DB) domain.NotificationRepository {
	return &postgresNotificationRepository{
		db:     db,
		logger: logger.GetLogger(),
	}
}

func (r *postgresNotificationRepository) Save(ctx context.Context, notification *domain.Notification) error {
	r.logger.Debug("Saving notification to PostgreSQL",
		zap.String("id", notification.ID),
		zap.String("type", string(notification.Type)),
		zap.String("action", string(notification.Action)))

	query := `
		INSERT INTO notifications (id, type, action, template_id, recipient, data, status, retry_count, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Convertir data a JSON
	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		r.logger.Error("Error marshaling notification data", zap.Error(err))
		return fmt.Errorf("error marshaling data: %w", err)
	}

	now := time.Now()
	notification.CreatedAt = now
	notification.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		notification.ID,
		notification.Type,
		notification.Action,
		notification.TemplateID,
		notification.Recipient,
		dataJSON,
		notification.Status,
		notification.RetryCount,
		notification.Error,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Error saving notification",
			zap.String("id", notification.ID),
			zap.Error(err))
		return fmt.Errorf("error saving notification: %w", err)
	}

	r.logger.Info("Notification saved successfully", zap.String("id", notification.ID))
	return nil
}

func (r *postgresNotificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	r.logger.Debug("Finding notification by ID", zap.String("id", id))

	query := `
		SELECT id, type, action, template_id, recipient, data, status, retry_count, error_message, created_at, updated_at
		FROM notifications 
		WHERE id = $1
	`

	var notification domain.Notification
	var dataJSON []byte
	var templateID sql.NullString
	var errorMessage sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&notification.ID,
		&notification.Type,
		&notification.Action,
		&templateID,
		&notification.Recipient,
		&dataJSON,
		&notification.Status,
		&notification.RetryCount,
		&errorMessage,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Notification not found", zap.String("id", id))
			return nil, fmt.Errorf("notification not found: %s", id)
		}
		r.logger.Error("Error finding notification by ID", zap.String("id", id), zap.Error(err))
		return nil, fmt.Errorf("error finding notification: %w", err)
	}

	// Convertir JSON a map
	if len(dataJSON) > 0 {
		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			r.logger.Error("Error unmarshaling notification data", zap.Error(err))
			return nil, fmt.Errorf("error unmarshaling data: %w", err)
		}
	}

	// Asignar valores nullable
	if templateID.Valid {
		notification.TemplateID = templateID.String
	}
	if errorMessage.Valid {
		notification.Error = errorMessage.String
	}

	r.logger.Debug("Notification found", zap.String("id", id))
	return &notification, nil
}

func (r *postgresNotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	r.logger.Debug("Updating notification", zap.String("id", notification.ID))

	query := `
		UPDATE notifications 
		SET type = $2, action = $3, template_id = $4, recipient = $5, data = $6, 
		    status = $7, retry_count = $8, error_message = $9, updated_at = $10
		WHERE id = $1
	`

	// Convertir data a JSON
	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		r.logger.Error("Error marshaling notification data", zap.Error(err))
		return fmt.Errorf("error marshaling data: %w", err)
	}

	notification.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.Type,
		notification.Action,
		notification.TemplateID,
		notification.Recipient,
		dataJSON,
		notification.Status,
		notification.RetryCount,
		notification.Error,
		notification.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Error updating notification", zap.String("id", notification.ID), zap.Error(err))
		return fmt.Errorf("error updating notification: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", notification.ID)
	}

	r.logger.Info("Notification updated successfully", zap.String("id", notification.ID))
	return nil
}

func (r *postgresNotificationRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, errorMessage string) error {
	r.logger.Debug("Updating notification status",
		zap.String("id", id),
		zap.String("status", string(status)))

	query := `
		UPDATE notifications 
		SET status = $2, error_message = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status, errorMessage, time.Now())
	if err != nil {
		r.logger.Error("Error updating notification status",
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("error updating notification status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", id)
	}

	r.logger.Info("Notification status updated successfully",
		zap.String("id", id),
		zap.String("status", string(status)))
	return nil
}

func (r *postgresNotificationRepository) FindPendingNotifications(ctx context.Context) ([]*domain.Notification, error) {
	r.logger.Debug("Finding pending notifications")

	query := `
		SELECT id, type, action, template_id, recipient, data, status, retry_count, error_message, created_at, updated_at
		FROM notifications 
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Error finding pending notifications", zap.Error(err))
		return nil, fmt.Errorf("error finding pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var notification domain.Notification
		var dataJSON []byte
		var templateID sql.NullString
		var errorMessage sql.NullString

		err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Action,
			&templateID,
			&notification.Recipient,
			&dataJSON,
			&notification.Status,
			&notification.RetryCount,
			&errorMessage,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Error scanning notification row", zap.Error(err))
			return nil, fmt.Errorf("error scanning notification: %w", err)
		}

		// Convertir JSON a map
		if len(dataJSON) > 0 {
			if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
				r.logger.Error("Error unmarshaling notification data", zap.Error(err))
				continue // Skip this row
			}
		}

		// Asignar valores nullable
		if templateID.Valid {
			notification.TemplateID = templateID.String
		}
		if errorMessage.Valid {
			notification.Error = errorMessage.String
		}

		notifications = append(notifications, &notification)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error iterating notification rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	r.logger.Info("Found pending notifications", zap.Int("count", len(notifications)))
	return notifications, nil
}
