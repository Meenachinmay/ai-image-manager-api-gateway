package handlers

import (
	"ai-image-microservice/api-gateway/internal/models"
	"ai-image-microservice/api-gateway/internal/rabbitmq"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	services := make(map[string]string)

	conn := rabbitmq.GetConnection()
	if _, err := conn.GetChannel(); err != nil {
		services["rabbitmq"] = "unhealthy"
	} else {
		services["rabbitmq"] = "healthy"
	}

	status := "healthy"
	for _, v := range services {
		if v != "healthy" {
			status = "degraded"
			break
		}
	}

	c.JSON(http.StatusOK, models.HealthCheckResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	conn := rabbitmq.GetConnection()
	if _, err := conn.GetChannel(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready": false,
			"error": "RabbitMQ not connected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}
