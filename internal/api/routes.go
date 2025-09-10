package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handler *Handler) {
	// Health check
	router.GET("/health", handler.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Sync routes
		v1.POST("/sync/trigger", handler.TriggerSync)
		v1.GET("/grades/status/:file_id", handler.GetGradesStatus)
	}
}
