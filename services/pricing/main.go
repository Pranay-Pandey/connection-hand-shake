package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"logistics-platform/lib/utils"

	"github.com/redis/go-redis/v9"

	"logistics-platform/lib/middlewares/cors"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gin-gonic/gin"
)

type PricingService struct {
	db          *pgxpool.Pool
	redisClient *redis.Client
}

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

var vehiclePricingData = map[string]VehiclePricing{
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

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisClient, err := utils.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(utils.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to parse pool config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	service := &PricingService{
		db:          pool,
		redisClient: redisClient,
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.POST("/pricing/estimate", service.handlePriceEstimate)

	server := &http.Server{
		Addr:    ":8086",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	utils.WaitForShutdown(server, redisClient)
}

func (s *PricingService) handlePriceEstimate(c *gin.Context) {
	var req utils.BookingRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	priceEstimate, err := s.EstimatePrice(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, struct {
		PriceEstimate float64 `json:"price"`
	}{
		PriceEstimate: priceEstimate.TotalPrice,
	})
}

func (s *PricingService) EstimatePrice(ctx context.Context, req utils.BookingRequest) (PriceEstimate, error) {
	// Step 1: Calculate distance and estimated duration
	distance := calculateDistance(req.Pickup, req.Dropoff)
	duration := estimateDuration(distance)

	// Step 2: Get vehicle pricing
	vehiclePricing, err := s.getVehiclePricing(req.VehicleType)
	if err != nil {
		return PriceEstimate{}, err
	}

	// Step 3: Calculate base price
	basePrice := vehiclePricing.BasePrice +
		(distance * vehiclePricing.PricePerKm) +
		(duration * vehiclePricing.PricePerMinute)

	// Step 4: Apply surge pricing
	surgeMultiplier := s.calculateSurgeMultiplier(ctx, req.Pickup, req.Dropoff)

	// Step 5: Calculate total price
	totalPrice := basePrice * surgeMultiplier

	return PriceEstimate{
		BasePrice:  basePrice,
		Distance:   distance,
		Duration:   duration,
		Surge:      surgeMultiplier,
		TotalPrice: totalPrice,
	}, nil
}

func calculateDistance(pickup, dropoff utils.GeoPoint) float64 {
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

func (s *PricingService) getVehiclePricing(vehicleType string) (VehiclePricing, error) {
	return vehiclePricingData[vehicleType], nil
}

func (s *PricingService) calculateSurgeMultiplier(ctx context.Context, pickup, dropoff utils.GeoPoint) float64 {
	// 1. Check current demand in the area
	demand, err := s.getCurrentDemand(pickup)
	if err != nil {
		log.Printf("Failed to get current demand: %v", err)
		return 1.0
	}

	// 2. Check time of day
	hour := time.Now().Hour()

	// 3. Apply surge based on demand and time
	surgeFactor := 1.0
	if demand > 0.8 { // High demand
		surgeFactor *= 1.5
	} else if demand > 0.6 { // Moderate demand
		surgeFactor *= 1.2
	}

	// Increase surge during peak hours (e.g., 7-9 AM and 5-7 PM)
	if (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19) {
		surgeFactor *= 1.2
	}

	return surgeFactor
}

func (s *PricingService) getCurrentDemand(location utils.GeoPoint) (float64, error) {
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
