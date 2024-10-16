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
	redisClient            *redis.Client
	kafkaReader            *kafka.Reader
	bookNotificationReader *kafka.Reader
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
		redisClient:            redisClient,
		kafkaReader:            utils.InitKafkaReader("driver_locations", "driver-location-service"),
		bookNotificationReader: utils.InitKafkaReader("booking_notifications", "booking_notification"),
	}

	go service.consumeDriverLocations()
	go service.consumeBookingNotifications()

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

		if location.Location.Latitude == 0 && location.Location.Longitude == 0 {
			if err := s.removeDriverFromCache(location.DriverID); err != nil {
				log.Printf("Error removing driver from cache: %v", err)
			}
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

func (s *DriverLocationService) removeDriverFromCache(driverID string) error {
	_, err := s.redisClient.ZRem(context.Background(), "driver_locations", driverID).Result()
	return err
}

func (s *DriverLocationService) consumeBookingNotifications() {
	for {
		msg, err := s.bookNotificationReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var notification utils.BookedNotification
		if err := json.Unmarshal(msg.Value, &notification); err != nil {
			log.Printf("Error unmarshaling notification: %v", err)
			continue
		}

		if notification.Status == "booked" {
			if err := s.removeDriverFromCache(notification.DriverID); err != nil {
				log.Printf("Error removing driver from cache: %v", err)
			}
		}
	}
}
