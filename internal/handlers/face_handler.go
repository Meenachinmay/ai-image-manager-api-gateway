package handlers

import (
	"ai-image-microservice/api-gateway/internal/models"
	"ai-image-microservice/api-gateway/internal/services"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FaceHandler struct {
	faceService *services.FaceService
}

func NewFaceHandler(faceService *services.FaceService) *FaceHandler {
	return &FaceHandler{
		faceService: faceService,
	}
}

// ProcessImage handles image upload and processing
func (h *FaceHandler) ProcessImage(c *gin.Context) {
	var req models.ProcessImageRequest

	// Parse multipart form
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ProcessImageResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// Validate file size
	if req.Image.Size > 10*1024*1024 { // 10MB limit
		c.JSON(http.StatusBadRequest, models.ProcessImageResponse{
			Success: false,
			Message: "File too large",
			Error:   "Maximum file size is 10MB",
		})
		return
	}

	// Validate mime type
	mimeType := req.Image.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/jpg" {
		c.JSON(http.StatusBadRequest, models.ProcessImageResponse{
			Success: false,
			Message: "Invalid file type",
			Error:   "Only JPEG and PNG images are supported",
		})
		return
	}

	// Open uploaded file
	file, err := req.Image.Open()
	if err != nil {
		logrus.Errorf("Failed to open uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, models.ProcessImageResponse{
			Success: false,
			Message: "Failed to process image",
			Error:   "Could not open uploaded file",
		})
		return
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		logrus.Errorf("Failed to read file: %v", err)
		c.JSON(http.StatusInternalServerError, models.ProcessImageResponse{
			Success: false,
			Message: "Failed to process image",
			Error:   "Could not read file content",
		})
		return
	}

	// Generate image ID
	imageID := uuid.New().String()

	// Parse metadata if provided
	var metadata map[string]interface{}
	if req.Metadata != "" {
		if err := json.Unmarshal([]byte(req.Metadata), &metadata); err != nil {
			logrus.Warnf("Failed to parse metadata: %v", err)
			// Don't fail the request, just log the warning
		}
	}

	// Create event data
	eventData := models.ImageReceivedEventData{
		ImageID:   imageID,
		ImageData: base64.StdEncoding.EncodeToString(imageData),
		FileName:  req.Image.Filename,
		FileSize:  req.Image.Size,
		MimeType:  mimeType,
		UserID:    req.UserID,
		Name:      req.Name,
		Metadata:  metadata,
	}

	// Process image through service
	if err := h.faceService.ProcessImage(c.Request.Context(), eventData); err != nil {
		logrus.Errorf("Failed to process image: %v", err)
		c.JSON(http.StatusInternalServerError, models.ProcessImageResponse{
			Success: false,
			Message: "Failed to process image",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, models.ProcessImageResponse{
		Success: true,
		Message: "Image received and queued for processing",
		ImageID: imageID,
	})
}

// GetProcessingStatus gets the processing status of an image
func (h *FaceHandler) GetProcessingStatus(c *gin.Context) {
	imageID := c.Param("image_id")

	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Image ID is required",
		})
		return
	}

	// For now, return a mock response
	// In production, this would query a status store
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"image_id": imageID,
		"status":   "processing",
		"message":  "Image is being processed",
	})
}
