# CLAUDE.md — notification-service

Notificaciones **email/SMS** con plantillas; correo vía **Resend**. Envío síncrono o asíncrono (SQS si está activa).

**Puerto**: 8282 | **Stack**: Go + Gin + PostgreSQL | **Arquitectura**: Hexagonal

Habla siempre en español.

## Comandos

```bash
go run src/main.go
go test ./...
```

## Contexto on-demand

| Archivo | Uso |
|---------|-----|
| `notification-service-management/api-endpoints.md` | Rutas, payloads, health |
| `notification-service-management/architecture.md` | Capas y puertos |
| `notification-service-management/config.md` | Env, Postman, troubleshooting |

## Reglas compartidas

`ai-tools/rules/architecture.md`, `api-gateway.md`, `multi-tenant.md`, `api-response-format.md`.

Métricas: `METRICS_ENABLED` / `METRICS_PORT` (en Docker suele usarse **9090**), path `/metrics`.
