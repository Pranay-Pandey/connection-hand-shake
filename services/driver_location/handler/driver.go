package handlers

import (
	"logistics-platform/lib/token"
	"logistics-platform/services/driver_location/models"
	"logistics-platform/services/driver_location/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ConnectDriverWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket upgrade failed"})
		return
	}
	defer conn.Close()

	// Extract and verify JWT
	tokenStr := c.Query("token")
	driverID, err := token.GetSubjectFromToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	go services.HandleWebSocketConnection(conn, driverID)
}

func UpdateLocation(c *gin.Context) {
	var location models.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Extract and verify JWT
	tokenStr := c.GetHeader("Authorization")[7:]
	driverID, err := token.GetSubjectFromToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Store in Redis and publish to Kafka
	go services.StoreAndPublishLocation(driverID, location)

	c.JSON(http.StatusOK, gin.H{"message": "Location updated"})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Service is healthy"})
}
