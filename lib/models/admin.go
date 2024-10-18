package models

type FleetStats struct {
	TotalVehicles        int            `json:"totalVehicles"`
	ActiveVehicles       int            `json:"activeVehicles"`
	VehicleTypeBreakdown map[string]int `json:"vehicleTypeBreakdown"`
}

type DriverPerformance struct {
	DriverID     int     `json:"driverID"`
	Name         string  `json:"name"`
	TripCount    int     `json:"tripCount"`
	AvgTripTime  float64 `json:"avgTripTime"`
	TotalRevenue float64 `json:"totalRevenue"`
}

type BookingAnalytics struct {
	TotalBookings     int     `json:"totalBookings"`
	CompletedBookings int     `json:"completedBookings"`
	CancelledBookings int     `json:"cancelledBookings"`
	AvgTripTime       float64 `json:"avgTripTime"`
	TotalRevenue      float64 `json:"totalRevenue"`
}
