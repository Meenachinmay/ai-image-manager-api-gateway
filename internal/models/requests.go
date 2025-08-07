package models

import (
	"mime/multipart"
)

type ProcessImageRequest struct {
	Image    *multipart.FileHeader `form:"image" binding:"required"`
	Name     string                `form:"name"`
	UserID   string                `form:"user_id"`
	Metadata string                `form:"metadata"` // JSON string
}

type ProcessImageResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	ImageID string      `json:"image_id,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Services  map[string]string `json:"services"`
}
