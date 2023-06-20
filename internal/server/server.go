package server

import (
	"fmt"
	"recorder/config"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
)

func Start_server() {
	fmt.Println("Starting server...")
	logger.InitLogger(config.Viper.GetString("LOG_FILE_PATH"))
	mariadb.InitDB()

}
