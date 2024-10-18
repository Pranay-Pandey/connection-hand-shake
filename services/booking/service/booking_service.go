package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"logistics-platform/lib/utils"
	"logistics-platform/services/booking/interfaces"
	"logistics-platform/services/booking/models"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	kafkaConfig "logistics-platform/lib/kafka"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
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
	PostgreSQLConn     *pgxpool.Pool
	shutdown           chan struct{}
	wg                 sync.WaitGroup
}

type Booking struct {
	ID          int32          `json:"id"`
	UserID      int32          `json:"user_id"`
	UserName    string         `json:"user_name,omitempty"`
	DriverID    int32          `json:"driver_id"`
	DriverName  string         `json:"driver_name,omitempty"`
	Price       float64        `json:"price"`
	Pickup      utils.GeoPoint `json:"pickup"`
	Dropoff     utils.GeoPoint `json:"dropoff"`
	BookedAt    time.Time      `json:"created_at"`
	CompletedAt time.Time      `json:"completed_at"`
	Status      string         `json:"status"`
}

func NewBookingService(mongoClient *mongo.Client, redisClient *redis.Client, pool *pgxpool.Pool) interfaces.BookingInterface {
	return &BookingService{
		mongoClient:        mongoClient,
		notificationWriter: kafkaConfig.InitKafkaWriter("driver_notification"),
		redisClient:        redisClient,
		PostgreSQLConn:     pool,
		bookingWriter:      kafkaConfig.InitKafkaWriter("booking_notifications"),
		shutdown:           make(chan struct{}),
	}
}

func (s *BookingService) HandleBookingUpdate(c *gin.Context) {
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

	go s.ProduceBookingEvent(userID, driver.UserID, driver.UserName, booking.Status)

	c.JSON(http.StatusOK, gin.H{"message": "Booking updated"})
}

func (s *BookingService) HandleBookingAccept(c *gin.Context) {
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
	var bookConReq models.BookingConfirmation
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
	bookConReq.DriverName = driver.UserName

	// delete the booking request
	if _, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error deleting booking request"})
		return
	}
	err = s.ProcessBooked(bookConReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking accepted", "user_id": bookConReq.BookingReq.UserID})
}

func (s *BookingService) ProcessBooked(bookConReq models.BookingConfirmation) error {
	// make a new booking in the postgres database
	bookingReq := bookConReq.BookingReq
	_, err := s.PostgreSQLConn.Exec(context.Background(), "INSERT INTO booking (user_id, driver_id, pickup_latitude, pickup_longitude, dropoff_latitude, dropoff_longitude, vehicle_type, price, status, pickup_name, dropoff_name) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", bookingReq.UserID, bookConReq.DriverID, bookingReq.Pickup.Latitude, bookingReq.Pickup.Longitude, bookingReq.Dropoff.Latitude, bookingReq.Dropoff.Longitude, bookingReq.VehicleType, bookingReq.Price, "enroute_to_pickup", bookConReq.BookingReq.Pickup.Name, bookConReq.BookingReq.Dropoff.Name)

	if err != nil {
		return fmt.Errorf("error storing booking: %w", err)
	}

	go s.ProduceBookingEvent(bookingReq.UserID, bookConReq.DriverID, bookConReq.DriverName, "booked")
	return nil
}

