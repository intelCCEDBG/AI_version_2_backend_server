package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"recorder/pkg/logger"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/rabbitmq"
	"time"

	"github.com/fsnotify/fsnotify"
)

var AI_list []string

type Message struct {
	Hostname string `json:"hostname"`
	Image    string `json:"image"`
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
			context := context.Background()
			go FS_monitor(element.Hostname, context)
		}
		time.Sleep(5 * time.Second)
	}
}

func FS_monitor(hostname string, ctx context.Context) {
	directory := "/home/media/image/" + hostname

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err.Error())
	}
	defer watcher.Close()

	err = watcher.Add(directory)
	if err != nil {
		logger.Error(err.Error())
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				filename := filepath.Base(event.Name)
				logger.Info("modified file:" + filename)
				if filename == hostname+".png" {
					err = Send_to_rabbitMQ(hostname)
					if err != nil {
						logger.Error(err.Error())
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Error(err.Error())
		}
	}
}

func Send_to_rabbitMQ(hostname string) (err error) {
	var message Message
	message.Hostname = hostname
	Path := "/home/media/image/" + hostname + "/" + hostname + ".png"
	imageFile, err := os.Open(Path)
	if err != nil {
		return err
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return err
	}
	message.Image = base64.StdEncoding.EncodeToString(imageData)
	jsonMessage, _ := json.Marshal(message)
	rabbitmq.Publish("AI_queue1", jsonMessage)
	imageFile.Close()
	return nil
}
