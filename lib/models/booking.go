package models

import "time"

type BookingConfirmation struct {
	BookingReq BookingRequest `json:"booking_request"`
	DriverID   string         `json:"driver_id"`
	DriverName string         `json:"driver_name"`
}

type Booking struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	DriverID    int32     `json:"driver_id"`
	DriverName  string    `json:"driver_name,omitempty"`
	Price       float64   `json:"price"`
	Pickup      GeoPoint  `json:"pickup"`
	Dropoff     GeoPoint  `json:"dropoff"`
	BookedAt    time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
	Status      string    `json:"status"`
}

type BookingRequest struct {
	UserID      string    `json:"user_id" bson:"user_id"`
	UserName    string    `json:"user_name" bson:"user_name"`
	Pickup      GeoPoint  `json:"pickup" bson:"pickup"`
	Dropoff     GeoPoint  `json:"dropoff" bson:"dropoff"`
	VehicleType string    `json:"vehicle_type" bson:"vehicle_type"`
	Price       float64   `json:"price" bson:"price"`
	MongoID     string    `json:"mongo_id,omitempty" bson:"mongo_id,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}
