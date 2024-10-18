package models

import "logistics-platform/lib/utils"

type BookingConfirmation struct {
	BookingReq utils.BookingRequest `json:"booking_request"`
	DriverID   string               `json:"driver_id"`
	DriverName string               `json:"driver_name"`
}
