package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Env      string
	Port     string
	LogLevel string
	RabbitMQ RabbitMQConfig
	API      APIConfig
}

type RabbitMQConfig struct {
	URL            string
	MaxRetries     int
	RetryDelay     time.Duration
	PublishTimeout time.Duration
	ConsumeTimeout time.Duration
	PrefetchCount  int
	ReconnectDelay time.Duration
}

type APIConfig struct {
	MaxUploadSize   int64
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

var AppConfig *Config

func LoadConfig() error {
	_ = godotenv.Load()

	AppConfig = &Config{
		Env:      getEnv("ENV", "development"),
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		RabbitMQ: RabbitMQConfig{
			URL:            getEnv("RABBITMQ_URL", "amqp://admin:admin123@localhost:5672/"),
			MaxRetries:     getEnvAsInt("RABBITMQ_MAX_RETRIES", 3),
			RetryDelay:     time.Duration(getEnvAsInt("RABBITMQ_RETRY_DELAY", 5)) * time.Second,
			PublishTimeout: time.Duration(getEnvAsInt("RABBITMQ_PUBLISH_TIMEOUT", 30)) * time.Second,
			ConsumeTimeout: time.Duration(getEnvAsInt("RABBITMQ_CONSUME_TIMEOUT", 30)) * time.Second,
			PrefetchCount:  getEnvAsInt("RABBITMQ_PREFETCH_COUNT", 10),
			ReconnectDelay: time.Duration(getEnvAsInt("RABBITMQ_RECONNECT_DELAY", 5)) * time.Second,
		},
		API: APIConfig{
			MaxUploadSize:   getEnvAsInt64("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB default
			RequestTimeout:  time.Duration(getEnvAsInt("REQUEST_TIMEOUT", 30)) * time.Second,
			ShutdownTimeout: time.Duration(getEnvAsInt("SHUTDOWN_TIMEOUT", 10)) * time.Second,
		},
	}

	setLogLevel(AppConfig.LogLevel)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func setLogLevel(level string) {
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
