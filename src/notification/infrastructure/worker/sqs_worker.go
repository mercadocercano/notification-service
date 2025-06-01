package worker

import (
	"context"
	"time"

	"notification-service/src/notification/domain"
	"notification-service/src/notification/ports/output"
	"notification-service/src/shared/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SQSWorker struct {
	queue            output.Queue
	emailSender      output.EmailSender
	notificationRepo domain.NotificationRepository
	logger           *zap.Logger
	stopChan         chan struct{}
	running          bool
}

// NewSQSWorker crea una nueva instancia del worker SQS
func NewSQSWorker(
	queue output.Queue,
	emailSender output.EmailSender,
	notificationRepo domain.NotificationRepository,
) *SQSWorker {
	return &SQSWorker{
		queue:            queue,
		emailSender:      emailSender,
		notificationRepo: notificationRepo,
		logger:           logger.GetLogger(),
		stopChan:         make(chan struct{}),
		running:          false,
	}
}

// Start inicia el worker para procesar mensajes de SQS
func (w *SQSWorker) Start(ctx context.Context) {
	if w.running {
		w.logger.Warn("SQS Worker is already running")
		return
	}

	w.running = true
	w.logger.Info("Starting SQS Worker")

	go w.processMessages(ctx)
}

// Stop detiene el worker
func (w *SQSWorker) Stop() {
	if !w.running {
		w.logger.Warn("SQS Worker is not running")
		return
	}

	w.logger.Info("Stopping SQS Worker")
	close(w.stopChan)
	w.running = false
}

// processMessages procesa mensajes de la cola SQS continuamente
func (w *SQSWorker) processMessages(ctx context.Context) {
	w.logger.Info("SQS Worker started and listening for messages")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("SQS Worker stopped due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Info("SQS Worker stopped")
			return
		default:
			w.processNextMessage(ctx)
		}
	}
}

// processNextMessage procesa el siguiente mensaje disponible en la cola
func (w *SQSWorker) processNextMessage(ctx context.Context) {
	// Crear un contexto con timeout para la operación de dequeue
	dequeueCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Intentar recibir un mensaje de la cola
	notificationRequest, err := w.queue.Dequeue(dequeueCtx)
	if err != nil {
		w.logger.Error("Failed to dequeue message from SQS", zap.Error(err))
		// Esperar un poco antes del siguiente intento
		time.Sleep(5 * time.Second)
		return
	}

	// Si no hay mensajes, continúa el bucle
	if notificationRequest == nil {
		w.logger.Debug("No messages available in queue")
		return
	}

	// Crear una notificación completa a partir del request limpio
	// Todo lo que viene de SQS es inherentemente asíncrono
	notification := &domain.Notification{
		ID:        notificationRequest.NotificationID,
		Type:      domain.NotificationType(notificationRequest.Type),
		Action:    domain.NotificationAction(notificationRequest.Action),
		Recipient: notificationRequest.Recipient,
		Status:    domain.StatusPending,
		Data:      notificationRequest.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Si el notification_id está vacío, generar uno nuevo (fallback)
	if notification.ID == "" {
		notification.ID = uuid.New().String()
		w.logger.Warn("No notification_id in SQS message, generating new one",
			zap.String("generated_id", notification.ID))
	}

	// Procesar la notificación
	w.logger.Info("Processing notification from SQS",
		zap.String("notification_id", notification.ID),
		zap.String("action", string(notification.Action)),
		zap.String("recipient", notification.Recipient))

	// Actualizar estado a "processing" si tenemos repositorio
	if w.notificationRepo != nil {
		notification.Status = domain.StatusPending
		if err := w.notificationRepo.Update(ctx, notification); err != nil {
			w.logger.Error("Failed to update notification status to processing",
				zap.String("notification_id", notification.ID),
				zap.Error(err))
		}
	}

	// Enviar la notificación
	if err := w.sendNotification(ctx, notification); err != nil {
		w.logger.Error("Failed to send notification",
			zap.String("notification_id", notification.ID),
			zap.String("action", string(notification.Action)),
			zap.Error(err))

		// Actualizar estado a "failed"
		notification.Status = domain.StatusFailed
		notification.Error = err.Error()
		if w.notificationRepo != nil {
			if updateErr := w.notificationRepo.Update(ctx, notification); updateErr != nil {
				w.logger.Error("Failed to update notification status to failed",
					zap.String("notification_id", notification.ID),
					zap.Error(updateErr))
			}
		}
		return
	}

	// Actualizar estado a "sent"
	w.logger.Info("Notification sent successfully",
		zap.String("notification_id", notification.ID),
		zap.String("action", string(notification.Action)))

	notification.Status = domain.StatusSent
	notification.Error = ""
	if w.notificationRepo != nil {
		if err := w.notificationRepo.Update(ctx, notification); err != nil {
			w.logger.Error("Failed to update notification status to sent",
				zap.String("notification_id", notification.ID),
				zap.Error(err))
		}
	}
}

// sendNotification envía la notificación usando el servicio apropiado
func (w *SQSWorker) sendNotification(ctx context.Context, notification *domain.Notification) error {
	w.logger.Debug("Sending notification",
		zap.String("notification_id", notification.ID),
		zap.String("type", string(notification.Type)),
		zap.String("action", string(notification.Action)),
		zap.String("recipient", notification.Recipient))

	switch notification.Type {
	case domain.EmailNotification:
		return w.emailSender.SendEmailByAction(ctx, notification.Recipient, notification.Action, notification.Type, notification.Data)
	default:
		w.logger.Warn("Unknown notification type", zap.String("type", string(notification.Type)))
		return nil // Ya validado anteriormente
	}
}

// IsRunning devuelve true si el worker está ejecutándose
func (w *SQSWorker) IsRunning() bool {
	return w.running
}

// GetQueueSize devuelve el tamaño actual de la cola
func (w *SQSWorker) GetQueueSize(ctx context.Context) (int64, error) {
	return w.queue.Size(ctx)
}
