package usecase_test

import (
	"context"
	"errors"
	"testing"

	"notification-service/pkg/validator"
	"notification-service/src/notification/application/request"
	"notification-service/src/notification/application/usecase"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/domain/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSendUseCase(
	notifRepo *testutil.MockNotificationRepository,
	tmplRepo *testutil.MockTemplateRepository,
	emailSender *testutil.MockEmailSender,
	queue *testutil.MockQueue,
) *usecase.SendNotificationUseCase {
	return usecase.NewSendNotificationUseCase(
		notifRepo,
		tmplRepo,
		emailSender,
		queue,
		validator.NewEmailValidator(),
	)
}

func validSendRequest() *request.SendNotificationRequest {
	return &request.SendNotificationRequest{
		Type:      "email",
		Action:    "WELCOME",
		Recipient: "user@example.com",
		Data:      map[string]interface{}{"name": "Juan"},
		Async:     false,
	}
}

// --- Sync Send Tests ---

func TestSendNotificationUseCase_Execute_SyncSuccess_ReturnsSuccess(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()
	emailSender := testutil.NewMockEmailSender()
	queue := testutil.NewMockQueue()

	tmplMother := testutil.NewTemplateMother()
	tmpl := tmplMother.Default()

	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(tmpl, nil)
	notifRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)
	emailSender.On("SendEmailByAction", mock.Anything, "user@example.com", domain.ActionWelcome, domain.EmailNotification, mock.Anything).
		Return(nil)
	notifRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)

	uc := newSendUseCase(notifRepo, tmplRepo, emailSender, queue)

	// Act
	result := uc.Execute(context.Background(), validSendRequest())

	// Assert
	require.True(t, result.Success)
	require.NotNil(t, result.Data)
	assert.Equal(t, "sent", result.Data.Status)
	assert.NotEmpty(t, result.Data.ID)

	notifRepo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	emailSender.AssertCalled(t, "SendEmailByAction", mock.Anything, "user@example.com", domain.ActionWelcome, domain.EmailNotification, mock.Anything)
	notifRepo.AssertNumberOfCalls(t, "Update", 1)
}

func TestSendNotificationUseCase_Execute_AsyncSuccess_ReturnsQueued(t *testing.T) {
	// Arrange
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()
	emailSender := testutil.NewMockEmailSender()
	queue := testutil.NewMockQueue()

	tmplMother := testutil.NewTemplateMother()
	tmpl := tmplMother.Default()

	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(tmpl, nil)
	notifRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)
	queue.On("Enqueue", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)
	notifRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)

	uc := newSendUseCase(notifRepo, tmplRepo, emailSender, queue)
	req := validSendRequest()
	req.Async = true

	// Act
	result := uc.Execute(context.Background(), req)

	// Assert
	require.True(t, result.Success)
	assert.Equal(t, "queued", result.Data.Status)
	queue.AssertCalled(t, "Enqueue", mock.Anything, mock.Anything)
	emailSender.AssertNotCalled(t, "SendEmailByAction")
}

// --- Validation Error Tests ---

func TestSendNotificationUseCase_Execute_WithEmptyType_ReturnsInvalidTypeError(t *testing.T) {
	uc := newSendUseCase(
		testutil.NewMockNotificationRepository(),
		testutil.NewMockTemplateRepository(),
		testutil.NewMockEmailSender(),
		testutil.NewMockQueue(),
	)

	req := validSendRequest()
	req.Type = ""

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_NOTIFICATION_TYPE", result.Error.Code)
}

func TestSendNotificationUseCase_Execute_WithNonEmailType_ReturnsInvalidTypeError(t *testing.T) {
	uc := newSendUseCase(
		testutil.NewMockNotificationRepository(),
		testutil.NewMockTemplateRepository(),
		testutil.NewMockEmailSender(),
		testutil.NewMockQueue(),
	)

	req := validSendRequest()
	req.Type = "sms"

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_NOTIFICATION_TYPE", result.Error.Code)
}

func TestSendNotificationUseCase_Execute_WithEmptyRecipient_ReturnsInvalidEmailError(t *testing.T) {
	uc := newSendUseCase(
		testutil.NewMockNotificationRepository(),
		testutil.NewMockTemplateRepository(),
		testutil.NewMockEmailSender(),
		testutil.NewMockQueue(),
	)

	req := validSendRequest()
	req.Recipient = ""

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_EMAIL", result.Error.Code)
}

func TestSendNotificationUseCase_Execute_WithInvalidEmail_ReturnsInvalidEmailError(t *testing.T) {
	uc := newSendUseCase(
		testutil.NewMockNotificationRepository(),
		testutil.NewMockTemplateRepository(),
		testutil.NewMockEmailSender(),
		testutil.NewMockQueue(),
	)

	req := validSendRequest()
	req.Recipient = "not-an-email"

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_EMAIL", result.Error.Code)
}

