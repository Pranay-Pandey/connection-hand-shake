package main

import (
	"log"
	"logistics-platform/lib/config"
	"logistics-platform/lib/database"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/services/authentication/router"
	"logistics-platform/services/authentication/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the database connection
	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(nil)

	// Create a new instance of the auth service
	authService := service.NewAuthService(db)

	// Setup the router
	r := gin.Default()
	r.Use(cors.CORSMiddleware())
	router.SetupRouter(r, authService)

	// Run the service
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
