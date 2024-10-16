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
	driverConnections      sync.Map
	userConnections        sync.Map
	driverUserConnections  sync.Map
	locationWriter         *kafka.Writer
	notificationReader     *kafka.Reader
	bookNotificationReader *kafka.Reader
}

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	service := &NotificationService{
		locationWriter:         utils.InitKafkaWriter("driver_locations"),
		notificationReader:     utils.InitKafkaReader("driver_notification", "driver_notification"),
		bookNotificationReader: utils.InitKafkaReader("booking_notifications", "booking_notification"),
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.GET("/driver/ws", service.handleDriverWebSocket)
	router.GET("/user/ws", service.handleUserWebSocket)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	go service.consumeNotifications()
	go service.consumeBookingNotifications()

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
	defer func() {
		s.driverConnections.Delete(driverID)
		// send empty location to kafka to remove driver from cache
		s.sendLocationUpdate(utils.DriverLocation{DriverID: driverID})
	}()

	// Proceed with WebSocket communication
	for {
		var location utils.DriverLocation
		if err := conn.ReadJSON(&location); err != nil {
			log.Printf("Error reading location JSON: %v", err)
			break
		}
		location.DriverID = driverID

		// Send the location update
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

	// if there is any driver user connection, send the notification to the user
	userID, ok := s.driverUserConnections.Load(location.DriverID)
	if ok {
		log.Print("Sending location update to user")
		conn, ok := s.userConnections.Load(userID.(string))
		if ok {
			err = conn.(*websocket.Conn).WriteJSON(location)
		} else {
			log.Print("User connection not found")
		}
	} else {
		log.Print("sending location to Kafka")
		err = s.locationWriter.WriteMessages(context.Background(),
			kafka.Message{Value: message})
	}

	return err
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

func (s *NotificationService) handleUserWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	defer conn.Close()

	// Expect an initial message with the authentication token
	var authMessage struct {
		Token string `json:"token"`
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
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "Invalid authentication token"))
		return
	}

	userID := user.UserID
	if userID == "" {
		log.Println("User ID is required")
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "User ID required"))
		return
	}

	// Store the authenticated connection
	s.userConnections.Store(userID, conn)
	defer s.userConnections.Delete(userID)

	// Proceed with WebSocket communication
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
	}
}

func (s *NotificationService) consumeBookingNotifications() {
	for {
		msg, err := s.bookNotificationReader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var notification utils.BookedNotification
		if err := json.Unmarshal(msg.Value, &notification); err != nil {
			log.Printf("Error unmarshaling notification: %v", err)
			continue
		}

		if notification.Status == "booked" {
			s.driverUserConnections.Store(notification.DriverID, notification.UserID)
		} else if notification.Status == "completed" {
			s.driverUserConnections.Delete(notification.DriverID)
		}
	}
}
