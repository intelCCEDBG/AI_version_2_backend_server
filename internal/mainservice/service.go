package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"recorder/internal/ffmpeg"
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	"syscall"
	"time"
)

var Stop_channel map[string]context.CancelFunc
var Stop_signal_out_channel chan string

func init() {
	Stop_channel = make(map[string]context.CancelFunc)
}

func Start_service() {
	Connection_close := make(chan int, 1)
	get_recording_kvm_back()
	go monitor_stop_signal()
	Quit := make(chan os.Signal, 1)
	signal.Notify(Quit, syscall.SIGINT, syscall.SIGTERM)
	<-Quit
	fmt.Println("Server is shutting down...")
	servershutdown()
	select {
	case <-Connection_close:
		logger.Info("Connection closed")
	case <-time.After(5 * time.Second):
		logger.Info("Connection close fail, force shutdown after 5 seconds")
	}
	fmt.Println("Server shutdown complete.")
}
func monitor_stop_signal() {
	for {
		hostname := <-Stop_signal_out_channel
		logger.Info(hostname + " stop recording")
		kvm.RecordtoIdle(hostname)
	}
}
func servershutdown() {
	stop_recording_all()

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

func stop_recording(hostname string) {
	Stop_channel[hostname]()
}
