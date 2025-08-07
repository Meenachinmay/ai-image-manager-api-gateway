package rabbitmq

import (
	"ai-image-microservice/api-gateway/internal/config"
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

type Connection struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	url             string
	mu              sync.RWMutex
	reconnectMu     sync.Mutex
	isReconnecting  bool
	notifyClose     chan *amqp.Error
	notifyReconnect chan bool
}

var (
	instance *Connection
	once     sync.Once
)

func GetConnection() *Connection {
	once.Do(func() {
		instance = &Connection{
			url:             config.AppConfig.RabbitMQ.URL,
			notifyReconnect: make(chan bool),
		}
		if err := instance.Connect(); err != nil {
			logrus.Fatalf("Failed to connect to RabbitMQ: %v", err)
		}
		go instance.handleReconnect()
	})
	return instance
}

func (c *Connection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Set QoS
	err = channel.Qos(
		config.AppConfig.RabbitMQ.PrefetchCount, // prefetch count
		0,                                       // prefetch size
		false,                                   // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	c.conn = conn
	c.channel = channel
	c.notifyClose = make(chan *amqp.Error)
	c.conn.NotifyClose(c.notifyClose)

	logrus.Info("Successfully connected to RabbitMQ")
	return nil
}

func (c *Connection) handleReconnect() {
	for {
		select {
		case err := <-c.notifyClose:
			if err != nil {
				logrus.Errorf("RabbitMQ connection closed: %v", err)
				c.reconnect()
			}
		}
	}
}

func (c *Connection) reconnect() {
	c.reconnectMu.Lock()
	if c.isReconnecting {
		c.reconnectMu.Unlock()
		return
	}
	c.isReconnecting = true
	c.reconnectMu.Unlock()

	defer func() {
		c.reconnectMu.Lock()
		c.isReconnecting = false
		c.reconnectMu.Unlock()
	}()

	for {
		logrus.Info("Attempting to reconnect to RabbitMQ...")

		if err := c.Connect(); err != nil {
			logrus.Errorf("Failed to reconnect: %v", err)
			time.Sleep(config.AppConfig.RabbitMQ.ReconnectDelay)
			continue
		}

		c.notifyReconnect <- true
		break
	}
}

func (c *Connection) GetChannel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}
	return c.channel, nil
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			logrus.Errorf("Failed to close channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}

// DeclareExchange declares an exchange
func (c *Connection) DeclareExchange(name, kind string) error {
	channel, err := c.GetChannel()
	if err != nil {
		return err
	}

	return channel.ExchangeDeclare(
		name,  // name
		kind,  // type
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
}

// DeclareQueue declares a queue
func (c *Connection) DeclareQueue(name string) (amqp.Queue, error) {
	channel, err := c.GetChannel()
	if err != nil {
		return amqp.Queue{}, err
	}

	return channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

// BindQueue binds a queue to an exchange
func (c *Connection) BindQueue(queueName, routingKey, exchangeName string) error {
	channel, err := c.GetChannel()
	if err != nil {
		return err
	}

	return channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
}

// PublishWithContext publishes a message with context
func (c *Connection) PublishWithContext(ctx context.Context, exchange, routingKey string, message []byte) error {
	channel, err := c.GetChannel()
	if err != nil {
		return err
	}

	return channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}
