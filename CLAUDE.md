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

## Memoria persistente (Engram)

Tenés acceso a memoria persistente entre sesiones vía las herramientas MCP de Engram (`mem_save`, `mem_search`, `mem_context`, etc.). Proyecto: **`mercado-cercano`** — memoria compartida con el resto del ecosistema.

**Cuándo guardar** — sin esperar que te lo pidan:
- Al resolver un bug no trivial: síntoma, causa raíz, fix aplicado.
- Al tomar una decisión de diseño: qué se decidió y por qué.
- Al descubrir un patrón o convención del proyecto que no está documentada.
- Al completar una feature o refactor significativo: qué cambió y dónde.

**Cuándo buscar** — antes de empezar cualquier tarea:
- `mem_context` al inicio de sesión o tras una compaction para recuperar el estado anterior.
- `mem_search` cuando el usuario menciona algo que puede tener historial ("el bug de autenticación", "la migración de la semana pasada").

**Al cerrar sesión**: llamar `mem_session_summary` para dejar un resumen recuperable.
