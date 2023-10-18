package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"recorder/internal/ffmpeg"
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	"recorder/pkg/redis"
	"syscall"
	"time"
)

var Stop_channel map[string]context.CancelFunc
var Stop_signal_out_channel chan string
var Start_signal_out_channel chan string

func init() {
	Stop_channel = make(map[string]context.CancelFunc)
}

func Start_service() {
	Connection_close := make(chan int, 1)
	get_recording_kvm_back()
	go monitor_stop_signal()
	go monitor_start_signal()
	go monitor_stop_abnormal_signal()
	Quit := make(chan os.Signal, 1)
	signal.Notify(Quit, syscall.SIGINT, syscall.SIGTERM)
	<-Quit
	fmt.Println("Server is shutting down...")
	servershutdown(Connection_close)
	select {
	case <-Connection_close:
		logger.Info("Connection closed")
	case <-time.After(5 * time.Second):
		logger.Info("Connection close fail, force shutdown after 5 seconds")
	}
	fmt.Println("Server shutdown complete.")
}
func monitor_stop_abnormal_signal() {
	for {
		hostname := <-Stop_signal_out_channel
		logger.Info(hostname + " stop abnormal recording")
		kvm.RecordtoIdle(hostname)
	}
}
func monitor_stop_signal() {
	for {
		keys := redis.Redis_get_by_pattern("kvm:*:stop")
		if keys != nil {
			for _, key := range keys {
				hostname := redis.Redis_get(key)
				logger.Info(hostname + " stop recording")
				Stop_recording(hostname)
				kvm.RecordtoIdle(hostname)
				redis.Redis_del("kvm:" + hostname + ":stop")
			}
		}
		time.Sleep(5 * time.Second)
	}
}
func monitor_start_signal() {
	for {
		keys := redis.Redis_get_by_pattern("kvm:*:recording")
		for _, key := range keys {
			var start_process = func() {
				hostname := redis.Redis_get(key)
				logger.Info(hostname + " start recording")
				ctx, cancel := context.WithCancel(context.Background())
				Stop_channel[hostname] = cancel
				Kvm := kvm.Get(hostname)
				time.Sleep(5 * time.Second)
				go ffmpeg.Record(Stop_signal_out_channel, &Kvm, ctx)
				kvm.IdletoRecord(hostname)
				redis.Redis_del("kvm:" + hostname + ":recording")
			}
			start_process()
		}
		time.Sleep(5 * time.Second)
	}
}
func servershutdown(Connection_close chan int) {
	stop_recording_all()
	time.Sleep(1 * time.Second)
	Connection_close <- 1
}
func get_recording_kvm_back() {
	for _, element := range kvm.Recording_kvm {
		ctx, cancel := context.WithCancel(context.Background())
		Stop_channel[element.Hostname] = cancel
		go ffmpeg.Record(Stop_signal_out_channel, element, ctx)
	}
}

func stop_recording_all() {
	for _, element := range Stop_channel {
		element()
	}
}

func Stop_recording(hostname string) {
	_, ok := Stop_channel[hostname]
	if ok {
		Stop_channel[hostname]()
	}
}
