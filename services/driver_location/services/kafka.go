package services

import (
	"log"
	"logistics-platform/services/driver_location/models"

	"github.com/IBM/sarama"
)

var kafkaProducer sarama.SyncProducer

func InitKafkaProducer() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		log.Fatal("Failed to start Sarama producer:", err)
	}
	kafkaProducer = producer
}

func PublishLocationToKafka(driverID string, location models.Location) {
	locationJSON := location.ToJSON()
	msg := &sarama.ProducerMessage{
		Topic: "driver-location-updates",
		Key:   sarama.StringEncoder(driverID),
		Value: sarama.StringEncoder(locationJSON),
	}
	_, _, err := kafkaProducer.SendMessage(msg)
	if err != nil {
		log.Println("Failed to publish location to Kafka:", err)
	}
}

// func StartKafkaConsumer() {
// 	consumer, err := sarama.NewConsumer([]string{config.KafkaBrokers}, nil)
// 	if err != nil {
// 		log.Fatal("Failed to start Kafka consumer:", err)
// 	}

// 	partitionConsumer, err := consumer.ConsumePartition("driver-location-updates", 0, sarama.OffsetNewest)
// 	if err != nil {
// 		log.Fatal("Failed to start partition consumer:", err)
// 	}
// 	defer partitionConsumer.Close()

// 	for message := range partitionConsumer.Messages() {
// 		var location models.Location
// 		// Unmarshal message.Value and store location in MongoDB
// 		StoreLocationInMongoDB(location)
// 	}
// }
