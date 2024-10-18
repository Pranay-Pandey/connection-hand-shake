package models

import "time"

type DriverLocation struct {
	DriverID  string    `json:"driver_id" bson:"driver_id"`
	Location  GeoPoint  `json:"location" bson:"location"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}
