package interfaces

import (
	"logistics-platform/lib/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationInterface interface {
	ConsumeNotifications()
	ConsumeBookingNotifications()
	NotifyUser(userID string, notification utils.BookedNotification)
	HandleDriverWebSocket(c *gin.Context)
	SendLocationUpdate(location utils.DriverLocation) error
	SendNotification(notification utils.BookingNotification) error
	HandleUserWebSocket(c *gin.Context)
	GracefulShutdown(server *http.Server)
}
