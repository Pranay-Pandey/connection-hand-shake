package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/lib/token"
	"logistics-platform/lib/utils"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	shutdown               chan struct{}
	wg                     sync.WaitGroup
}

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	service := &NotificationService{
		locationWriter:         utils.InitKafkaWriter("driver_locations"),
		notificationReader:     utils.InitKafkaReader("driver_notification", "driver_notification"),
		bookNotificationReader: InitKafkaReader("booking_notifications", "notification_service_group"),
		shutdown:               make(chan struct{}),
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.GET("/driver/ws", service.handleDriverWebSocket)
	router.GET("/user/ws", service.handleUserWebSocket)
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go service.consumeNotifications()
	go service.consumeBookingNotifications()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a timeout context for the entire shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Trigger shutdown for all goroutines
	close(service.shutdown)

	// Shutdown the HTTP server
	go func() {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	// Wait for all goroutines to finish or timeout
	done := make(chan struct{})
	go func() {
		service.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All goroutines finished")
	case <-ctx.Done():
		log.Println("Shutdown timed out")
	}

	// Close Kafka connections with timeout
	closeWithTimeout := func(closer func() error, name string) {
		ch := make(chan struct{})
		go func() {
			defer close(ch)
			if err := closer(); err != nil {
				log.Printf("Error closing %s: %v", name, err)
			}
		}()
		select {
		case <-ch:
			log.Printf("%s closed successfully", name)
		case <-time.After(5 * time.Second):
			log.Printf("Timeout closing %s", name)
		}
	}

	closeWithTimeout(service.locationWriter.Close, "location writer")
	closeWithTimeout(service.notificationReader.Close, "notification reader")
	closeWithTimeout(service.bookNotificationReader.Close, "book notification reader")

	log.Println("Server exiting")
	os.Exit(0)
}

// Update the InitKafkaReader function in your utils package
func InitKafkaReader(topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"localhost:9092"}, // Replace with your Kafka broker addresses
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
}

func (s *NotificationService) consumeNotifications() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			log.Println("Stopping notification consumer")
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg, err := s.notificationReader.FetchMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// No message available, wait a bit before trying again
					time.Sleep(1 * time.Second)
				} else {
					log.Printf("Error fetching message: %v", err)
				}
				continue
			}

			// Process the message
			var notification utils.BookingNotification
			if err := json.Unmarshal(msg.Value, &notification); err != nil {
				log.Printf("Error unmarshaling notification: %v", err)
			} else {
				if err := s.sendNotification(notification); err != nil {
					log.Printf("Error sending notification: %v", err)
				}
			}

			// Commit the message
			if err := s.notificationReader.CommitMessages(context.Background(), msg); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

func (s *NotificationService) consumeBookingNotifications() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case <-s.shutdown:
			log.Println("Stopping booking notification consumer")
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			msg, err := s.bookNotificationReader.FetchMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// log.Println("No new messages, waiting before next fetch")
					time.Sleep(1 * time.Second)
					continue
				}
				log.Printf("Error fetching booking notification message: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// log.Printf("Received message: %s", string(msg.Value))

			var notification utils.BookedNotification
			if err := json.Unmarshal(msg.Value, &notification); err != nil {
				log.Printf("Error unmarshaling notification: %v", err)
				// Consider handling this error (e.g., dead-letter queue)
				continue
			}

			// log.Printf("Processing notification: %+v", notification)

			// Handle notification status
			switch notification.Status {
			case "booked":
				s.driverUserConnections.Store(notification.DriverID, notification.UserID)
				s.notifyUser(notification.UserID, notification)
			case "completed":
				s.driverUserConnections.Delete(notification.DriverID)
				s.notifyUser(notification.UserID, notification)
				s.userConnections.Delete(notification.UserID)
			default:
				s.notifyUser(notification.UserID, notification)
			}

			if err := s.bookNotificationReader.CommitMessages(context.Background(), msg); err != nil {
				log.Printf("Error committing message: %v", err)
				// Consider handling this error (e.g., retry logic)
			} else {
				// log.Println("Message processed and committed successfully")
			}
		}
	}
}

func (s *NotificationService) notifyUser(userID string, notification utils.BookedNotification) {
	conn, ok := s.userConnections.Load(userID)
	if ok {
		if err := conn.(*websocket.Conn).WriteJSON(notification); err != nil {
			log.Printf("Error sending notification to user %s: %v", userID, err)
		}
	} else {
		log.Printf("User connection not found for userID: %s", userID)
	}
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
		conn, ok := s.userConnections.Load(userID.(string))
		if ok {
			err = conn.(*websocket.Conn).WriteJSON(location)
		} else {
			log.Print("User connection not found")
		}
	} else {
		err = s.locationWriter.WriteMessages(context.Background(),
			kafka.Message{Value: message})
	}

	return err
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
