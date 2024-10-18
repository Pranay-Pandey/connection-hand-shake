package service

import (
	"context"
	"log"
	"logistics-platform/lib/models"
	"logistics-platform/lib/utils"
	"logistics-platform/services/pricing/interfaces"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PricingService struct {
	db          *pgxpool.Pool
	redisClient *redis.Client
}

var vehiclePricingData = map[string]models.VehiclePricing{
	"light_truck": {
		Type:           "light_truck",
		BasePrice:      20.0,
		PricePerKm:     0.1117,
		PricePerMinute: 0.34,
	},
	"van": {
		Type:           "van",
		BasePrice:      20.0,
		PricePerKm:     0.1791,
		PricePerMinute: 0.34,
	},
	"truck": {
		Type:           "truck",
		BasePrice:      50.0,
		PricePerKm:     0.2924,
		PricePerMinute: 0.5,
	},
	"heavy_truck": {
		Type:           "heavy_truck",
		BasePrice:      100.0,
		PricePerKm:     0.3488,
		PricePerMinute: 0.6,
	},
	"trailer": {
		Type:           "trailer",
		BasePrice:      200.0,
		PricePerKm:     0.7859,
		PricePerMinute: 0.8,
	},
}

func NewPricingService(db *pgxpool.Pool, redisClient *redis.Client) interfaces.PricingInterface {
	return &PricingService{
		db:          db,
		redisClient: redisClient,
	}
}

func (s *PricingService) HandlePriceEstimate(c *gin.Context) {
	var req models.BookingRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	PriceEstimate, err := s.EstimatePrice(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, struct {
		PriceEstimate float64 `json:"price"`
	}{
		PriceEstimate: PriceEstimate.TotalPrice,
	})
}

func (s *PricingService) EstimatePrice(ctx context.Context, req models.BookingRequest) (models.PriceEstimate, error) {
	distance := calculateDistance(req.Pickup, req.Dropoff)
	duration := estimateDuration(distance)

	vehiclePricing, err := s.GetVehiclePricing(req.VehicleType)
	if err != nil {
		return models.PriceEstimate{}, err
	}

	basePrice := vehiclePricing.BasePrice +
		(distance * vehiclePricing.PricePerKm) +
		(duration * vehiclePricing.PricePerMinute)

	surgeMultiplier := s.CalculateSurgeMultiplier(ctx, req.Pickup, req.Dropoff)

	totalPrice := basePrice * surgeMultiplier

	return models.PriceEstimate{
		BasePrice:  basePrice,
		Distance:   distance,
		Duration:   duration,
		Surge:      surgeMultiplier,
		TotalPrice: totalPrice,
	}, nil
}

func calculateDistance(pickup, dropoff models.GeoPoint) float64 {
	// Haversine formula for calculating distance between two points on a sphere
	const earthRadius = 6371 // km

	lat1 := toRadians(pickup.Latitude)
	lon1 := toRadians(pickup.Longitude)
	lat2 := toRadians(dropoff.Latitude)
	lon2 := toRadians(dropoff.Longitude)

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func estimateDuration(distance float64) float64 {
	// Assuming an average speed of 40 km/h
	averageSpeed := 40.0                // km/h
	return distance / averageSpeed * 60 // Convert to minutes
}

func (s *PricingService) GetVehiclePricing(vehicleType string) (models.VehiclePricing, error) {
	return vehiclePricingData[vehicleType], nil
}

func (s *PricingService) CalculateSurgeMultiplier(ctx context.Context, pickup, dropoff models.GeoPoint) float64 {
	demand, err := s.GetCurrentDemand(pickup)
	if err != nil {
		log.Printf("Failed to get current demand: %v", err)
		return 1.0
	}

	hour := time.Now().Hour()

	surgeFactor := 1.0
	if demand > 0.8 { // High demand
		surgeFactor *= 1.5
	} else if demand > 0.6 { // Moderate demand
		surgeFactor *= 1.2
	}

	// Increase surge during peak hours (7-9 AM and 5-7 PM)
	if (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19) {
		surgeFactor *= 1.2
	}

	return surgeFactor
}

func (s *PricingService) GetCurrentDemand(location models.GeoPoint) (float64, error) {
	// Get nearby drivers
	nearby, err := s.redisClient.GeoRadius(context.Background(), "driver_locations", location.Longitude, location.Latitude, &redis.GeoRadiusQuery{
		Radius: 1000,
		Unit:   "km",
	}).Result()
	if err != nil {
		return 0, err
	}

	demand := 1.0 - math.Min(1.0, float64(len(nearby))/100.0)
	return demand, nil
}

func (s *PricingService) GracefulShutdown(server *http.Server) {
	utils.WaitForShutdown(server, s.redisClient)
}
