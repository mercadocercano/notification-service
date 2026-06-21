-- Crear tabla templates
CREATE TABLE IF NOT EXISTS templates (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    subject VARCHAR(500) NOT NULL,
    file_path VARCHAR(1000) NOT NULL,
    action VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Crear índices para optimizar consultas
CREATE INDEX IF NOT EXISTS idx_templates_action_type ON templates (action, type);
CREATE INDEX IF NOT EXISTS idx_templates_action_type_version ON templates (action, type, version);
CREATE INDEX IF NOT EXISTS idx_templates_active ON templates (is_active);
CREATE INDEX IF NOT EXISTS idx_templates_created_at ON templates (created_at);

-- Insertar templates de ejemplo para las acciones principales
INSERT INTO templates (id, name, subject, file_path, action, type, version, is_active) VALUES
('template-welcome-001', 'welcome_email', 'Bienvenido a nuestra plataforma', './templates/email/welcome_email.html', 'WELCOME', 'email', 1, true),
('template-verification-001', 'verification_email', 'Verificación de email', './templates/email/verification_email.html', 'EMAIL_VERIFICATION', 'email', 1, true),
('template-password-reset-001', 'password_reset_email', 'Recuperación de contraseña', './templates/email/password_reset.html', 'PASSWORD_RESET', 'email', 1, true),
('template-order-confirmation-001', 'order_confirmation_email', 'Confirmación de pedido', './templates/email/order_confirmation.html', 'ORDER_CONFIRMATION', 'email', 1, true),
('template-shipping-notification-001', 'shipping_notification_email', 'Tu pedido ha sido enviado', './templates/email/shipping_notification.html', 'SHIPPING_NOTIFICATION', 'email', 1, true)
ON CONFLICT (id) DO NOTHING;
