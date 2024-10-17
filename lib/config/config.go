package config

import (
	"github.com/spf13/viper"
)

func LoadConfig() error {
	viper.SetConfigFile(".env")
	return viper.ReadInConfig()
}

func GetDBConnectionString() string {
	return viper.GetString("POSTGRES_URL")
}
