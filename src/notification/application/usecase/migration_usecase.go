package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"notification-service/src/shared/logger"

	"go.uber.org/zap"
)

type MigrationUseCase struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewMigrationUseCase(db *sql.DB) *MigrationUseCase {
	return &MigrationUseCase{
		db:     db,
		logger: logger.GetLogger(),
	}
}

type MigrationResult struct {
	Success            bool     `json:"success"`
	Message            string   `json:"message"`
	ExecutedMigrations []string `json:"executed_migrations,omitempty"`
	Error              string   `json:"error,omitempty"`
}

func (uc *MigrationUseCase) RunMigrations(ctx context.Context) *MigrationResult {
	uc.logger.Info("Starting database migrations")

	// 1. Crear tabla de migraciones si no existe
	if err := uc.createMigrationTable(ctx); err != nil {
		uc.logger.Error("Failed to create migration table", zap.Error(err))
		return &MigrationResult{
			Success: false,
			Error:   fmt.Sprintf("Error creating migration table: %v", err),
		}
	}

	// 2. Detectar migraciones ya aplicadas manualmente y marcarlas
	if err := uc.detectAndMarkExistingTables(ctx); err != nil {
		uc.logger.Warn("Failed to detect existing tables", zap.Error(err))
		// No fallar completamente, solo continuar
	}

	// 3. Obtener migraciones ejecutadas
	executedMigrations, err := uc.getExecutedMigrations(ctx)
	if err != nil {
		uc.logger.Error("Failed to get executed migrations", zap.Error(err))
		return &MigrationResult{
			Success: false,
			Error:   fmt.Sprintf("Error getting executed migrations: %v", err),
		}
	}

	// 4. Obtener archivos de migración
	migrationFiles, err := uc.getMigrationFiles()
	if err != nil {
		uc.logger.Error("Failed to get migration files", zap.Error(err))
		return &MigrationResult{
			Success: false,
			Error:   fmt.Sprintf("Error getting migration files: %v", err),
		}
	}

	// 5. Ejecutar migraciones pendientes
	var newlyExecuted []string
	for _, file := range migrationFiles {
		if !contains(executedMigrations, file) {
			uc.logger.Info("Executing migration", zap.String("file", file))

			if err := uc.executeMigration(ctx, file); err != nil {
				// Si la tabla ya existe, marcarla como ejecutada y continuar
				if uc.isTableAlreadyExistsError(err) {
					uc.logger.Info("Table already exists, marking migration as executed",
						zap.String("file", file))

					if markErr := uc.markMigrationAsExecuted(ctx, file); markErr != nil {
						uc.logger.Error("Failed to mark existing migration as executed",
							zap.String("file", file),
							zap.Error(markErr))
					} else {
						newlyExecuted = append(newlyExecuted, file+" (already existed)")
					}
					continue
				}

				uc.logger.Error("Failed to execute migration",
					zap.String("file", file),
					zap.Error(err))
				return &MigrationResult{
					Success:            false,
					Error:              fmt.Sprintf("Error executing migration %s: %v", file, err),
					ExecutedMigrations: newlyExecuted,
				}
			}

			// Marcar como ejecutada
			if err := uc.markMigrationAsExecuted(ctx, file); err != nil {
				uc.logger.Error("Failed to mark migration as executed",
					zap.String("file", file),
					zap.Error(err))
				return &MigrationResult{
					Success:            false,
					Error:              fmt.Sprintf("Error marking migration as executed %s: %v", file, err),
					ExecutedMigrations: newlyExecuted,
				}
			}

			newlyExecuted = append(newlyExecuted, file)
		}
	}

	message := "Database migrations completed successfully"
	if len(newlyExecuted) == 0 {
		message = "No pending migrations found - database is up to date"
	}

	uc.logger.Info("Migration process completed",
		zap.Int("newly_executed", len(newlyExecuted)),
		zap.Strings("executed_files", newlyExecuted))

	return &MigrationResult{
		Success:            true,
		Message:            message,
		ExecutedMigrations: newlyExecuted,
	}
}

