package event

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"notification-service/pkg/validator"
	"notification-service/src/notification/application/request"
	"notification-service/src/notification/application/response"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/domain/testutil"
)

// fakeEvent implementa eventbus.DomainEvent para tests (el paquete eventbus no exporta
// constructores de Event a nivel público).
type fakeEvent struct {
	id      string
	aggID   string
	aggType string
	evType  string
	payload []byte
}

func (e fakeEvent) ID() string            { return e.id }
func (e fakeEvent) AggregateID() string   { return e.aggID }
func (e fakeEvent) AggregateType() string { return e.aggType }
func (e fakeEvent) EventType() string     { return e.evType }
func (e fakeEvent) Payload() []byte       { return e.payload }
func (e fakeEvent) OccurredAt() time.Time { return time.Unix(0, 0).UTC() }
func (e fakeEvent) PublishedBy() string   { return "onboarding-service" }

// fakeSender registra las llamadas y devuelve un resultado configurable.
type fakeSender struct {
	calls  []*request.SendNotificationRequest
	result *response.SendNotificationResult
}

func (f *fakeSender) Execute(_ context.Context, req *request.SendNotificationRequest) *response.SendNotificationResult {
	f.calls = append(f.calls, req)
	if f.result != nil {
		return f.result
	}
	return response.NewSendNotificationSuccess(&response.SendNotificationResponse{Status: "sent"})
}

func tenantRegisteredEvent(t *testing.T, eventID string) fakeEvent {
	t.Helper()
	payload, err := json.Marshal(TenantRegisteredPayload{
		Namespace: "mc",
		TenantID:  "tenant-1",
		UserID:    "user-1",
		Type:      "email",
		Action:    "WELCOME",
		Recipient: "owner@example.com",
		Data:      map[string]interface{}{"name": "Ana", "company": "Almacén Ana"},
	})
	assert.NoError(t, err)
	return fakeEvent{
		id:      eventID,
		aggID:   "tenant-1",
		aggType: "tenant",
		evType:  "onboarding.tenant.registered",
		payload: payload,
	}
}

func TestHandle_TenantRegistered_MapsToWelcome(t *testing.T) {
	sender := &fakeSender{}
	h := NewNotificationEventHandler(sender, nil)

	err := h.Handle(context.Background(), tenantRegisteredEvent(t, "evt-1"))

	assert.NoError(t, err)
	assert.Len(t, sender.calls, 1)
	req := sender.calls[0]
	assert.Equal(t, "WELCOME", req.Action)
	assert.Equal(t, "email", req.Type)
	assert.Equal(t, "mc", req.Namespace)
	assert.Equal(t, "tenant-1", req.TenantID)
	assert.Equal(t, "owner@example.com", req.Recipient)
	assert.Equal(t, "evt-1", req.DedupKey, "dedup_key debe ser el event_id para idempotencia at-least-once")
	assert.False(t, req.Async)
	assert.Equal(t, "Ana", req.Data["name"])
}

func TestHandle_UnknownEventType_IsAcked(t *testing.T) {
	sender := &fakeSender{}
	h := NewNotificationEventHandler(sender, nil)

	err := h.Handle(context.Background(), fakeEvent{
		id:      "evt-x",
		evType:  "sales.order.confirmed",
		payload: []byte(`{}`),
	})

	assert.NoError(t, err)
	assert.Empty(t, sender.calls, "eventos de otros dominios no deben disparar envío")
}

func TestHandle_MissingRecipient_IsAcked(t *testing.T) {
	sender := &fakeSender{}
	h := NewNotificationEventHandler(sender, nil)
	payload, _ := json.Marshal(TenantRegisteredPayload{TenantID: "tenant-1", Action: "WELCOME"})

	err := h.Handle(context.Background(), fakeEvent{
		id:      "evt-2",
		evType:  "onboarding.tenant.registered",
		payload: payload,
	})

	assert.NoError(t, err)
	assert.Empty(t, sender.calls)
}

func TestHandle_TransientFailure_ReturnsErrorForRetry(t *testing.T) {
	sender := &fakeSender{result: response.NewInternalServerError("boom")} // 500
	h := NewNotificationEventHandler(sender, nil)

	err := h.Handle(context.Background(), tenantRegisteredEvent(t, "evt-3"))

	assert.Error(t, err, "un fallo 5xx debe devolver error para que el EventBus reintente")
}

func TestHandle_PermanentFailure_IsAcked(t *testing.T) {
	sender := &fakeSender{result: response.NewTemplateNotFoundError()} // 400
	h := NewNotificationEventHandler(sender, nil)

	err := h.Handle(context.Background(), tenantRegisteredEvent(t, "evt-4"))

	assert.NoError(t, err, "un fallo 4xx es permanente → ack para no envenenar la cola")
}

// TestHandle_Replay_DoesNotDuplicate ejercita el handler contra el use case real con un repo
// mock: el primer evento envía; el replay del MISMO event_id es deduplicado (no Save, no Send).
func TestHandle_Replay_DoesNotDuplicate(t *testing.T) {
	repo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()
	emailSender := testutil.NewMockEmailSender()

	dummyTemplate := &domain.Template{ID: "tpl-welcome", Name: "welcome"}
	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).Return(dummyTemplate, nil)

	// 1er evento: no existe → procesa (Save + Send + Update).
	repo.On("ExistsByDedupKey", mock.Anything, "mc", "tenant-1", "evt-dup").Return(false, nil).Once()
	repo.On("Save", mock.Anything, mock.Anything).Return(nil).Once()
	emailSender.On("SendEmailByAction", mock.Anything, "owner@example.com", domain.ActionWelcome, domain.EmailNotification, mock.Anything).Return(nil).Once()
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	// Replay: ya existe → skip (sin Save/Send adicionales).
	repo.On("ExistsByDedupKey", mock.Anything, "mc", "tenant-1", "evt-dup").Return(true, nil).Once()

	uc := usecase.NewSendNotificationUseCase(repo, tmplRepo, emailSender, nil, validator.NewEmailValidator())
	h := NewNotificationEventHandler(uc, nil)

	ev := tenantRegisteredEvent(t, "evt-dup")
	assert.NoError(t, h.Handle(context.Background(), ev))
	assert.NoError(t, h.Handle(context.Background(), ev)) // replay

	emailSender.AssertNumberOfCalls(t, "SendEmailByAction", 1)
	repo.AssertNumberOfCalls(t, "Save", 1)
}
