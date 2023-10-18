package redis

import (
	"context"
	"fmt"
	"recorder/config"
	"recorder/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func Redis_init() *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "10.227.106.11:" + config.Viper.GetString("REDIS_PORT"),
		Password: config.Viper.GetString("REDIS_PASSWORD"),
		DB:       config.Viper.GetInt("REDIS_DB"),
	})
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("Error connecting to redis")
		return nil
	}
	return RedisClient
}

func Redis_close() {
	RedisClient.Close()
}

func Redis_set(key string, value string) {
	RedisClient.Set(context.Background(), key, value, 0)
}

func Redis_get(key string) string {
	val, _ := RedisClient.Get(context.Background(), key).Result()
	return val
}

func Redis_get_by_pattern(pattern string) []string {
	val, err := RedisClient.Keys(context.Background(), pattern).Result()
	if err != nil {
		logger.Error("Search pattern in redis error: " + err.Error())
		return nil
	}
	return val
}

func Redis_append(key string, value string) {
	RedisClient.Append(context.Background(), key, value)
}

func Redis_del(key string) {
	RedisClient.Del(context.Background(), key)
}
