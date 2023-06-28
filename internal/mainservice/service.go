package service

import (
	"context"
	"recorder/internal/ffmpeg"
	"recorder/internal/machine"
	"recorder/pkg/logger"
)

var Stop_channel map[string]context.CancelFunc
var Stop_signal_out_channel chan string

func Start_service() {
	get_recording_machine_back()
	go monitor_stop_signal()

}
func monitor_stop_signal() {
	for {
		hostname := <-Stop_signal_out_channel
		logger.Info(hostname + " stop recording")
		machine.RecordtoIdle(hostname)
	}
}

func get_recording_machine_back() {
	for _, element := range machine.Recording_machine {
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
