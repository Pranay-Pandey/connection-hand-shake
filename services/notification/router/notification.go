package router

import (
	"logistics-platform/services/notification/interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine, service interfaces.NotificationInterface) {

	router.GET("/driver/ws", service.HandleDriverWebSocket)
	router.GET("/user/ws", service.HandleUserWebSocket)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

}
