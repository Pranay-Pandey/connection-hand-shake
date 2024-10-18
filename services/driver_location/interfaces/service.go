package interfaces

import (
	"logistics-platform/lib/utils"
)

type DriverLocationInterface interface {
	ConsumeDriverLocations()
	UpdateDriverLocation(location utils.DriverLocation) error
	RemoveDriverFromCache(driverID string) error
	ConsumeBookingNotifications()
	GracefulShutdown()
}
