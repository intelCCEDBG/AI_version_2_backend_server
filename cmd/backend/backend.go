package main

import (
	"fmt"
	"recorder/config"
	"recorder/internal/router"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
)

func main() {
	fmt.Println("Starting server...")
	config.LoadConfig()
	logger.InitLogger(config.Viper.GetString("LOG_FILE_PATH"))
	err := mariadb.ConnectDB()
	if err != nil {
		logger.Error("Connect to mariadb error: " + err.Error())
		return
	}
	router.Start_backend()
}
