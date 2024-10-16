// lib/utils/utils.go
package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GeoPoint struct {
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
	Name      string  `json:"name" bson:"name"`
}

type DriverLocation struct {
	DriverID  string    `json:"driver_id" bson:"driver_id"`
	Location  GeoPoint  `json:"location" bson:"location"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

type BookingNotification struct {
	UserID   string   `json:"user_id" bson:"user_id"`
	DriverID string   `json:"driver_id" bson:"driver_id"`
	Price    float64  `json:"price" bson:"price"`
	Pickup   GeoPoint `json:"pickup" bson:"pickup"`
	Dropoff  GeoPoint `json:"dropoff" bson:"dropoff"`
	UserName string   `json:"user_name" bson:"user_name"`
	MongoID  string   `json:"mongo_id" bson:"mongo_id"`
}

type BookedNotification struct {
	UserID   string `json:"user_id"`
	DriverID string `json:"driver_id"`
	Status   string `json:"status"`
}

type UserRequest struct {
	UserID   string `json:"user_id" bson:"user_id"`
	UserName string `json:"user_name" bson:"user_name"`
}

type BookingRequest struct {
	UserID      string   `json:"user_id" bson:"user_id"`
	UserName    string   `json:"user_name" bson:"user_name"`
	Pickup      GeoPoint `json:"pickup" bson:"pickup"`
	Dropoff     GeoPoint `json:"dropoff" bson:"dropoff"`
	VehicleType string   `json:"vehicle_type" bson:"vehicle_type"`
	Price       float64  `json:"price" bson:"price"`
	MongoID     string   `json:"mongo_id" bson:"mongo_id"`
}

func InitPostgres() (*pgx.Conn, error) {
	dsn := viper.GetString("POSTGRES_URL")
	PostgreSQLConn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v\n", err)
	}
	return PostgreSQLConn, nil
}

func InitRedis() (*redis.Client, error) {
	redisURL := viper.GetString("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing Redis URL: %w", err)
	}
	redisClient := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}
	return redisClient, nil
}

func InitMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(viper.GetString("MONGO_URI")))
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %w", err)
	}
	return client, nil
}

func InitKafkaWriter(topic string) *kafka.Writer {
	brokers := viper.GetStringSlice("KAFKA_ADDR")
	return &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func InitKafkaReader(topic, groupID string) *kafka.Reader {
	brokers := viper.GetStringSlice("KAFKA_ADDR")
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
}

func WaitForShutdown(closers ...interface{ Close() error }) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			log.Printf("Error closing: %v", err)
		}
	}
}

func LoadConfig() error {
	viper.SetConfigFile(".env")
	return viper.ReadInConfig()
}

func GetDBConnectionString() string {
	return viper.GetString("POSTGRES_URL")
}
