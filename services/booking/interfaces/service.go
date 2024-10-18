package interfaces

import (
	"logistics-platform/lib/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookingInterface interface {
	HandleBookingUpdate(c *gin.Context)
	HandleBookingAccept(c *gin.Context)
	ProcessBooked(bookConReq models.BookingConfirmation) error
	ProduceBookingEvent(userID, driverID, driverName, status string)
	HandleBookingRequest(c *gin.Context)
	ProcessBookingRequest(bookingReq models.BookingRequest) error
	FindAndNotifyNearbyDrivers(bookingReq models.BookingRequest, vehicleType string) error
	NotifyDriver(driverID string, bookingReq models.BookingRequest) error
	HandleUserBookingCheck(c *gin.Context)
	HandleUserBookingHistory(c *gin.Context)
	HandleDriverBookingCheck(c *gin.Context)
	HandleDriverBookingHistory(c *gin.Context)
	GetVehicleType(driverID string) (string, error)
	GracefulShutdown(server *http.Server)
}
