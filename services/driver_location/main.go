// services/driver_location/main.go
package main

import (
	"context"
	"encoding/json"
	"log"

	"logistics-platform/lib/utils"

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

		log.Printf("Received message: %s", string(msg.Value))

		var location utils.DriverLocation
		if err := json.Unmarshal(msg.Value, &location); err != nil {
			log.Printf("Error unmarshaling location update: %v", err)
			continue
		}

		log.Printf("Updating driver location: %v", location)

		if err := s.updateDriverLocation(location); err != nil {
			log.Printf("Error updating driver location: %v", err)
		}
		log.Printf("Driver location updated successfully")
	}
}

// set the driver location in Redis so that it can be used by the booking service later to find nearby drivers
func (s *DriverLocationService) updateDriverLocation(location utils.DriverLocation) error {
	err := s.redisClient.GeoAdd(context.Background(), "driver_locations", &redis.GeoLocation{
		Name:      location.DriverID,
		Longitude: location.Location.Longitude,
		Latitude:  location.Location.Latitude,
	}).Err()

	return err
}
