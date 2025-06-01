package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"notification-service/src/notification/application/request"
	"notification-service/src/notification/domain"
	"notification-service/src/notification/ports/output"
	"notification-service/src/shared/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

type sqsQueue struct {
	client   *sqs.SQS
	queueURL string
	logger   *zap.Logger
}

type SQSConfig struct {
	QueueURL string
	Region   string
}

// NewSQSQueue crea una nueva instancia de cola SQS
func NewSQSQueue(config SQSConfig) (output.Queue, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &sqsQueue{
		client:   sqs.New(sess),
		queueURL: config.QueueURL,
		logger:   logger.GetLogger(),
	}, nil
}

// Enqueue envía una notificación a la cola SQS
func (q *sqsQueue) Enqueue(ctx context.Context, notification *domain.Notification) error {
	q.logger.Info("Enqueueing notification to SQS",
		zap.String("notification_id", notification.ID),
		zap.String("queue_url", q.queueURL))

	// Crear el formato limpio del request (como el que recibe el endpoint)
	requestFormat := map[string]interface{}{
		"notification_id": notification.ID, // ID para tracking en base de datos
		"type":            string(notification.Type),
		"action":          string(notification.Action),
		"recipient":       notification.Recipient,
		"data":            notification.Data,
	}

	// Serializar el formato limpio a JSON
	messageBody, err := json.Marshal(requestFormat)
	if err != nil {
		q.logger.Error("Failed to marshal notification request for SQS",
			zap.String("notification_id", notification.ID),
			zap.Error(err))
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	// Preparar el mensaje SQS
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.queueURL),
		MessageBody: aws.String(string(messageBody)),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"notification_id": {
				DataType:    aws.String("String"),
				StringValue: aws.String(notification.ID),
			},
			"action": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(notification.Action)),
			},
			"type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(notification.Type)),
			},
		},
	}

	// Enviar mensaje a SQS
	result, err := q.client.SendMessageWithContext(ctx, input)
	if err != nil {
		q.logger.Error("Failed to send message to SQS",
			zap.String("notification_id", notification.ID),
			zap.String("queue_url", q.queueURL),
			zap.Error(err))
		return fmt.Errorf("failed to send message to SQS: %w", err)
	}

	q.logger.Info("Message sent to SQS successfully",
		zap.String("notification_id", notification.ID),
		zap.String("message_id", *result.MessageId),
		zap.String("queue_url", q.queueURL))

	return nil
}

// Dequeue recibe una notificación de la cola SQS
func (q *sqsQueue) Dequeue(ctx context.Context) (*request.SendNotificationRequest, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.queueURL),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(20), // Long polling
		MessageAttributeNames: []*string{
			aws.String("All"),
		},
	}

	result, err := q.client.ReceiveMessageWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message from SQS: %w", err)
	}

	// Si no hay mensajes, retorna nil
	if len(result.Messages) == 0 {
		return nil, nil
	}

	message := result.Messages[0]

	// Deserializar el formato de request limpio
	var notificationRequest request.SendNotificationRequest
	if err := json.Unmarshal([]byte(*message.Body), &notificationRequest); err != nil {
		q.logger.Error("Failed to unmarshal notification request from SQS",
			zap.String("message_body", *message.Body),
			zap.Error(err))

		// Eliminar mensaje malformado
		if deleteErr := q.deleteMessage(ctx, message.ReceiptHandle); deleteErr != nil {
			q.logger.Error("Failed to delete malformed message", zap.Error(deleteErr))
		}
		return nil, fmt.Errorf("failed to unmarshal notification request: %w", err)
	}

	// Obtener el notification_id de los atributos del mensaje (si existe)
	var notificationID string
	if attrs := message.MessageAttributes; attrs != nil {
		if idAttr := attrs["notification_id"]; idAttr != nil && idAttr.StringValue != nil {
			notificationID = *idAttr.StringValue
		}
	}

	// Log exitoso
	q.logger.Info("Message dequeued from SQS successfully",
		zap.String("notification_id", notificationID),
		zap.String("message_id", *message.MessageId))

	// Eliminar el mensaje de la cola
	if err := q.deleteMessage(ctx, message.ReceiptHandle); err != nil {
		q.logger.Error("Failed to delete message from SQS after processing",
			zap.String("notification_id", notificationID),
			zap.Error(err))
		// No retornamos error aquí porque ya procesamos el mensaje
	}

	return &notificationRequest, nil
}

// Size obtiene el número aproximado de mensajes en la cola
func (q *sqsQueue) Size(ctx context.Context) (int64, error) {
	q.logger.Debug("Getting queue size", zap.String("queue_url", q.queueURL))

	input := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(q.queueURL),
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
		},
	}

	result, err := q.client.GetQueueAttributesWithContext(ctx, input)
	if err != nil {
		q.logger.Error("Failed to get queue attributes",
			zap.String("queue_url", q.queueURL),
			zap.Error(err))
		return 0, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	// Extraer el número de mensajes
	if countStr, exists := result.Attributes["ApproximateNumberOfMessages"]; exists {
		var count int64
		if _, err := fmt.Sscanf(*countStr, "%d", &count); err != nil {
			q.logger.Error("Failed to parse message count", zap.Error(err))
			return 0, fmt.Errorf("failed to parse message count: %w", err)
		}

		q.logger.Debug("Queue size retrieved",
			zap.String("queue_url", q.queueURL),
			zap.Int64("size", count))

		return count, nil
	}

	return 0, fmt.Errorf("ApproximateNumberOfMessages attribute not found")
}

// deleteMessage elimina un mensaje de la cola SQS
func (q *sqsQueue) deleteMessage(ctx context.Context, receiptHandle *string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.queueURL),
		ReceiptHandle: receiptHandle,
	}

	_, err := q.client.DeleteMessageWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
