package server

import (
	"fmt"
	"os"
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
	logger.InitLogger(config.Viper.GetString("RECORDER_LOG_FILE_PATH"))
	err := mariadb.ConnectDB()
	if err != nil {
		logger.Error("Connect to mariadb error: " + err.Error())
		return
	}
	// loadstatus.Loadstatus()
	redis.RedisInit()
	redis.RedisSet("test", "test")
	redis.RedisClear()
	rabbitmq.Rabbit_init()
	folder_check()
	// go loadstatus.Sync_with_mariadb()
	service.Start_service()
}

func folder_check() {
	paths := config.Find_suffix("PATH")
	for _, value := range paths {
		createFolderIfNotExist(value)
	}
}

func createFolderIfNotExist(folderPath string) error {
	_, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		// Folder does not exist, create it
		err := os.MkdirAll(folderPath, 0755) // 0755 is the permission mode for the folder
		if err != nil {
			return err
		}
		fmt.Println("Folder created successfully:", folderPath)
	} else if err != nil {
		// Some other error occurred
		return err
	} else {
		fmt.Println("Folder already exists:", folderPath)
	}
	return nil
}
