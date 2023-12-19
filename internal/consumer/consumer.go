package consumer

import (
	"encoding/json"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/rabbitmq"
)

func Start_consumer() {
	rabbitmq.Rabbit_init()
	queue, err := rabbitmq.Consume("result_queue")
	if err != nil {
		logger.Error(err.Error())
	}
	for msg := range queue {
		var data structure.Result_message
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			logger.Error(err.Error())
		} else {
			logger.Info("Received from AI: " + data.Hostname + " " + data.Coord)
		}
	}
}
