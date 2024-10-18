package router

import (
	"logistics-platform/services/admin/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, service interfaces.AdminInterface) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	adminGroup := router.Group("/admin")
	adminGroup.GET("/fleet-stats", service.GetFleetStats)
	adminGroup.GET("/driver-performance", service.GetDriverPerformance)
	adminGroup.GET("/booking-analytics", service.GetBookingAnalytics)
	adminGroup.GET("/vehicle-locations", service.GetVehicleLocations)
	adminGroup.POST("/update-vehicle", service.UpdateVehicle)

}
