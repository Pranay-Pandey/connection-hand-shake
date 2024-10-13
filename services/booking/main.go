// services/booking/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	redisClient        *redis.Client
	mongoClient        *mongo.Client
	NotificationWriter *kafka.Writer
)

type BookingRequest struct {
	UserID      string   `json:"user_id" bson:"user_id"`
	Pickup      GeoPoint `json:"pickup" bson:"pickup"`
	Dropoff     GeoPoint `json:"dropoff" bson:"dropoff"`
	VehicleType string   `json:"vehicle_type" bson:"vehicle_type"`
	Price       float64  `json:"price" bson:"price"`
}

type GeoPoint struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type DriverLocation struct {
	DriverID  string    `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Load configuration
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	// Initialize connections
	initRedis()
	initMongoDB()
	initKafka()

	// Set up Gin router for HTTP endpoints
	router := gin.Default()
	router.POST("/booking", handleBookingRequest)

	// Start the HTTP server
	go func() {
		if err := router.Run(":8084"); err != nil {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	// Close connections
	redisClient.Close()
	mongoClient.Disconnect(context.Background())
	NotificationWriter.Close()

	log.Println("Server exiting")
}

func initRedis() {
	redisURL := viper.GetString("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Error parsing Redis URL: ", err)
	}
	redisClient = redis.NewClient(opt)
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Error connecting to Redis: ", err)
	}
}

func initMongoDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(viper.GetString("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}
	mongoClient = client
}

func initKafka() {
	NotificationWriter = &kafka.Writer{
		Addr:     kafka.TCP(viper.GetString("KAFKA_ADDR")),
		Topic:    "driver_notification",
		Balancer: &kafka.LeastBytes{},
	}
}

func handleBookingRequest(c *gin.Context) {
	var bookingReq BookingRequest
	if err := c.ShouldBindJSON(&bookingReq); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	go processBookingRequest(bookingReq)

	c.JSON(200, gin.H{"message": "Booking request received"})
}

func processBookingRequest(bookingReq BookingRequest) {
	// Store booking request in MongoDB
	collection := mongoClient.Database("logistics").Collection("booking_requests")
	_, err := collection.InsertOne(context.Background(), bookingReq)
	if err != nil {
		log.Printf("Error storing booking request: %v", err)
		return
	}

	// Find nearby drivers
	nearbyDrivers := findNearbyDrivers(bookingReq.Pickup, bookingReq.VehicleType)

	// Notify nearby drivers
	for _, driver := range nearbyDrivers {
		notifyDriver(driver, bookingReq)
	}
}

func findNearbyDrivers(pickup GeoPoint, vehicleType string) []string {
	// This is a simplified version. In a real-world scenario, you'd use more sophisticated
	// geospatial queries and filtering based on vehicle type.
	ctx := context.Background()
	drivers, err := redisClient.GeoRadius(ctx, "driver_locations", pickup.Longitude, pickup.Latitude, &redis.GeoRadiusQuery{
		Radius: 50, // 5 km radius
		Unit:   "km",
	}).Result()

	if err != nil {
		log.Printf("Error finding nearby drivers: %v", err)
		return nil
	}

	var nearbyDrivers []string
	for _, driver := range drivers {
		nearbyDrivers = append(nearbyDrivers, driver.Name)
	}

	return nearbyDrivers
}

func notifyDriver(driverID string, bookingReq BookingRequest) {
	notification := map[string]interface{}{
		"type":     "new_booking",
		"driverID": driverID,
		"booking":  bookingReq,
	}
	notificationJSON, _ := json.Marshal(notification)

	err := NotificationWriter.WriteMessages(context.Background(), kafka.Message{
		Value: notificationJSON,
	})
	if err != nil {
		log.Printf("Error sending notification to driver %s: %v", driverID, err)
	}

}
