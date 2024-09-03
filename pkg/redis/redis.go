package redis

import (
	"context"
	"fmt"
	"recorder/config"
	"recorder/pkg/logger"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func RedisInit() *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Viper.GetString("REDIS_HOST") + ":" + config.Viper.GetString("REDIS_PORT"),
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

func RedisClear() {
	statusCmd := RedisClient.FlushAll(context.Background())
	if err := statusCmd.Err(); err != nil {
		fmt.Println("Error flushing database:", err)
		return
	}
	fmt.Println("All data flushed from the Redis database.")
}

func RedisClose() {
	RedisClient.Close()
}

func RedisSet(key string, value string) {
	RedisClient.Set(context.Background(), key, value, 0)
}

func RedisGet(key string) string {
	val, _ := RedisClient.Get(context.Background(), key).Result()
	return val
}

func RedisGetByPattern(pattern string) []string {
	val, err := RedisClient.Keys(context.Background(), pattern).Result()
	if err != nil {
		logger.Error("Search pattern in redis error: " + err.Error())
		return nil
	}
	return val
}

func RedisAppend(key string, value string) {
	RedisClient.Append(context.Background(), key, value)
}

func RedisDel(key string) {
	RedisClient.Del(context.Background(), key)
}
