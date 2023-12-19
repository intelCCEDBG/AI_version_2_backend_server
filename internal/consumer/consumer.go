package consumer

import (
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

	}
}
