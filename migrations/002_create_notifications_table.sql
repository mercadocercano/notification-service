-- +migrate Up
-- Crear tabla notifications
CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    template_id VARCHAR(36),
    recipient VARCHAR(500) NOT NULL,
    data JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Crear índices para optimizar consultas
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications (type);
CREATE INDEX IF NOT EXISTS idx_notifications_action ON notifications (action);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications (status);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications (recipient);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_status_created ON notifications (status, created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_pending ON notifications (status) WHERE status = 'pending';

-- +migrate Down
-- Eliminar índices
DROP INDEX IF EXISTS idx_notifications_pending;
DROP INDEX IF EXISTS idx_notifications_status_created;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_recipient;
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_action;
DROP INDEX IF EXISTS idx_notifications_type;

-- Eliminar tabla notifications
DROP TABLE IF EXISTS notifications; 