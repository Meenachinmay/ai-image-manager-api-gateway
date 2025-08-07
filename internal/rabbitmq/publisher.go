package rabbitmq

import (
	"ai-image-microservice/api-gateway/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	ExchangeName = "image_processing"
	ExchangeType = "topic"
)

const (
	TopicImageReceived   = "image.received"
	TopicFaceRecognition = "face.recognition"
	TopicDataSaved       = "data.saved"
)

type Publisher struct {
	conn *Connection
}

var publisher *Publisher

func InitPublisher() error {
	conn := GetConnection()

	if err := conn.DeclareExchange(ExchangeName, ExchangeType); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	publisher = &Publisher{
		conn: conn,
	}

	logrus.Info("RabbitMQ Publisher initialized")
	return nil
}

func GetPublisher() *Publisher {
	if publisher == nil {
		if err := InitPublisher(); err != nil {
			logrus.Fatalf("Failed to initialize publisher: %v", err)
		}
	}
	return publisher
}

func (p *Publisher) PublishEvent(topic string, data interface{}) error {
	return p.PublishEventWithContext(context.Background(), topic, data)
}

func (p *Publisher) PublishEventWithContext(ctx context.Context, topic string, data interface{}) error {
	event := map[string]interface{}{
		"event_id":   uuid.New().String(),
		"event_type": topic,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"data":       data,
	}

	message, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	publishCtx, cancel := context.WithTimeout(ctx, config.AppConfig.RabbitMQ.PublishTimeout)
	defer cancel()

	var lastErr error
	for i := 0; i < config.AppConfig.RabbitMQ.MaxRetries; i++ {
		if err := p.conn.PublishWithContext(publishCtx, ExchangeName, topic, message); err != nil {
			lastErr = err
			logrus.Warnf("Failed to publish message (attempt %d/%d): %v", i+1, config.AppConfig.RabbitMQ.MaxRetries, err)
			time.Sleep(config.AppConfig.RabbitMQ.RetryDelay)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"topic":    topic,
			"event_id": event["event_id"],
		}).Debug("Event published successfully")

		return nil
	}

	return fmt.Errorf("failed to publish after %d attempts: %w", config.AppConfig.RabbitMQ.MaxRetries, lastErr)
}

func PublishImageReceived(imageData interface{}) error {
	return GetPublisher().PublishEvent(TopicImageReceived, imageData)
}

func PublishFaceRecognition(faceData interface{}) error {
	return GetPublisher().PublishEvent(TopicFaceRecognition, faceData)
}

func PublishDataSaved(savedData interface{}) error {
	return GetPublisher().PublishEvent(TopicDataSaved, savedData)
}
