// services/booking/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"logistics-platform/lib/middlewares/auth"
	"logistics-platform/lib/middlewares/cors"
	"logistics-platform/lib/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingService struct {
	mongoClient        *mongo.Client
	notificationWriter *kafka.Writer
	bookingWriter      *kafka.Writer
	redisClient        *redis.Client
	PostgreSQLConn     *pgx.Conn
	shutdown           chan struct{}
	wg                 sync.WaitGroup
}

type Booking struct {
	ID          int32          `json:"id"`
	UserID      int32          `json:"user_id"`
	DriverID    int32          `json:"driver_id"`
	Price       float64        `json:"price"`
	Pickup      utils.GeoPoint `json:"pickup"`
	Dropoff     utils.GeoPoint `json:"dropoff"`
	BookedAt    time.Time      `json:"created_at"`
	CompletedAt time.Time      `json:"completed_at"`
	Status      string         `json:"status"`
}

type BookingConfirmation struct {
	BookingReq utils.BookingRequest `json:"booking_request"`
	DriverID   string               `json:"driver_id"`
}

func main() {
	if err := utils.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mongoClient, err := utils.InitMongoDB()
	defer mongoClient.Disconnect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	redisClient, err := utils.InitRedis()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	postgresConn, err := utils.InitPostgres()
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	service := &BookingService{
		mongoClient:        mongoClient,
		notificationWriter: utils.InitKafkaWriter("driver_notification"),
		redisClient:        redisClient,
		PostgreSQLConn:     postgresConn,
		bookingWriter:      utils.InitKafkaWriter("booking_notifications"),
		shutdown:           make(chan struct{}),
	}

	router := gin.Default()
	router.Use(cors.CORSMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	userGroup := router.Group("/user")
	userGroup.Use(auth.AuthInjectionMiddleware())
	{
		userGroup.GET("/booking", service.handleUserBookingCheck)
		userGroup.GET("/booking-history", service.handleUserBookingHistory)
	}

	driverGroup := router.Group("/driver")
	driverGroup.Use(auth.AuthInjectionMiddleware())
	{
		driverGroup.GET("/booking", service.handleDriverBookingCheck)
		driverGroup.GET("/booking-history", service.handleDriverBookingHistory)
	}

	router.Use(auth.AuthInjectionMiddleware())
	{
		router.POST("/booking/accept", service.handleBookingAccept)
		router.POST("/booking", service.handleBookingRequest)
		router.PATCH("/booking/:userId", service.handleBookingUpdate)
	}

	server := &http.Server{
		Addr:    ":8084",
		Handler: router,
	}

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

	closeWithTimeout(service.notificationWriter.Close, "notification writer")
	closeWithTimeout(service.bookingWriter.Close, "book notification writer")

	// Close PostgreSQL connection
	if err := postgresConn.Close(context.Background()); err != nil {
		log.Printf("Error closing PostgreSQL connection: %v", err)
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Server exiting")
	os.Exit(0)
}

func (s *BookingService) handleBookingUpdate(c *gin.Context) {
	authDriver, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	driver, ok := authDriver.(utils.UserRequest)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	userID := c.Param("userId")
	var booking Booking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if booking.Status == "completed" || booking.Status == "cancelled" {
		pgComm, err := s.PostgreSQLConn.Exec(context.Background(), "UPDATE booking SET status = $1, completed_at = NOW() WHERE user_id = $2 AND driver_id = $3 AND status != $4", booking.Status, userID, driver.UserID, "completed")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating booking"})
			return
		}

		if pgComm.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
	} else {
		pgComm, err := s.PostgreSQLConn.Exec(context.Background(), "UPDATE booking SET status = $1 WHERE user_id = $2 AND driver_id = $3 AND status != $4", booking.Status, userID, driver.UserID, "completed")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating booking"})
			return
		}

		if pgComm.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
			return
		}
	}

	go s.produceBookingEvent(userID, driver.UserID, booking.Status)

	c.JSON(http.StatusOK, gin.H{"message": "Booking updated"})
}

