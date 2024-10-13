// services/notification/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"

	"net/http"
	"time"
)

var (
	upgrader           = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	driverConnections  = make(map[string]*websocket.Conn)
	userConnections    = make(map[string]*websocket.Conn)
	LocationWriter     *kafka.Writer
	NotificationReader *kafka.Reader
)

type DriverLocation struct {
	DriverID  string    `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

type BookingNotification struct {
	UserID string  `json:"user_id"`
	Price  float64 `json:"price"`
}

func main() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	initKafka()

	router := gin.Default()
	router.GET("/driver/ws", driverWebSocket)

	go consumeNotifications()

	router.Run(":8080")
}

func driverWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	driverID := c.Query("driver_id")
	driverConnections[driverID] = conn

	for {
		var location DriverLocation
		err := conn.ReadJSON(&location)
		if err != nil {
			log.Println(err)
			break
		}

		message, err := json.Marshal(location)
		if err != nil {
			log.Println(err)
			break
		}

		LocationWriter.WriteMessages(c, kafka.Message{
			Value: message,
		})
	}
}

func initKafka() {
	brokers := viper.GetStringSlice("KAFKA_ADDR")

	LocationWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    "driver_locations",
		Balancer: &kafka.LeastBytes{},
	})

	NotificationReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "driver_notification",
		GroupID: "notification",
	})
}

func consumeNotifications() {
	for {
		msg, err := NotificationReader.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		var notification BookingNotification
		err = json.Unmarshal(msg.Value, &notification)
		if err != nil {
			log.Println(err)
			continue
		}

		message := "You have a new booking request from " + notification.UserID + " for $" + fmt.Sprintf("%.2f", notification.Price)
		err = sendNotification(notification.UserID, message)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func sendNotification(userID, message string) error {
	conn, ok := driverConnections[userID]
	if !ok {
		return fmt.Errorf("driver not connected")
	}

	return conn.WriteJSON(map[string]string{
		"message": message,
	})
}
