package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"recorder/config"
	"recorder/pkg/logger"
	"recorder/pkg/rabbitmq"

	"github.com/fsnotify/fsnotify"
)

var AI_list []string

type Message struct {
	Hostname string `json:"hostname"`
	Image    string `json:"image"`
	// Path     string `json:"path"`
}

func Start_ai_monitoring(ctx context.Context) {
	_, err := rabbitmq.Declare("AI_queue1")
	if err != nil {
		logger.Error("Declare to rabbit error: " + err.Error())
		return
	}
	go FS_monitor_ramdisk(ctx)

	<-ctx.Done()

}

//	func FS_monitor(hostname string, ctx context.Context) {
//		directory := "/home/media/image/" + hostname
//		watcher, err := fsnotify.NewWatcher()
//		if err != nil {
//			logger.Error(err.Error())
//		}
//		defer watcher.Close()
//		err = watcher.Add(directory)
//		if err != nil {
//			logger.Error(err.Error())
//		}
//			for {
//				select {
//				case event, ok := <-watcher.Events:
//					if !ok {
//						return
//					}
//					if event.Op&fsnotify.Write == fsnotify.Write {
//						filename := filepath.Base(event.Name)
//						logger.Info("modified file:" + filename)
//						if filename == hostname+".png" {
//							err = Send_to_rabbitMQ(hostname)
//							if err != nil {
//								logger.Error(err.Error())
//							}
//						}
//					}
//				case err, ok := <-watcher.Errors:
//					if !ok {
//						return
//					}
//					logger.Error(err.Error())
//				}
//			}
//		}
func FS_monitor_ramdisk(ctx context.Context) {
	ramdisk_path := config.Viper.GetString("ramdisk_path")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err.Error())
	}
	defer watcher.Close()

	err = watcher.Add(ramdisk_path)
	if err != nil {
		logger.Error(err.Error())
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			// logger.Info("Get event!")
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				filename := filepath.Base(event.Name)
				// logger.Info("modified file:" + filename)
				hostname := filename[:len(filename)-4]
				go Send_to_rabbitMQ(hostname, ramdisk_path+filename, "2000")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Error(err.Error())
		case <-ctx.Done():
			return
		}
	}
}
func Send_to_rabbitMQ(hostname string, path string, expire_time string) (err error) {
	var message Message
	message.Hostname = hostname
	imageFile, err := os.Open(path)
	if err != nil {
		return err
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return err
	}
	message.Image = base64.StdEncoding.EncodeToString(imageData)
	jsonMessage, _ := json.Marshal(message)
	rabbitmq.Publish_with_expiration("AI_queue1", jsonMessage, expire_time)
	imageFile.Close()
	return nil
}
