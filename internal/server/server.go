package server

import (
	"fmt"
	"recorder/config"
	"recorder/internal/loadstatus"
	service "recorder/internal/mainservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
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
	loadstatus.Loadstatus()
	service.Start_service()
}
