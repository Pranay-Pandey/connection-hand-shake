package main

import (
	"log"
	"logistics-platform/lib/config"
	"logistics-platform/lib/database"
	"logistics-platform/services/driver_location/service"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisClient, err := database.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	server := service.NewDriverLocationService(redisClient)

	go server.ConsumeDriverLocations()
	go server.ConsumeBookingNotifications()

	server.GracefulShutdown()
}
