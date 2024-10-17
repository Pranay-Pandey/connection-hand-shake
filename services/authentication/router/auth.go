package router

import (
	"logistics-platform/lib/middlewares/auth"
	"logistics-platform/services/authentication/interfaces"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, service interfaces.AuthService) {
	router.GET("/health", service.HealthCheck)

	user := router.Group("/user")
	{
		user.POST("/register", service.RegisterUser)
		user.POST("/login", service.Login)
		user.Use(auth.AuthInjectionMiddleware())
		{
			user.GET("/profile", service.GetProfile)
		}
	}

	driver := router.Group("/driver")
	{
		driver.POST("/register", service.RegisterDriver)
		driver.POST("/login", service.LoginDriver)
		driver.Use(auth.AuthInjectionMiddleware())
		{
			driver.GET("/profile/:id", service.GetDriverProfile)
		}
	}

	admin := router.Group("/admin")
	{
		admin.POST("/login", service.AdminLogin)
	}
}
