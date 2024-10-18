package main

import (
	"log"
	"net/http"

	"logistics-platform/lib/config"
	"logistics-platform/lib/database"
	"logistics-platform/services/pricing/router"
	"logistics-platform/services/pricing/service"

	"logistics-platform/lib/middlewares/cors"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisClient, err := database.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	service := service.NewPricingService(redisClient)

	r := gin.Default()
	r.Use(cors.CORSMiddleware())
	router.SetupRouter(r, service)

	server := &http.Server{
		Addr:    ":8086",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	service.GracefulShutdown(server)
}