func TestSendNotificationUseCase_Execute_WithEmptyAction_ReturnsError(t *testing.T) {
	uc := newSendUseCase(
		testutil.NewMockNotificationRepository(),
		testutil.NewMockTemplateRepository(),
		testutil.NewMockEmailSender(),
		testutil.NewMockQueue(),
	)

	req := validSendRequest()
	req.Action = ""

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_NOTIFICATION_TYPE", result.Error.Code)
}

func TestSendNotificationUseCase_Execute_WithInvalidAction_ReturnsError(t *testing.T) {
	uc := newSendUseCase(
		testutil.NewMockNotificationRepository(),
		testutil.NewMockTemplateRepository(),
		testutil.NewMockEmailSender(),
		testutil.NewMockQueue(),
	)

	req := validSendRequest()
	req.Action = "INVALID_ACTION"

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INVALID_NOTIFICATION_TYPE", result.Error.Code)
}

// --- Template Error Tests ---

func TestSendNotificationUseCase_Execute_WhenTemplateNotFound_ReturnsTemplateError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()

	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(nil, nil)

	uc := newSendUseCase(notifRepo, tmplRepo, testutil.NewMockEmailSender(), testutil.NewMockQueue())

	result := uc.Execute(context.Background(), validSendRequest())

	assert.False(t, result.Success)
	assert.Equal(t, "TEMPLATE_NOT_FOUND", result.Error.Code)
}

func TestSendNotificationUseCase_Execute_WhenTemplateRepoFails_ReturnsInternalError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()

	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(nil, errors.New("db error"))

	uc := newSendUseCase(notifRepo, tmplRepo, testutil.NewMockEmailSender(), testutil.NewMockQueue())

	result := uc.Execute(context.Background(), validSendRequest())

	assert.False(t, result.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Error.Code)
}

// --- Repository Error Tests ---

func TestSendNotificationUseCase_Execute_WhenSaveFails_ReturnsInternalError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()

	tmplMother := testutil.NewTemplateMother()
	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(tmplMother.Default(), nil)
	notifRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(errors.New("save failed"))

	uc := newSendUseCase(notifRepo, tmplRepo, testutil.NewMockEmailSender(), testutil.NewMockQueue())

	result := uc.Execute(context.Background(), validSendRequest())

	assert.False(t, result.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Error.Code)
	assert.Contains(t, result.Error.Details, "guardar")
}

// --- Email Sending Error Tests ---

func TestSendNotificationUseCase_Execute_WhenEmailSendFails_ReturnsInternalError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()
	emailSender := testutil.NewMockEmailSender()

	tmplMother := testutil.NewTemplateMother()
	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(tmplMother.Default(), nil)
	notifRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)
	emailSender.On("SendEmailByAction", mock.Anything, "user@example.com", domain.ActionWelcome, domain.EmailNotification, mock.Anything).
		Return(errors.New("SMTP error"))
	notifRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)

	uc := newSendUseCase(notifRepo, tmplRepo, emailSender, testutil.NewMockQueue())

	result := uc.Execute(context.Background(), validSendRequest())

	assert.False(t, result.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Error.Code)
	assert.Contains(t, result.Error.Details, "enviar")
}

// --- Queue Error Tests ---

func TestSendNotificationUseCase_Execute_WhenEnqueueFails_ReturnsInternalError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()
	queue := testutil.NewMockQueue()

	tmplMother := testutil.NewTemplateMother()
	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(tmplMother.Default(), nil)
	notifRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)
	queue.On("Enqueue", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(errors.New("SQS error"))

	uc := newSendUseCase(notifRepo, tmplRepo, testutil.NewMockEmailSender(), queue)
	req := validSendRequest()
	req.Async = true

	result := uc.Execute(context.Background(), req)

	assert.False(t, result.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Error.Code)
	assert.Contains(t, result.Error.Details, "encolar")
}

// --- Update Status Error Tests ---

func TestSendNotificationUseCase_Execute_WhenUpdateAfterSendFails_ReturnsInternalError(t *testing.T) {
	notifRepo := testutil.NewMockNotificationRepository()
	tmplRepo := testutil.NewMockTemplateRepository()
	emailSender := testutil.NewMockEmailSender()

	tmplMother := testutil.NewTemplateMother()
	tmplRepo.On("FindByAction", mock.Anything, domain.ActionWelcome, domain.EmailNotification).
		Return(tmplMother.Default(), nil)
	notifRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(nil)
	emailSender.On("SendEmailByAction", mock.Anything, "user@example.com", domain.ActionWelcome, domain.EmailNotification, mock.Anything).
		Return(nil)
	notifRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Notification")).
		Return(errors.New("update failed"))

	uc := newSendUseCase(notifRepo, tmplRepo, emailSender, testutil.NewMockQueue())

	result := uc.Execute(context.Background(), validSendRequest())

	assert.False(t, result.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", result.Error.Code)
	assert.Contains(t, result.Error.Details, "actualizar")
}
