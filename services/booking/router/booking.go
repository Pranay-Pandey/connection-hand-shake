package router

import (
	"logistics-platform/lib/middlewares/auth"
	"logistics-platform/services/booking/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, service interfaces.BookingInterface) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	userGroup := router.Group("/user")
	userGroup.Use(auth.AuthInjectionMiddleware())
	{
		userGroup.GET("/booking", service.HandleUserBookingCheck)
		userGroup.GET("/booking-history", service.HandleUserBookingHistory)
	}

	driverGroup := router.Group("/driver")
	driverGroup.Use(auth.AuthInjectionMiddleware())
	{
		driverGroup.GET("/booking", service.HandleDriverBookingCheck)
		driverGroup.GET("/booking-history", service.HandleDriverBookingHistory)
	}

	router.Use(auth.AuthInjectionMiddleware())
	{
		router.POST("/booking/accept", service.HandleBookingAccept)
		router.POST("/booking", service.HandleBookingRequest)
		router.PATCH("/booking/:userId", service.HandleBookingUpdate)
	}
}
