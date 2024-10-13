package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	PostgreSQLConn *pgx.Conn
	MongoDBClient  *mongo.Client
	RedisClient    *redis.Client
)

func InitPostgres() {
	dsn := viper.GetString("POSTGRES_URL")
	var err error
	PostgreSQLConn, err = pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v\n", err)
	}

	log.Println("Connected to PostgreSQL")
}

func InitMongo() {
	mongoURI := viper.GetString("MONGO_URL")
	var err error
	MongoDBClient, err = mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v\n", err)
	}
	log.Println("Connected to MongoDB")
}

func InitRedis() {
	redisAddr := viper.GetString("REDIS_URL")

	// Updated redis client initialization for v9
	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // use default DB
	})

	// Ping to check connection
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}

	log.Println("Connected to Redis")
}
