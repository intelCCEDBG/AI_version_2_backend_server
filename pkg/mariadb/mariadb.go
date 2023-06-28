package mariadb

import (
	"database/sql"
	"fmt"
	"recorder/config"
)

var DB *sql.DB

func ConnectDB() error {
	var err error
	dbUser := config.Viper.GetString("MARIA_USER")
	dbPass := config.Viper.GetString("MARIA_PASSWORD")
	dbHost := config.Viper.GetString("MARIA_HOST")
	dbPort := config.Viper.GetInt("MARIA_PORT")
	dbName := config.Viper.GetString("MARIA_DB")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	DB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	return nil
}
