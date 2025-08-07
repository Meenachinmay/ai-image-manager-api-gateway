package services

import (
	"ai-image-microservice/api-gateway/internal/models"
	"ai-image-microservice/api-gateway/internal/rabbitmq"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

type FaceService struct {
	publisher *rabbitmq.Publisher
}

func NewFaceService() *FaceService {
	return &FaceService{
		publisher: rabbitmq.GetPublisher(),
	}
}

func (s *FaceService) ProcessImage(ctx context.Context, imageData models.ImageReceivedEventData) error {
	if imageData.ImageID == "" {
		return fmt.Errorf("image ID is required")
	}

	if imageData.ImageData == "" {
		return fmt.Errorf("image data is required")
	}

	if err := s.publisher.PublishEventWithContext(
		ctx,
		rabbitmq.TopicImageReceived,
		imageData,
	); err != nil {
		return fmt.Errorf("failed to publish image received event: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"image_id":  imageData.ImageID,
		"file_name": imageData.FileName,
		"file_size": imageData.FileSize,
	}).Info("Image processing initiated")

	return nil
}