func (uc *MigrationUseCase) createMigrationTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := uc.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	uc.logger.Debug("Migration table created or already exists")
	return nil
}

func (uc *MigrationUseCase) getExecutedMigrations(ctx context.Context) ([]string, error) {
	query := `SELECT migration_name FROM schema_migrations ORDER BY executed_at`

	rows, err := uc.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query executed migrations: %w", err)
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan migration name: %w", err)
		}
		migrations = append(migrations, name)
	}

	return migrations, nil
}

func (uc *MigrationUseCase) getMigrationFiles() ([]string, error) {
	migrationDir := "migrations"

	var files []string
	err := filepath.WalkDir(migrationDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".sql") {
			files = append(files, d.Name())
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Ordenar archivos para ejecutar en orden
	sort.Strings(files)

	uc.logger.Debug("Found migration files", zap.Strings("files", files))
	return files, nil
}

func (uc *MigrationUseCase) executeMigration(ctx context.Context, filename string) error {
	migrationPath := filepath.Join("migrations", filename)

	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filename, err)
	}

	// Extraer solo la parte "Up" de la migración
	migrationSQL := uc.extractUpMigration(string(content))

	if migrationSQL == "" {
		return fmt.Errorf("no up migration found in file %s", filename)
	}

	// Ejecutar en transacción
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Ejecutar SQL
	if _, err := tx.ExecContext(ctx, migrationSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	uc.logger.Info("Migration executed successfully", zap.String("file", filename))
	return nil
}

func (uc *MigrationUseCase) extractUpMigration(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUpSection := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.Contains(trimmedLine, "+migrate Up") || strings.Contains(trimmedLine, "-- +migrate Up") {
			inUpSection = true
			continue
		}

		if strings.Contains(trimmedLine, "+migrate Down") || strings.Contains(trimmedLine, "-- +migrate Down") {
			break
		}

		if inUpSection && !strings.HasPrefix(trimmedLine, "--") {
			upLines = append(upLines, line)
		}
	}

	return strings.Join(upLines, "\n")
}

func (uc *MigrationUseCase) markMigrationAsExecuted(ctx context.Context, filename string) error {
	query := `INSERT INTO schema_migrations (migration_name) VALUES ($1)`

	_, err := uc.db.ExecContext(ctx, query, filename)
	if err != nil {
		return fmt.Errorf("failed to mark migration as executed: %w", err)
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (uc *MigrationUseCase) detectAndMarkExistingTables(ctx context.Context) error {
	// Verificar si las tablas principales ya existen
	existingTables := []struct {
		tableName     string
		migrationFile string
	}{
		{"templates", "001_create_templates_table.sql"},
		{"notifications", "002_create_notifications_table.sql"},
	}

	for _, table := range existingTables {
		exists, err := uc.tableExists(ctx, table.tableName)
		if err != nil {
			uc.logger.Warn("Failed to check if table exists",
				zap.String("table", table.tableName),
				zap.Error(err))
			continue
		}

		if exists {
			// Verificar si la migración ya está marcada como ejecutada
			executed, err := uc.isMigrationExecuted(ctx, table.migrationFile)
			if err != nil {
				uc.logger.Warn("Failed to check migration status",
					zap.String("migration", table.migrationFile),
					zap.Error(err))
				continue
			}

			if !executed {
				uc.logger.Info("Found existing table, marking migration as executed",
					zap.String("table", table.tableName),
					zap.String("migration", table.migrationFile))

				if err := uc.markMigrationAsExecuted(ctx, table.migrationFile); err != nil {
					uc.logger.Error("Failed to mark existing table migration as executed",
						zap.String("migration", table.migrationFile),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

func (uc *MigrationUseCase) tableExists(ctx context.Context, tableName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)
	`

	var exists bool
	err := uc.db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	return exists, err
}

func (uc *MigrationUseCase) isMigrationExecuted(ctx context.Context, migrationName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE migration_name = $1)`

	var exists bool
	err := uc.db.QueryRowContext(ctx, query, migrationName).Scan(&exists)
	return exists, err
}

func (uc *MigrationUseCase) isTableAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "relation") && strings.Contains(errStr, "already exists")
}
