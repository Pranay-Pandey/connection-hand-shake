// services/driver_location/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"logistics-platform/lib/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type DriverLocationService struct {
	redisClient *redis.Client
	kafkaReader *kafka.Reader
}

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisClient, err := utils.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	service := &DriverLocationService{
		redisClient: redisClient,
		kafkaReader: utils.InitKafkaReader("driver_locations", "driver-location-service"),
	}

	go service.consumeDriverLocations()

	utils.WaitForShutdown(redisClient, service.kafkaReader)
}

func (s *DriverLocationService) consumeDriverLocations() {
	for {
		msg, err := s.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message from Kafka: %v", err)
			continue
		}

		var location utils.DriverLocation
		if err := json.Unmarshal(msg.Value, &location); err != nil {
			log.Printf("Error unmarshaling location update: %v", err)
			continue
		}

		if err := s.updateDriverLocation(location); err != nil {
			log.Printf("Error updating driver location: %v", err)
		}
	}
}

func (s *DriverLocationService) updateDriverLocation(location utils.DriverLocation) error {
	err := s.redisClient.GeoAdd(context.Background(), "driver_locations", &redis.GeoLocation{
		Name:      location.DriverID,
		Longitude: location.Location.Longitude,
		Latitude:  location.Location.Latitude,
	}).Err()

	if err != nil {
		return err
	}

	// Set expiration time for the driver's location to 5 minutes
	expiration := 5 * time.Minute
	_, err = s.redisClient.Expire(context.Background(), "driver_locations:"+location.DriverID, expiration).Result()

	return err
}
