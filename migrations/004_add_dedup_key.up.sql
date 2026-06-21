-- Dedup de notificaciones (entrega at-least-once del EventBus + Idempotency-Key del API sync).
-- La clave compuesta scopea el dedup por proyecto (namespace) + tenant.
-- COALESCE(tenant_id, '') porque tenant_id es NULLABLE y en un índice UNIQUE los NULL son
-- distintos entre sí (dos filas con tenant_id NULL NO colisionarían) → normalizamos a ''.
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS dedup_key VARCHAR(255);

CREATE UNIQUE INDEX IF NOT EXISTS uq_notifications_dedup
    ON notifications (namespace, COALESCE(tenant_id, ''), dedup_key)
    WHERE dedup_key IS NOT NULL;
