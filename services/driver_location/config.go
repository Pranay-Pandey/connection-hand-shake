package main

import "github.com/spf13/viper"

var (
    KafkaBrokers string
    MongoDBUri   string
)

func LoadConfig() {
    viper.SetConfigFile(".env")
    viper.ReadInConfig()

    KafkaBrokers = viper.GetString("KAFKA_BROKERS")
    MongoDBUri = viper.GetString("
