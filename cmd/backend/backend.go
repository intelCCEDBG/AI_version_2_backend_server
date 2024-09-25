package main

import (
	"fmt"
	"recorder/config"
	"recorder/internal/router"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
	user_query "recorder/pkg/mariadb/user"
	"recorder/pkg/redis"
)

func main() {
	fmt.Println("Starting server...")
	config.LoadConfig()
	redis.RedisInit()
	logger.InitLogger(config.Viper.GetString("API_LOG_FILE_PATH"))
	err := mariadb.ConnectDB()
	if err != nil {
		logger.Error("Connect to mariadb error: " + err.Error())
		return
	}
	user_query.InsertInitUser()
	router.Start_backend()
}
