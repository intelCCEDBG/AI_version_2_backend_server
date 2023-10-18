package ai

import (
	"encoding/json"
	"recorder/internal/kvm"
	"recorder/pkg/rabbitmq"
	"time"
)

var AI_list []string

type Message struct {
	Hostname string `json:"hostname"`
	// Path     string `json:"path"`
}

func Start_ai_monitoring() {
	rabbitmq.Declare("AI_queue")
	for {
		for _, element := range kvm.Recording_kvm {
			var message Message
			message.Hostname = element.Hostname
			// message.Path = "/home/media/image" + element.Hostname + "/"
			jsonMessage, _ := json.Marshal(message)
			rabbitmq.Publish("AI_queue", jsonMessage)
		}
		time.Sleep(5 * time.Second)
	}
}
