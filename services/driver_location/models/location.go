package models

import (
	"encoding/json"
	"time"
)

type Location struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

func (l Location) ToJSON() string {
	data, _ := json.Marshal(l)
	return string(data)
}
