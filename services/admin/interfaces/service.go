package interfaces

import (
	"github.com/gin-gonic/gin"
)

type AdminInterface interface {
	GetFleetStats(c *gin.Context)
	GetDriverPerformance(c *gin.Context)
	GetBookingAnalytics(c *gin.Context)
	GetVehicleLocations(c *gin.Context)
	UpdateVehicle(c *gin.Context)
}
