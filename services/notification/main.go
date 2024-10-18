package main

import (
	"log"
	"logistics-platform/lib/config"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/services/notification/router"
	"logistics-platform/services/notification/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	service := service.NewNotificationService()

	r := gin.Default()
	r.Use(cors.CORSMiddleware())
	router.SetupRouter(r, service)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go service.ConsumeNotifications()
	go service.ConsumeBookingNotifications()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	service.GracefulShutdown(server)
}
