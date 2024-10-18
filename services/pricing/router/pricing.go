package router

import (
	"logistics-platform/services/pricing/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, service interfaces.PricingInterface) {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	router.POST("/pricing/estimate", service.HandlePriceEstimate)

}
