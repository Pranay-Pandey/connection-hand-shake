package models

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
