// services/booking/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"logistics-platform/lib/config"
	"logistics-platform/lib/database"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/services/booking/router"
	"logistics-platform/services/booking/service"

	"github.com/gin-gonic/gin"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mongoClient, err := database.InitMongoDB()
	defer mongoClient.Disconnect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	redisClient, err := database.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.GetDBConnectionString())
	if err != nil {
		log.Fatalf("Failed to parse pool config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pool.Close()

	service := service.NewBookingService(mongoClient, redisClient, pool)

	r := gin.Default()
	r.Use(cors.CORSMiddleware())
	router.SetupRouter(r, service)

	server := &http.Server{
		Addr:    ":8084",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	service.GracefulShutdown(server)
}
