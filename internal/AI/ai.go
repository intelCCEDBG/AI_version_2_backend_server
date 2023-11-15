package ai

import (
	"encoding/json"
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	"recorder/pkg/rabbitmq"
	"time"
)

var AI_list []string

type Message struct {
	Hostname string `json:"hostname"`
	// Path     string `json:"path"`
}

func Start_ai_monitoring() {
	_, err := rabbitmq.Declare("AI_queue1")
	if err != nil {
		logger.Error("Declare to rabbit error: " + err.Error())
		return
	}
	// var message Message
	// message.Hostname = "18F_AI05"
	// // message.Path = "/home/media/image" + element.Hostname + "/"
	// jsonMessage, _ := json.Marshal(message)
	// err = rabbitmq.Publish("AI_queue1", jsonMessage)
	// if err != nil {
	// 	logger.Error("Send to rabbit error: " + err.Error())
	// 	return
	// }
	// logger.Info("send rabbit")
	for {
		for _, element := range kvm.Recording_kvm {
			var message Message
			message.Hostname = element.Hostname
			// message.Path = "/home/media/image" + element.Hostname + "/"
			jsonMessage, _ := json.Marshal(message)
			rabbitmq.Publish("AI_queue1", jsonMessage)
		}
		time.Sleep(5 * time.Second)
	}
}
