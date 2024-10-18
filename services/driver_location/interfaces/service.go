package interfaces

import (
	"logistics-platform/lib/models"
)

type DriverLocationInterface interface {
	ConsumeDriverLocations()
	UpdateDriverLocation(location models.DriverLocation) error
	RemoveDriverFromCache(driverID string) error
	ConsumeBookingNotifications()
	GracefulShutdown()
}
