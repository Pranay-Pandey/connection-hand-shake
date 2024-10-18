package models

type PriceEstimate struct {
	BasePrice  float64
	Distance   float64
	Duration   float64
	Surge      float64
	TotalPrice float64
}

type VehiclePricing struct {
	Type           string
	BasePrice      float64
	PricePerKm     float64
	PricePerMinute float64
}
