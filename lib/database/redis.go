package database

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedis() (*redis.Client, error) {
	redisURL := viper.GetString("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
