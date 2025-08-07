package router

import (
	"ai-image-microservice/api-gateway/internal/handlers"
	"ai-image-microservice/api-gateway/internal/middleware"
	"ai-image-microservice/api-gateway/internal/services"
	_ "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	faceService := services.NewFaceService()

	healthHandler := handlers.NewHealthHandler()
	faceHandler := handlers.NewFaceHandler(faceService)

	v1 := router.Group("/api/v1")
	{
		health := v1.Group("/health")
		{
			health.GET("/", healthHandler.Health)
			health.GET("/ready", healthHandler.Ready)
		}

		face := v1.Group("/face")
		{
			face.POST("/process", faceHandler.ProcessImage)
			face.GET("/status/:image_id", faceHandler.GetProcessingStatus)
		}
	}

	return router
}
