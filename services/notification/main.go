// services/notification/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/lib/token"
	"net/http"
	"sync"

	"logistics-platform/lib/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
)

type NotificationService struct {
	driverConnections  sync.Map
	locationWriter     *kafka.Writer
	notificationReader *kafka.Reader
}

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	service := &NotificationService{
		locationWriter:     utils.InitKafkaWriter("driver_locations"),
		notificationReader: utils.InitKafkaReader("driver_notification", "test-consumer-group"),
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.GET("/driver/ws", service.handleDriverWebSocket)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	go service.consumeNotifications()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	utils.WaitForShutdown(server, service.locationWriter, service.notificationReader)
}

func (s *NotificationService) handleDriverWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Expect an initial message with the authentication token
	var authMessage struct {
		Token    string `json:"token"`
		DriverID string `json:"driver_id"`
	}

	if err := conn.ReadJSON(&authMessage); err != nil {
		log.Printf("Failed to read auth message: %v", err)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Authentication required"))
		return
	}

	// Validate the token
	user, err := token.GetUserFromToken(authMessage.Token) // Reuse token validation logic
	if err != nil {
		log.Printf("Invalid token: %v", err.Error())
		log.Printf("Token user ID: %s, Driver ID: %s", user.UserID, authMessage.DriverID)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Invalid authentication token"))
		return
	}

	driverID := user.UserID
	if driverID == "" {
		log.Println("Driver ID is required")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Driver ID required"))
		return
	}

	// Store the authenticated connection
	s.driverConnections.Store(driverID, conn)
	defer s.driverConnections.Delete(driverID)

	// Proceed with WebSocket communication
	for {
		var location utils.DriverLocation
		if err := conn.ReadJSON(&location); err != nil {
			log.Printf("Error reading location JSON: %v", err)
			break
		}
		location.DriverID = driverID

		// Send the location update to Kafka
		if err := s.sendLocationUpdate(location); err != nil {
			log.Printf("Error sending location update: %v", err)
		}
	}
}

func (s *NotificationService) sendLocationUpdate(location utils.DriverLocation) error {
	message, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	return s.locationWriter.WriteMessages(context.Background(),
		kafka.Message{Value: message})
}

func (s *NotificationService) consumeNotifications() {
	for {
		msg, err := s.notificationReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var notification utils.BookingNotification
		if err := json.Unmarshal(msg.Value, &notification); err != nil {
			log.Printf("Error unmarshaling notification: %v", err)
			continue
		}

		if err := s.sendNotification(notification); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}
}

func (s *NotificationService) sendNotification(notification utils.BookingNotification) error {
	conn, ok := s.driverConnections.Load(notification.DriverID)
	if !ok {
		return fmt.Errorf("driver %s not connected", notification.DriverID)
	}

	return conn.(*websocket.Conn).WriteJSON(notification)
}
