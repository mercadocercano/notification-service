package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mercadocercano/eventbus"

	"notification-service/src/notification/application/request"
	"notification-service/src/notification/application/response"
	"notification-service/src/notification/domain/port"
)

// NotificationSender es el subconjunto del use case de envío que el handler necesita.
// Interface chica → el handler es testeable sin construir todo el use case.
// La implementa *usecase.SendNotificationUseCase.
type NotificationSender interface {
	Execute(ctx context.Context, req *request.SendNotificationRequest) *response.SendNotificationResult
}

// TenantRegisteredPayload es el payload de `onboarding.tenant.registered` v1.
// El productor (onboarding) manda parámetros, nunca HTML ni PII fuera de `recipient`.
// `data` es el set de variables que el template resuelve (ownership de presentación
// se mueve al consumer en F3; por ahora se reenvía tal cual para paridad de comportamiento).
type TenantRegisteredPayload struct {
	Namespace string                 `json:"namespace"`
	TenantID  string                 `json:"tenant_id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`      // canal, default "email"
	Action    string                 `json:"action"`    // default "WELCOME"
	Recipient string                 `json:"recipient"` // email destino
	Data      map[string]interface{} `json:"data"`
}

// NotificationEventHandler consume eventos de dominio del EventBus y los mapea a
// notificaciones. Patrón copiado de ledger-service (ConsumerName() + Handle()).
type NotificationEventHandler struct {
	sender      NotificationSender
	eventLogger port.NotificationEventLogger
}

// NewNotificationEventHandler crea el handler. eventLogger puede ser nil (nil-safe).
func NewNotificationEventHandler(sender NotificationSender, eventLogger port.NotificationEventLogger) *NotificationEventHandler {
	return &NotificationEventHandler{
		sender:      sender,
		eventLogger: eventLogger,
	}
}

// ConsumerName identifica al consumidor en el event_consumers del EventBus (idempotencia por consumidor).
func (h *NotificationEventHandler) ConsumerName() string {
	return "notification-service"
}

func (h *NotificationEventHandler) logEvent(e port.NotificationEvent) {
	if h.eventLogger != nil {
		h.eventLogger.Log(e)
	}
}

// Handle rutea por tipo de evento. Eventos desconocidos se ack-ean (return nil) para no
// bloquear el cursor del worker con eventos de otros dominios.
func (h *NotificationEventHandler) Handle(ctx context.Context, event eventbus.DomainEvent) error {
	switch event.EventType() {
	case "onboarding.tenant.registered":
		return h.handleTenantRegistered(ctx, event)
	default:
		h.logEvent(port.NotificationEvent{
			Event:  "notification.event_unknown",
			Reason: "unknown event type: " + event.EventType(),
		})
		return nil
	}
}

// handleTenantRegistered → email de bienvenida (WELCOME). dedup_key = event.ID() para que
// la entrega at-least-once del EventBus no genere correos duplicados.
func (h *NotificationEventHandler) handleTenantRegistered(ctx context.Context, event eventbus.DomainEvent) error {
	var p TenantRegisteredPayload
	if err := json.Unmarshal(event.Payload(), &p); err != nil {
		h.logEvent(port.NotificationEvent{
			Event:  "notification.event_parse_failed",
			Reason: err.Error(),
		})
		// Payload corrupto: ack (return nil) — reintentar no lo va a arreglar (poison message).
		return nil
	}

	if p.Recipient == "" {
		// Sin destinatario no hay nada que enviar; ack para no quedar en loop.
		h.logEvent(port.NotificationEvent{
			Event:    "notification.event_skipped",
			TenantID: p.TenantID,
			UserID:   p.UserID,
			Reason:   "missing recipient",
		})
		return nil
	}

	notifType := p.Type
	if notifType == "" {
		notifType = "email"
	}
	action := p.Action
	if action == "" {
		action = "WELCOME"
	}
	namespace := p.Namespace
	if namespace == "" {
		namespace = "mc"
	}

	req := &request.SendNotificationRequest{
		Namespace: namespace,
		TenantID:  p.TenantID,
		Type:      notifType,
		Action:    action,
		Recipient: p.Recipient,
		Data:      p.Data,
		Async:     false,
		DedupKey:  event.ID(), // idempotencia at-least-once
	}

	h.logEvent(port.NotificationEvent{
		Event:            "notification.event_consumed",
		TenantID:         p.TenantID,
		UserID:           p.UserID,
		NotificationType: notifType,
		Action:           action,
	})

	result := h.sender.Execute(ctx, req)
	if result != nil && !result.Success {
		reason := "unknown"
		if result.Error != nil {
			reason = result.Error.Code
		}
		h.logEvent(port.NotificationEvent{
			Event:            "notification.event_consume_failed",
			TenantID:         p.TenantID,
			UserID:           p.UserID,
			NotificationType: notifType,
			Action:           action,
			Reason:           reason,
		})
		// 5xx = transitorio → devolver error para que el EventBus reintente.
		// 4xx (validación/template) = permanente → ack para no envenenar la cola (F2 añadirá DLQ).
		if result.HTTPStatus >= 500 {
			return fmt.Errorf("notification send failed (transient): %s", reason)
		}
		return nil
	}

	return nil
}
