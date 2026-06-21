DROP TABLE IF EXISTS project_config;

ALTER TABLE templates DROP CONSTRAINT IF EXISTS templates_namespace_name_key;
DROP INDEX IF EXISTS idx_templates_ns_action_type;
ALTER TABLE templates ADD CONSTRAINT templates_name_key UNIQUE (name);
ALTER TABLE templates DROP COLUMN IF EXISTS namespace;

DROP INDEX IF EXISTS idx_notifications_ns_status;
DROP INDEX IF EXISTS idx_notifications_namespace_tenant;
ALTER TABLE notifications DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE notifications DROP COLUMN IF EXISTS namespace;
