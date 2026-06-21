-- Multi-proyecto (IDP) + multi-tenant: scope namespace(proyecto) → tenant_id → user_id.
-- El `namespace` es el claim del JWT (go-shared lo valida: 'mc' vs otro proyecto).
-- Default 'mc' para las filas preexistentes del proyecto actual.

-- notifications: namespace + tenant_id (la cola/log scopeada por proyecto y tenant)
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS namespace VARCHAR(50) NOT NULL DEFAULT 'mc';
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(36);
CREATE INDEX IF NOT EXISTS idx_notifications_namespace_tenant ON notifications (namespace, tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_ns_status ON notifications (namespace, status);

-- templates: namespace + unique por (namespace, name) para branding/templates por proyecto
ALTER TABLE templates ADD COLUMN IF NOT EXISTS namespace VARCHAR(50) NOT NULL DEFAULT 'mc';
ALTER TABLE templates DROP CONSTRAINT IF EXISTS templates_name_key;
ALTER TABLE templates ADD CONSTRAINT templates_namespace_name_key UNIQUE (namespace, name);
CREATE INDEX IF NOT EXISTS idx_templates_ns_action_type ON templates (namespace, action, type);

-- project_config: configuración por proyecto (identidad de envío, proveedor, canales, cuotas).
-- provider_creds_ref es una REFERENCIA a un secret del cluster, nunca la credencial en claro.
CREATE TABLE IF NOT EXISTS project_config (
    namespace            VARCHAR(50) PRIMARY KEY,
    from_email           VARCHAR(255) NOT NULL,
    from_name            VARCHAR(255) NOT NULL,
    provider_creds_ref   VARCHAR(255),
    channels_enabled     JSONB NOT NULL DEFAULT '["email"]',
    default_template_set VARCHAR(100),
    quota                JSONB,
    created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Seed del proyecto actual (mercado-cercano). provider_creds_ref apunta al secret default de plataforma.
INSERT INTO project_config (namespace, from_email, from_name, provider_creds_ref, channels_enabled)
VALUES ('mc', 'noreply@mercadocercano.com', 'Mercado Cercano', 'RESEND_API_KEY', '["email"]')
ON CONFLICT (namespace) DO NOTHING;
