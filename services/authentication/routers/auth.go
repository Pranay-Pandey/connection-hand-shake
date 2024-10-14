package routers

import (
	"logistics-platform/services/authentication/handler"

	"logistics-platform/lib/middlewares/auth"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.Engine) {
	user := router.Group("/user")
	{
		user.POST("/register", handler.RegisterUser)
		user.GET("/health", handler.HealthCheck)
		user.POST("/login", handler.Login)
		user.Use(auth.AuthInjectionMiddleware())
		{
			user.GET("/profile", handler.GetProfile)
		}
	}
}

func RegisterDriverRoutes(router *gin.Engine) {
	driver := router.Group("/driver")
	{
		driver.POST("/register", handler.RegisterDriver)
		driver.POST("/login", handler.LoginDriver)
		driver.Use(auth.AuthInjectionMiddleware())
		{
			driver.GET("/profile/:id", handler.GetDriverProfile)
		}
	}
}
