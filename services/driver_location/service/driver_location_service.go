package service

import (
	"context"
	"encoding/json"
	"log"
	"logistics-platform/lib/utils"
	"logistics-platform/services/driver_location/interfaces"

	kafkaConfig "logistics-platform/lib/kafka"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type DriverLocationService struct {
	redisClient            *redis.Client
	kafkaReader            *kafka.Reader
	bookNotificationReader *kafka.Reader
}

func NewDriverLocationService(redisClient *redis.Client) interfaces.DriverLocationInterface {
	return &DriverLocationService{
		redisClient:            redisClient,
		kafkaReader:            kafkaConfig.InitKafkaReader("driver_locations", "driver-location-service"),
		bookNotificationReader: kafkaConfig.InitKafkaReader("booking_notifications", "driver_location_service_group"),
	}
}

func (s *DriverLocationService) ConsumeDriverLocations() {
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
			if err := s.RemoveDriverFromCache(location.DriverID); err != nil {
				log.Printf("Error removing driver from cache: %v", err)
			}
			continue
		}

		if err := s.UpdateDriverLocation(location); err != nil {
			log.Printf("Error updating driver location: %v", err)
		}
	}
}

func (s *DriverLocationService) UpdateDriverLocation(location utils.DriverLocation) error {
	err := s.redisClient.GeoAdd(context.Background(), "driver_locations", &redis.GeoLocation{
		Name:      location.DriverID,
		Longitude: location.Location.Longitude,
		Latitude:  location.Location.Latitude,
	}).Err()
	return err
}

func (s *DriverLocationService) RemoveDriverFromCache(driverID string) error {
	_, err := s.redisClient.ZRem(context.Background(), "driver_locations", driverID).Result()
	return err
}

func (s *DriverLocationService) ConsumeBookingNotifications() {
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
			if err := s.RemoveDriverFromCache(notification.DriverID); err != nil {
				log.Printf("Error removing driver from cache: %v", err)
			}
		}
	}
}

func (s *DriverLocationService) GracefulShutdown() {
	utils.WaitForShutdown(s.redisClient, s.kafkaReader)
}
