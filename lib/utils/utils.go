package utils

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

type GeoPoint struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Name      string  `json:"name" bson:"name"`
}

type DriverLocation struct {
	DriverID  string    `json:"driver_id" bson:"driver_id"`
	Location  GeoPoint  `json:"location" bson:"location"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type BookingNotification struct {
	UserID   string   `json:"user_id" bson:"user_id"`
	DriverID string   `json:"driver_id" bson:"driver_id"`
	Price    float64  `json:"price" bson:"price"`
	Pickup   GeoPoint `json:"pickup" bson:"pickup"`
	Dropoff  GeoPoint `json:"dropoff" bson:"dropoff"`
	UserName string   `json:"user_name" bson:"user_name"`
	MongoID  string   `json:"mongo_id" bson:"mongo_id"`
}

type BookedNotification struct {
	UserID     string `json:"user_id"`
	DriverID   string `json:"driver_id"`
	DriverName string `json:"driver_name"`
	Status     string `json:"status"`
}

type UserRequest struct {
	UserID   string `json:"user_id" bson:"user_id"`
	UserName string `json:"user_name" bson:"user_name"`
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

func WaitForShutdown(closers ...interface{ Close() error }) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			// Consider using a proper logging package here
			println("Error closing:", err.Error())
		}
	}
}
