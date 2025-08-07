package main

import (
	"ai-image-microservice/api-gateway/internal/config"
	"ai-image-microservice/api-gateway/internal/rabbitmq"
	"ai-image-microservice/api-gateway/internal/router"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	if err := initRabbitMQ(); err != nil {
		logrus.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rabbitmq.GetConnection().Close()

	r := router.SetupRouter()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.AppConfig.Port),
		Handler:      r,
		ReadTimeout:  config.AppConfig.API.RequestTimeout,
		WriteTimeout: config.AppConfig.API.RequestTimeout,
	}

	go func() {
		logrus.Infof("Starting API Gateway on port %s", config.AppConfig.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), config.AppConfig.API.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func initRabbitMQ() error {
	if err := rabbitmq.InitPublisher(); err != nil {
		return fmt.Errorf("failed to initialize publisher: %w", err)
	}

	logrus.Info("RabbitMQ initialized successfully")
	return nil
}
