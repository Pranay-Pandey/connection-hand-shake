package interfaces

import (
	"logistics-platform/lib/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationInterface interface {
	ConsumeNotifications()
	ConsumeBookingNotifications()
	NotifyUser(userID string, notification models.BookedNotification)
	HandleDriverWebSocket(c *gin.Context)
	SendLocationUpdate(location models.DriverLocation) error
	SendNotification(notification models.BookingNotification) error
	HandleUserWebSocket(c *gin.Context)
	GracefulShutdown(server *http.Server)
}
