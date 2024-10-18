package interfaces

import (
	"logistics-platform/lib/utils"
	"logistics-platform/services/booking/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookingInterface interface {
	HandleBookingUpdate(c *gin.Context)
	HandleBookingAccept(c *gin.Context)
	ProcessBooked(bookConReq models.BookingConfirmation) error
	ProduceBookingEvent(userID, driverID, driverName, status string)
	HandleBookingRequest(c *gin.Context)
	ProcessBookingRequest(bookingReq utils.BookingRequest) error
	FindAndNotifyNearbyDrivers(bookingReq utils.BookingRequest, vehicleType string) error
	NotifyDriver(driverID string, bookingReq utils.BookingRequest) error
	HandleUserBookingCheck(c *gin.Context)
	HandleUserBookingHistory(c *gin.Context)
	HandleDriverBookingCheck(c *gin.Context)
	HandleDriverBookingHistory(c *gin.Context)
	GetVehicleType(driverID string) (string, error)
	GracefulShutdown(server *http.Server)
}
