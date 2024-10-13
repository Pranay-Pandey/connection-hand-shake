// services/booking/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/lib/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingService struct {
	mongoClient        *mongo.Client
	notificationWriter *kafka.Writer
	redisClient        *redis.Client
}

type BookingRequest struct {
	UserID      string         `json:"user_id" bson:"user_id"`
	Pickup      utils.GeoPoint `json:"pickup" bson:"pickup"`
	Dropoff     utils.GeoPoint `json:"dropoff" bson:"dropoff"`
	VehicleType string         `json:"vehicle_type" bson:"vehicle_type"`
	Price       float64        `json:"price" bson:"price"`
}

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mongoClient, err := utils.InitMongoDB()
	defer mongoClient.Disconnect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	redisClient, err := utils.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	service := &BookingService{
		mongoClient:        mongoClient,
		notificationWriter: utils.InitKafkaWriter("driver_notification"),
		redisClient:        redisClient,
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.POST("/booking", service.handleBookingRequest)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	server := &http.Server{
		Addr:    ":8084",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	utils.WaitForShutdown(server, service.notificationWriter, redisClient)
}

func (s *BookingService) handleBookingRequest(c *gin.Context) {
	var bookingReq BookingRequest
	if err := c.ShouldBindJSON(&bookingReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.processBookingRequest(c.Request.Context(), bookingReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process booking request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking request received"})
}

func (s *BookingService) processBookingRequest(ctx context.Context, bookingReq BookingRequest) error {
	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	_, err := collection.InsertOne(ctx, bookingReq)
	if err != nil {
		return fmt.Errorf("error storing booking request: %w", err)
	}

	nearbyDrivers, err := s.findNearbyDrivers(ctx, bookingReq.Pickup, bookingReq.VehicleType)
	if err != nil {
		return fmt.Errorf("error finding nearby drivers: %w", err)
	}

	for _, driverID := range nearbyDrivers {
		go func(driverID string) {
			if err := s.notifyDriver(driverID, bookingReq); err != nil {
				log.Printf("Error notifying driver %s: %v", driverID, err)
			}
		}(driverID)
	}

	return nil
}

func (s *BookingService) findNearbyDrivers(ctx context.Context, pickup utils.GeoPoint, vehicleType string) ([]string, error) {
	drivers, err := s.redisClient.GeoRadius(ctx, "driver_locations", pickup.Longitude, pickup.Latitude, &redis.GeoRadiusQuery{
		Radius: 100,
		Unit:   "km",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("error finding nearby drivers: %w", err)
	}

	var nearbyDrivers []string
	for _, driver := range drivers {
		nearbyDrivers = append(nearbyDrivers, driver.Name)
	}
	return nearbyDrivers, nil
}

func (s *BookingService) notifyDriver(driverID string, bookingReq BookingRequest) error {
	notification := utils.BookingNotification{
		UserID:   bookingReq.UserID,
		Price:    bookingReq.Price,
		DriverID: driverID,
	}
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	return s.notificationWriter.WriteMessages(context.Background(), kafka.Message{Value: notificationJSON})
}
