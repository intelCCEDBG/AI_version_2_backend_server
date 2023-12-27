package consumer

import (
	"encoding/json"
	"fmt"
	"recorder/config"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
	dut_query "recorder/pkg/mariadb/dut"
	"recorder/pkg/rabbitmq"
	"recorder/pkg/redis"
)

func Start_consumer() {
	config.LoadConfig()
	logger.InitLogger(config.Viper.GetString("CONSUMER_LOG_FILE_PATH"))
	err := mariadb.ConnectDB()
	if err != nil {
		logger.Error("Connect to mariadb error: " + err.Error())
		return
	}
	rabbitmq.Rabbit_init()
	redis.Redis_init()
	queue, err := rabbitmq.Consume("result_queue")
	if err != nil {
		logger.Error(err.Error())
	}
	for msg := range queue {
		var data structure.Result_message
		fmt.Println(string(msg.Body[:]))
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			logger.Error(err.Error())
		} else {
			dut_query.Update_AI_result(data.Hostname, data.Label, data.Coords)
			// logger.Info("Received from AI: " + data.Hostname)
		}
	}
}
