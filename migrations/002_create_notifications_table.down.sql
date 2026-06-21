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
