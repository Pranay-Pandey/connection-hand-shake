package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"
)

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
		Brokers:        brokers, // Replace with your Kafka broker addresses
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
}
