package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	ai "recorder/internal/AI"
	"recorder/internal/ffmpeg"
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/redis"
	"syscall"
	"time"
)

var Stop_channel map[string]context.CancelFunc
var Stop_signal_out_channel chan string
var Start_signal_out_channel chan string

func init() {
	Stop_channel = make(map[string]context.CancelFunc)
	Stop_signal_out_channel = make(chan string)
}

func Start_service() {
	Connection_close := make(chan int, 1)
	ctx, cancel := context.WithCancel(context.Background())
	// get_recording_kvm_back()
	go monitor_stop_signal()
	go monitor_start_signal()
	go monitor_error_signal()
	go monitor_stop_abnormal_signal()
	go ai.Start_ai_monitoring(ctx)
	go ai.FS_monitor_slow(ctx)
	Quit := make(chan os.Signal, 1)
	signal.Notify(Quit, syscall.SIGINT, syscall.SIGTERM)
	<-Quit
	fmt.Println("Server is shutting down...")
	cancel()
	servershutdown(Connection_close)
	select {
	case <-Connection_close:
		logger.Info("Connection closed")
	case <-time.After(40 * time.Second):
		logger.Info("Connection close fail, force shutdown after 40 seconds")
	}
	fmt.Println("Server shutdown complete.")
}
func monitor_stop_abnormal_signal() {
	for {
		hostname := <-Stop_signal_out_channel
		logger.Info(hostname + " stop abnormal recording")
		kvm.RecordtoError(hostname)
	}
}
func monitor_stop_signal() {
	for {
		keys := redis.RedisGetByPattern("kvm:*:stop")
		for _, key := range keys {
			hostname := redis.RedisGet(key)
			logger.Info(hostname + " stop recording")
			Stop_recording(hostname)
			kvm.RecordtoIdle(hostname)
			redis.RedisDel("kvm:" + hostname + ":stop")
		}
		time.Sleep(5 * time.Second)
	}
}
func monitor_error_signal() {
	for {
		keys := redis.RedisGetByPattern("kvm:*:error")
		for _, key := range keys {
			hostname := redis.RedisGet(key)
			logger.Info(hostname + " error occur while recording")
			Stop_recording(hostname)
			kvm.RecordtoError(hostname)
			redis.RedisDel("kvm:" + hostname + ":error")
		}
		time.Sleep(5 * time.Second)
	}
}
func monitor_start_signal() {
	for {
		keys := redis.RedisGetByPattern("kvm:*:recording")
		for _, key := range keys {
			var start_process = func() {
				hostname := redis.RedisGet(key)
				logger.Info(hostname + " start recording")
				ctx, cancel := context.WithCancel(context.Background())
				Stop_channel[hostname] = cancel
				Kvm := kvm_query.GetStatus(hostname)
				time.Sleep(5 * time.Second)
				go ffmpeg.Record(Stop_signal_out_channel, Kvm, ctx)
				go record_buffer(hostname)
				kvm.IdletoRecord(hostname)
				redis.RedisDel(key)
			}
			start_process()
		}
		// time.Sleep(2 * time.Second)
	}
}
func servershutdown(Connection_close chan int) {
	stop_recording_all()
	time.Sleep(30 * time.Second)
	Connection_close <- 1
}
func get_recording_kvm_back() {
	kvms := kvm_query.GetRecordingKvms()
	for _, element := range kvms {
		ctx, cancel := context.WithCancel(context.Background())
		Stop_channel[element.Hostname] = cancel
		go ffmpeg.Record(Stop_signal_out_channel, element, ctx)
	}
}

func record_buffer(hostname string) { //for machines who freeze since the start of the recording. Prevent the corruption of error video file.
	redis.RedisSet("kvm:"+hostname+":holding", hostname)
	time.Sleep(120 * time.Second)
	redis.RedisDel("kvm:" + hostname + ":holding")
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
		delete(Stop_channel, hostname)
	}
}
