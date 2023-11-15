package ai

import (
	"encoding/json"
	"recorder/pkg/logger"
	kvm_query "recorder/pkg/mariadb/kvm"
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
	for {
		kvms := kvm_query.Get_recording_kvms()
		for _, element := range kvms {
			var message Message
			message.Hostname = element.Hostname
			// message.Path = "/home/media/image" + element.Hostname + "/"
			jsonMessage, _ := json.Marshal(message)
			rabbitmq.Publish("AI_queue1", jsonMessage)
		}
		time.Sleep(5 * time.Second)
	}
}