func (s *BookingService) ProduceBookingEvent(userID, driverID, driverName, status string) {
	bookingEvent := utils.BookedNotification{
		UserID:     userID,
		DriverID:   driverID,
		DriverName: driverName,
		Status:     status,
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

func (s *BookingService) HandleBookingRequest(c *gin.Context) {
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

	go s.ProcessBookingRequest(bookingReq)

	c.JSON(http.StatusOK, gin.H{"message": "Booking request received"})
}

func (s *BookingService) ProcessBookingRequest(bookingReq utils.BookingRequest) error {
	bookingReq.CreatedAt = time.Now()

	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	res, err := collection.InsertOne(context.Background(), bookingReq)
	if err != nil {
		return fmt.Errorf("error storing booking request: %w", err)
	}

	newBookingMongoID := res.InsertedID.(primitive.ObjectID).Hex()
	bookingReq.MongoID = newBookingMongoID

	return s.FindAndNotifyNearbyDrivers(bookingReq, bookingReq.VehicleType)
}

func (s *BookingService) FindAndNotifyNearbyDrivers(bookingReq utils.BookingRequest, vehicleType string) error {
	pickup := bookingReq.Pickup
	drivers, err := s.redisClient.GeoRadius(context.Background(), "driver_locations", pickup.Longitude, pickup.Latitude, &redis.GeoRadiusQuery{
		Radius: 1000,
		Unit:   "km",
	}).Result()
	if err != nil {
		return fmt.Errorf("error finding nearby drivers: %w", err)
	}

	for _, driver := range drivers {
		go func(driverID string) {

			// check if the driver has the same vehicle type
			driverVehicleType, err := s.GetVehicleType(driverID)
			if err != nil {
				log.Printf("Error getting vehicle type for driver %s: %v", driverID, err)
				return
			}

			if driverVehicleType != vehicleType {
				return
			}

			if err := s.NotifyDriver(driverID, bookingReq); err != nil {
				log.Printf("Error notifying driver %s: %v", driverID, err)
			}
		}(driver.Name)
	}
	return nil
}

func (s *BookingService) NotifyDriver(driverID string, bookingReq utils.BookingRequest) error {
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

func (s *BookingService) HandleUserBookingCheck(c *gin.Context) {
	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	user, _ := authUser.(utils.UserRequest)

	// Check if the user has any booking request made in MongoDB
	collection := s.mongoClient.Database("logistics").Collection("booking_requests")
	var bookingReq utils.BookingRequest
	err := collection.FindOne(context.Background(), bson.M{"user_id": user.UserID}).Decode(&bookingReq)

	if err == nil {
		// Booking request found in MongoDB
		c.JSON(http.StatusOK, gin.H{"booking_request": bookingReq})
		return
	} else if err != mongo.ErrNoDocuments {
		// An error occurred that is not a 'not found' error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error", "err": err})
		return
	}

	// If no booking request found in MongoDB, check PostgreSQL
	userId, err := strconv.Atoi(user.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Check if the user has any booking made in PostgreSQL where status is not completed or cancelled
	var booking Booking
	err = s.PostgreSQLConn.QueryRow(context.Background(),
		"SELECT b.user_id, b.driver_id, b.price, b.pickup_latitude, b.pickup_longitude, b.dropoff_latitude, b.dropoff_longitude, b.created_at, b.status, b.pickup_name, b.dropoff_name, d.name FROM booking b INNER JOIN vehicle_drivers d ON d.id=b.driver_id WHERE b.user_id=$1 AND status!=$2 AND status != $3",
		userId, "completed", "cancelled").Scan(&booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &booking.Status, &booking.Pickup.Name, &booking.Dropoff.Name, &booking.DriverName)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error", "err": err})
		return
	}

	// If a booking is found in PostgreSQL
	c.JSON(http.StatusOK, gin.H{"booking_request": nil, "booking": booking})
}

func (s *BookingService) HandleUserBookingHistory(c *gin.Context) {
	authUser, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	user, _ := authUser.(utils.UserRequest)

	// check if the user has any booking made which is in postgres
	rows, err := s.PostgreSQLConn.Query(context.Background(), "SELECT b.user_id, b.driver_id, b.price, b.pickup_latitude, b.pickup_longitude, b.dropoff_latitude, b.dropoff_longitude, b.created_at, b.completed_at, b.status, b.pickup_name, b.dropoff_name, d.name FROM booking b INNER JOIN vehicle_drivers d on d.id=b.driver_id WHERE user_id=$1", user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history", "err": err})
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		completedAt := new(time.Time)
		if err := rows.Scan(&booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &completedAt, &booking.Status, &booking.Pickup.Name, &booking.Dropoff.Name, &booking.DriverName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading booking history",
				"err": err})
			return
		}
		if completedAt != nil {
			booking.CompletedAt = *completedAt
		}
		bookings = append(bookings, booking)
	}

	if len(bookings) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (s *BookingService) HandleDriverBookingCheck(c *gin.Context) {
	authDriver, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	driver, _ := authDriver.(utils.UserRequest)

	driverID, err := strconv.Atoi(driver.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver id"})
		return
	}
	// Check if the user has any booking made in PostgreSQL where status is not completed or cancelled
	var booking Booking
	err = s.PostgreSQLConn.QueryRow(context.Background(),
		"SELECT b.user_id, b.driver_id, b.price, b.pickup_latitude, b.pickup_longitude, b.dropoff_latitude, b.dropoff_longitude, b.created_at, b.status, b.pickup_name, b.dropoff_name, u.name FROM booking b INNER JOIN users u ON u.id=b.user_id WHERE driver_id=$1 AND status!=$2 AND status!=$3",
		driverID, "completed", "cancelled").Scan(&booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &booking.Status, &booking.Pickup.Name, &booking.Dropoff.Name, &booking.UserName)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error", "err": err})
		return
	}

	// If a booking is found in PostgreSQL
	c.JSON(http.StatusOK, gin.H{"booking": booking})
}

func (s *BookingService) HandleDriverBookingHistory(c *gin.Context) {
	authDriver, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth token"})
		return
	}

	driver, _ := authDriver.(utils.UserRequest)

	// check if the driver has any booking made which is in postgres
	rows, err := s.PostgreSQLConn.Query(context.Background(), "SELECT b.user_id, b.driver_id, b.price, b.pickup_latitude, b.pickup_longitude, b.dropoff_latitude, b.dropoff_longitude, b.created_at, b.completed_at, b.status, b.pickup_name, b.dropoff_name, u.name FROM booking b INNER JOIN users u on u.id=b.user_id WHERE driver_id=$1", driver.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history"})
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var booking Booking
		completedAt := new(time.Time)
		if err := rows.Scan(&booking.UserID, &booking.DriverID, &booking.Price, &booking.Pickup.Latitude, &booking.Pickup.Longitude, &booking.Dropoff.Latitude, &booking.Dropoff.Longitude, &booking.BookedAt, &completedAt, &booking.Status, &booking.Pickup.Name, &booking.Dropoff.Name, &booking.UserName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching booking history"})
			return
		}
		if completedAt != nil {
			booking.CompletedAt = *completedAt
		}
		bookings = append(bookings, booking)
	}

	if len(bookings) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no booking found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (s *BookingService) GetVehicleType(driverID string) (string, error) {
	// check if redis has the vehicle type using key driverId-veh: type
	vehicleType, err := s.redisClient.Get(context.Background(), driverID+"-veh").Result()
	// if error is nill and vehicle type is not empty, return the vehicle type
	if err == nil && vehicleType != "" {
		return vehicleType, nil
	}

	// if vehicle type is not found in redis, fetch it from postgres
	if vehicleType == "" {
		err := s.PostgreSQLConn.QueryRow(context.Background(), "SELECT vehicle_type FROM vehicle_drivers WHERE id=$1", driverID).Scan(&vehicleType)
		if err != nil {
			return "", fmt.Errorf("error fetching vehicle type: %w", err)
		}
	}

	if vehicleType == "" {
		return "", fmt.Errorf("vehicle type not found")
	}

	expiration := 1 * time.Hour
	// set the vehicle type in redis
	_ = s.redisClient.Set(context.Background(), driverID+"-veh", vehicleType, expiration).Err()

	return vehicleType, nil
}

func (s *BookingService) GracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a timeout context for the entire shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Trigger shutdown for all goroutines
	close(s.shutdown)

	// Shutdown the HTTP server
	go func() {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	// Wait for all goroutines to finish or timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
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

	closeWithTimeout(s.notificationWriter.Close, "notification writer")
	closeWithTimeout(s.bookingWriter.Close, "book notification writer")

	// Close Redis connection
	if err := s.redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Server exiting")
	os.Exit(0)
}

// func createTTLIndex(collection *mongo.Collection) error {
// 	indexModel := mongo.IndexModel{
// 		Keys: bson.D{
// 			{Key: "created_at", Value: 1}, // Create index on CreatedAt
// 		},
// 		Options: options.Index().SetExpireAfterSeconds(600), // time in seconds
// 	}

// 	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
// 	return err
// }
