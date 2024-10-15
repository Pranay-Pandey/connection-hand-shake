package main

import (
	"logistics-platform/lib/database"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/services/authentication/routers"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	router := gin.Default()
	router.Use(cors.CORSMiddleware())

	// Load environment variables
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	// Initialize the database connection
	database.InitPostgres()

	// Register user routes
	routers.RegisterUserRoutes(router)
	routers.RegisterDriverRoutes(router)
	routers.RegisterAdminRoutes(router)

	// Run the user service on a specific port
	router.Run(":8081")
}
