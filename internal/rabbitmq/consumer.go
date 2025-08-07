package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type MessageHandler func(message []byte) error

type Consumer struct {
	conn     *Connection
	handlers map[string]MessageHandler
}

func NewConsumer() *Consumer {
	return &Consumer{
		conn:     GetConnection(),
		handlers: make(map[string]MessageHandler),
	}
}

func (c *Consumer) RegisterHandler(routingKey string, handler MessageHandler) {
	c.handlers[routingKey] = handler
}

func (c *Consumer) StartConsuming(queueName string, routingKeys []string) error {
	queue, err := c.conn.DeclareQueue(queueName)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	for _, key := range routingKeys {
		if err := c.conn.BindQueue(queue.Name, key, ExchangeName); err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	channel, err := c.conn.GetChannel()
	if err != nil {
		return err
	}

	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go c.handleMessages(msgs)

	logrus.Infof("Started consuming from queue: %s", queueName)
	return nil
}

func (c *Consumer) handleMessages(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		go c.processMessage(msg)
	}
}

func (c *Consumer) processMessage(msg amqp.Delivery) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logrus.Errorf("Failed to unmarshal message: %v", err)
		msg.Nack(false, false)
		return
	}

	eventType, ok := event["event_type"].(string)
	if !ok {
		logrus.Error("Event type not found in message")
		msg.Nack(false, false)
		return
	}

	handler, exists := c.handlers[eventType]
	if !exists {
		logrus.Warnf("No handler registered for event type: %s", eventType)
		msg.Nack(false, false)
		return
	}

	if err := handler(msg.Body); err != nil {
		logrus.Errorf("Handler error for %s: %v", eventType, err)
		msg.Nack(false, true) // requeue
		return
	}

	msg.Ack(false)
}

func (c *Consumer) Stop(ctx context.Context) error {
	return c.conn.Close()
}
