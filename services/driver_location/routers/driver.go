package routers

import (
	"logistics-platform/services/driver_location/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterLocationRoutes(router *gin.Engine) {
	location := router.Group("/location")
	{
		location.GET("/connect", handlers.ConnectDriverWebSocket)
		location.POST("/update", handlers.UpdateLocation)
		location.GET("/health", handlers.HealthCheck)
	}
}
