package server

import (
	"fmt"
	"recorder/config"
	service "recorder/internal/mainservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
	"recorder/pkg/rabbitmq"
	"recorder/pkg/redis"
)

func Start_server() {
	fmt.Println("Starting server...")
	config.LoadConfig()
	logger.InitLogger(config.Viper.GetString("LOG_FILE_PATH"))
	err := mariadb.ConnectDB()
	if err != nil {
		logger.Error("Connect to mariadb error: " + err.Error())
		return
	}
	// loadstatus.Loadstatus()
	redis.Redis_init()
	redis.Redis_set("test", "test")
	redis.Redis_clear()
	rabbitmq.Rabbit_init()
	// go loadstatus.Sync_with_mariadb()
	service.Start_service()
}
