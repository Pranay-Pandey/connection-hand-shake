package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"logistics-platform/lib/config"
	"logistics-platform/lib/database"
	"logistics-platform/services/pricing/router"
	"logistics-platform/services/pricing/service"

	"logistics-platform/lib/middlewares/cors"

	"github.com/jackc/pgx/v4/pgxpool"

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

	service := service.NewPricingService(pool, redisClient)

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
