package routers

import (
	"logistics-platform/services/authentication/handler"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.Engine) {
	user := router.Group("/user")
	{
		user.POST("/register", handler.RegisterUser)
		user.GET("/health", handler.HealthCheck)
		user.POST("/login", handler.Login)
		user.GET("/profile", handler.GetProfile)
	}
}

func RegisterDriverRoutes(router *gin.Engine) {
	driver := router.Group("/driver")
	{
		driver.POST("/register", handler.RegisterDriver)
		driver.GET("/profile/:id", handler.GetDriverProfile)
		driver.POST("/login", handler.LoginDriver)
	}
}
