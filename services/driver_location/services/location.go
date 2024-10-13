package services

import (
	"context"
	"log"
	"logistics-platform/models"
)

func StoreAndPublishLocation(driverID string, location models.Location) {
	go StoreLocationInRedis(driverID, location)
	go PublishLocationToKafka(driverID, location)
}

func StoreLocationInRedis(driverID string, location models.Location) {
	err := RedisClient.Set(ctx, "driver_location:"+driverID, location.ToJSON(), 0).Err()
	if err != nil {
		log.Println("Failed to store location in Redis:", err)
	}
}

func StoreLocationInMongoDB(location models.Location) {
	collection := MongoDB.Collection("location_history")
	_, err := collection.InsertOne(context.TODO(), location)
	if err != nil {
		log.Println("Failed to store location in MongoDB:", err)
	}
}
