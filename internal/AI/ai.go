package ai

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"recorder/config"
	"recorder/internal/cropping"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	"recorder/pkg/rabbitmq"
	"strconv"
	"sync"
	"time"

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

var (
	mutex       sync.Mutex
	debounceMap = make(map[string]time.Time)
)

func debounceEvent(eventName string, duration time.Duration, action func()) {
	mutex.Lock()
	defer mutex.Unlock()

	if _, ok := debounceMap[eventName]; !ok {
		debounceMap[eventName] = time.Now()
		go func() {
			time.Sleep(duration)
			mutex.Lock()
			delete(debounceMap, eventName)
			mutex.Unlock()
			action()
		}()
	} else {
		debounceMap[eventName] = time.Now()
	}
}

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
				debounceEvent(hostname, 500*time.Millisecond, func() {
					Send_to_rabbitMQ(hostname, ramdisk_path+filename, "2000")
				})
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
func FS_monitor_slow(ctx context.Context) {
	ramdisk_path := config.Viper.GetString("ramdisk_slow_path")
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
				go Process_AI_result(hostname)
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
func Process_AI_result(hostname string) {
	Ai_result := dut_query.Get_AI_result(hostname)
	if Ai_result.Hostname == "null" {
		logger.Error("Machine " + hostname + " not found in database")
		return
	}
	slow_path := config.Viper.GetString("ramdisk_slow_path")
	cropping.Switch_picture_if_exist(slow_path + "/cropped/" + hostname + "_cropped.png")
	if Ai_result.Label == 0 {
		err := cropping.Crop_image(slow_path+hostname+".png", Ai_result.Coords, slow_path+"/cropped/"+hostname+"_cropped.png")
		if err != nil {
			logger.Error(err.Error())
		}
		dut_query.Update_dut_status(hostname, 0)
		dut_query.Update_dut_cnt(hostname, 0)
	} else {
		dut_info := dut_query.Get_dut_status(hostname)
		cropping.Crop_image(slow_path+hostname+".png", Ai_result.Coords, slow_path+"/cropped/"+hostname+"_cropped.png")
		//todo: if old file not exist
		if !fileoperation.FileExists(slow_path + "/cropped/" + hostname + "_cropped_old.png") {
			return
		}
		ssim_result := ssim_cal(slow_path+"/cropped/"+hostname+"_cropped.png", slow_path+"/cropped/"+hostname+"_cropped_old.png")
		if ssim_result < dut_info.Ssim {
			dut_query.Update_dut_cnt(hostname, dut_info.Cycle_cnt+1)
			dut_info.Cycle_cnt++
		} else {
			dut_query.Update_dut_cnt(hostname, 0)
		}
		if dut_info.Cycle_cnt >= dut_info.Threshhold {
			dut_query.Update_dut_status(hostname, 4)
		}
		logger.Info("SSIM result: " + strconv.FormatFloat(ssim_result, 'f', 6, 64))
	}
	if Ai_result.Label == 2 {
		//todo: handle restart type
	}

}
func ssim_cal(image1 string, image2 string) (ssim float64) {
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-i", image1, "-i", image2, "-lavfi", "ssim", "-f", "null", "-")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	cmd.Start()
	buf := bufio.NewReader(stdout)
	line, _, err := buf.ReadLine()
	if err != nil {
		logger.Error(err.Error())
	}
	ssim_str := string(line[len(line)-6:])
	ssim, err = strconv.ParseFloat(ssim_str, 64)
	if err != nil {
		logger.Error(err.Error())
	}
	return ssim
}
func Send_to_rabbitMQ(hostname string, path string, expire_time string) (err error) {
	var message Message
	message.Hostname = hostname
	time.Sleep(100 * time.Millisecond)
	logger.Info(path)
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
