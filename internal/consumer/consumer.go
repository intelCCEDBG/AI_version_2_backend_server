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
	"time"
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
	redis.RedisInit()
	queue, err := rabbitmq.Consume("result_queue")
	for err != nil {
		logger.Error(err.Error())
		println("Waiting 5 seconds to reconnect...")
		time.Sleep(5 * time.Second)
		queue, err = rabbitmq.Consume("result_queue")
	}
	for msg := range queue {
		var data structure.ResultMessage
		fmt.Println(string(msg.Body[:]))
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			logger.Error(err.Error())
		} else {
			dut_query.UpdateAIResult(data.Hostname, data.Label, data.Coords)
			// logger.Info("Received from AI: " + data.Hostname)
		}
	}
}
