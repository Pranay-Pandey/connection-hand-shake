package models

type GeoPoint struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Name      string  `json:"name" bson:"name"`
}

type UserRequest struct {
	UserID   string `json:"user_id" bson:"user_id"`
	UserName string `json:"user_name" bson:"user_name"`
}
