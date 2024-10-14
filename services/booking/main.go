// services/booking/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"logistics-platform/lib/token"
	"net/http"
	"strings"

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
	UserName    string         `json:"user_name" bson:"user_name"`
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

	// authenticate user
	headerAuthToken := c.GetHeader("Authorization")
	if headerAuthToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
		return
	}

	authToken := strings.TrimPrefix(headerAuthToken, "Bearer ")
	if authToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	// validate token
	user, err := token.GetUserFromToken(authToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	bookingReq.UserID = user.UserID
	bookingReq.UserName = user.UserName

	go s.processBookingRequest(bookingReq)

	c.JSON(http.StatusOK, gin.H{"message": "Booking request received"})
}

func (s *BookingService) processBookingRequest(bookingReq BookingRequest) error {
	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	_, err := collection.InsertOne(context.Background(), bookingReq)
	if err != nil {
		return fmt.Errorf("error storing booking request: %w", err)
	}

	nearbyDrivers, err := s.findNearbyDrivers(bookingReq, bookingReq.VehicleType)
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

func (s *BookingService) findNearbyDrivers(bookingReq BookingRequest, vehicleType string) ([]string, error) {
	pickup := bookingReq.Pickup
	drivers, err := s.redisClient.GeoRadius(context.Background(), "driver_locations", pickup.Longitude, pickup.Latitude, &redis.GeoRadiusQuery{
		Radius: 100,
		Unit:   "km",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("error finding nearby drivers: %w", err)
	}

	var nearbyDrivers []string
	for _, driver := range drivers {
		go func(driverID string) {
			if err := s.notifyDriver(driverID, bookingReq); err != nil {
				log.Printf("Error notifying driver %s: %v", driverID, err)
			}
		}(driver.Name)
	}
	return nearbyDrivers, nil
}

func (s *BookingService) notifyDriver(driverID string, bookingReq BookingRequest) error {
	notification := utils.BookingNotification{
		UserID:   bookingReq.UserID,
		Price:    bookingReq.Price,
		DriverID: driverID,
		Pickup:   bookingReq.Pickup,
		Dropoff:  bookingReq.Dropoff,
		UserName: bookingReq.UserName,
	}
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	return s.notificationWriter.WriteMessages(context.Background(), kafka.Message{Value: notificationJSON})
}
