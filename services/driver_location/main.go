// services/driver_location/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

var (
	redisClient *redis.Client
	kafkaReader *kafka.Reader
)

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
	initKafka()

	// Start consuming driver location updates in a Goroutine
	go consumeDriverLocations()

	// Graceful shutdown
	waitForShutdown()
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	// Close connections
	redisClient.Close()
	kafkaReader.Close()

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

func initKafka() {
	kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{viper.GetString("KAFKA_ADDR")},
		Topic:   "driver_locations",
		GroupID: "driver-location-service",
	})
}

func consumeDriverLocations() {
	for {
		msg, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka: %v", err)
			continue
		}

		var location DriverLocation
		err = json.Unmarshal(msg.Value, &location)
		if err != nil {
			log.Printf("Error unmarshaling location update: %v", err)
			continue
		}

		storeDriverLocation(location)
	}
}

func storeDriverLocation(location DriverLocation) {
	// Store driver location in Redis
	ctx := context.Background()
	err := redisClient.Set(ctx, location.DriverID, location, 10*time.Second).Err()
	if err != nil {
		log.Printf("Error storing driver location in Redis: %v", err)
	}
}