func (s *BookingService) handleBookingAccept(c *gin.Context) {
	driverUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	driver, ok := driverUser.(utils.UserRequest)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	var mongoID struct {
		MongoID string `json:"mongo_id"`
	}
	if err := c.ShouldBindJSON(&mongoID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	var bookConReq BookingConfirmation
	objectID, err := primitive.ObjectIDFromHex(mongoID.MongoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid mongo id"})
		return
	}
	if err := collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&bookConReq.BookingReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking request not found"})
		return
	}

	bookConReq.DriverID = driver.UserID

	// delete the booking request
	if _, err := collection.DeleteOne(context.Background(), bson.M{"_id": mongoID.MongoID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting booking request"})
		return
	}
	err = s.processBooked(bookConReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking accepted", "user_id": bookConReq.BookingReq.UserID})
}

func (s *BookingService) processBooked(bookConReq BookingConfirmation) error {
	// make a new booking in the postgres database
	bookingReq := bookConReq.BookingReq
	_, err := s.PostgreSQLConn.Exec(context.Background(), "INSERT INTO booking (user_id, driver_id, pickup_latitude, pickup_longitude, dropoff_latitude, dropoff_longitude, vehicle_type, price, status, pickup_name, dropoff_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", bookingReq.UserID, bookConReq.DriverID, bookingReq.Pickup.Latitude, bookingReq.Pickup.Longitude, bookingReq.Dropoff.Latitude, bookingReq.Dropoff.Longitude, bookingReq.VehicleType, bookingReq.Price, "enroute_to_pickup", bookConReq.BookingReq.Pickup.Name, bookConReq.BookingReq.Dropoff.Name)

	if err != nil {
		return fmt.Errorf("error storing booking: %w", err)
	}

	go s.produceBookingEvent(bookingReq.UserID, bookConReq.DriverID, "booked")
	return nil
}

func (s *BookingService) produceBookingEvent(userID, driverID, status string) {
	bookingEvent := utils.BookedNotification{
		UserID:   userID,
		DriverID: driverID,
		Status:   status,
	}

	bookingEventJSON, err := json.Marshal(bookingEvent)
	if err != nil {
		log.Printf("Error marshaling booking event: %v", err)
		return
	}

	if err := s.bookingWriter.WriteMessages(context.Background(), kafka.Message{Value: bookingEventJSON}); err != nil {
		log.Printf("Error writing booking event: %v", err)
		return
	}
}

func (s *BookingService) handleBookingRequest(c *gin.Context) {
	var bookingReq utils.BookingRequest
	if err := c.ShouldBindJSON(&bookingReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	user, _ := authUser.(utils.UserRequest)

	bookingReq.UserID = user.UserID
	bookingReq.UserName = user.UserName

	go s.processBookingRequest(bookingReq)

	c.JSON(http.StatusOK, gin.H{"message": "Booking request received"})
}

func (s *BookingService) processBookingRequest(bookingReq utils.BookingRequest) error {
	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	res, err := collection.InsertOne(context.Background(), bookingReq)
	if err != nil {
		return fmt.Errorf("error storing booking request: %w", err)
	}

	newBookingMongoID := res.InsertedID.(primitive.ObjectID).Hex()
	bookingReq.MongoID = newBookingMongoID

	nearbyDrivers, err := s.findNearbyDrivers(bookingReq, bookingReq.VehicleType)
	if err != nil {
		return fmt.Errorf("error finding nearby drivers: %w", err)
	}

	for _, driverID := range nearbyDrivers {
		go func(driverID string) {
			if err := s.notifyDriver(driverID, bookingReq); err != nil {
				log.Printf("Error notifying driver %s: %v", driverID, err)
			}
		}(driverID)
	}

	return nil
}

func (s *BookingService) findNearbyDrivers(bookingReq utils.BookingRequest, vehicleType string) ([]string, error) {
	pickup := bookingReq.Pickup
	drivers, err := s.redisClient.GeoRadius(context.Background(), "driver_locations", pickup.Longitude, pickup.Latitude, &redis.GeoRadiusQuery{
		Radius: 1000,
		Unit:   "km",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("error finding nearby drivers: %w", err)
	}

	var nearbyDrivers []string
	for _, driver := range drivers {
		go func(driverID string) {
			if err := s.notifyDriver(driverID, bookingReq); err != nil {
				log.Printf("Error notifying driver %s: %v", driverID, err)
			}
		}(driver.Name)
	}
	return nearbyDrivers, nil
}

func (s *BookingService) notifyDriver(driverID string, bookingReq utils.BookingRequest) error {
	notification := utils.BookingNotification{
		UserID:   bookingReq.UserID,
		Price:    bookingReq.Price,
		DriverID: driverID,
		Pickup:   bookingReq.Pickup,
		Dropoff:  bookingReq.Dropoff,
		UserName: bookingReq.UserName,
		MongoID:  bookingReq.MongoID,
	}
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("error marshaling notification: %w", err)
	}

	return s.notificationWriter.WriteMessages(context.Background(), kafka.Message{Value: notificationJSON})
}

func (s *BookingService) handleUserBookingCheck(c *gin.Context) {
	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	user, _ := authUser.(utils.UserRequest)

	// check if the user has any booking request made which is in mongodb
	// then check if the user has any booking made which is in postgres
	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	var bookingReq utils.BookingRequest
	if err := collection.FindOne(context.Background(), bson.M{"user_id": user.UserID}).Decode(&bookingReq); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking request found"})
		return
	}

	// check if the user has any booking made which is in postgres where status is not completed or not cancelled
	var booking Booking
	if err := s.PostgreSQLConn.QueryRow(context.Background(), "SELECT * FROM booking WHERE user_id = $1 AND status != $2 AND status != $3", user.UserID, "completed", "cancelled").Scan(&booking.ID, &booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &booking.CompletedAt, &booking.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking_request": bookingReq, "booking": booking})
}

func (s *BookingService) handleUserBookingHistory(c *gin.Context) {
	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	user, _ := authUser.(utils.UserRequest)

	// check if the user has any booking made which is in postgres
	rows, err := s.PostgreSQLConn.Query(context.Background(), "SELECT user_id, driver_id, price, pickup_latitude, pickup_longitude, dropoff_latitude, dropoff_longitude, created_at, completed_at, status, pickup_name, dropoff_name FROM booking WHERE user_id = $1", user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history", "err": err})
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &booking.CompletedAt, &booking.Status, &booking.Pickup.Name, &booking.Dropoff.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history",
				"err": err})
			return
		}
		bookings = append(bookings, booking)
	}

	if len(bookings) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (s *BookingService) handleDriverBookingCheck(c *gin.Context) {
	authDriver, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	driver, _ := authDriver.(utils.UserRequest)

	// check if the driver has any booking made which is in postgres where status is not completed or not cancelled
	var booking Booking
	if err := s.PostgreSQLConn.QueryRow(context.Background(), "SELECT * FROM booking WHERE driver_id = $1 AND status != $2 AND status != $3", driver.UserID, "completed", "cancelled").Scan(&booking.ID, &booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &booking.CompletedAt, &booking.Status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"booking": booking})
}

func (s *BookingService) handleDriverBookingHistory(c *gin.Context) {
	authDriver, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	driver, _ := authDriver.(utils.UserRequest)

	// check if the driver has any booking made which is in postgres
	rows, err := s.PostgreSQLConn.Query(context.Background(), "SELECT user_id, driver_id, price, pickup_latitude, pickup_longitude, dropoff_latitude, dropoff_longitude, created_at, completed_at, status, pickup_name, dropoff_name FROM booking WHERE driver_id = $1", driver.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history"})
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &booking.CompletedAt, &booking.Status, &booking.Pickup.Name, &booking.Dropoff.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history"})
			return
		}
		bookings = append(bookings, booking)
	}

	if len(bookings) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}
