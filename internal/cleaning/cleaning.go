package cleaning

import (
	"fmt"
	"os"
	"recorder/config"
	"recorder/pkg/logger"
	"time"
)

func Clean_service(cycle int, period int) {
	config.LoadConfig()
	logger.InitLogger(config.Viper.GetString("CLEANING_LOG_FILE_PATH"))
	logger.Info("Start cleaning service")
	for {
		currenttime := time.Now()
		logger.Debug("Cleaning at " + currenttime.String())
		cleaning(currenttime, period)
		time.Sleep(time.Duration(cycle) * time.Second)
	}
}

func cleaning(currenttime time.Time, period int) {
	path := config.Viper.GetString("RECORDING_PATH")
	items, _ := os.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			subitems, _ := os.ReadDir(path + item.Name())
			for _, subitem := range subitems {
				if !subitem.IsDir() {
					// fmt.Println(subitem.Name())
					if len(subitem.Name()) > 19 {
						tm, err := time.Parse("2006-01-02_15-04-05", subitem.Name()[:19])
						if err != nil {
							fmt.Println(err)
						}
						if currenttime.Sub(tm).Hours() > float64(period) && err == nil {
							os.Remove(path + item.Name() + "/" + subitem.Name())
							fmt.Println("Removed " + item.Name() + "/" + subitem.Name())
						}

					}
				}
			}
		}
	}
}
