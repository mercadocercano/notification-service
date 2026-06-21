-- Eliminar índices
DROP INDEX IF EXISTS idx_templates_created_at;
DROP INDEX IF EXISTS idx_templates_active;
DROP INDEX IF EXISTS idx_templates_action_type_version;
DROP INDEX IF EXISTS idx_templates_action_type;

-- Eliminar tabla templates
DROP TABLE IF EXISTS templates;
