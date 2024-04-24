package logpicqueue

import (
	"errors"
	"image"
	"recorder/config"
	dut_query "recorder/pkg/mariadb/dut"
	"strconv"
	"sync"
)

type LogPicChannel struct {
	ch      chan image.Image
	maxSize int
}

var LogPicChannel_channel map[string]*LogPicChannel
var mutexmap map[string]*sync.Mutex
var mutexused map[string]bool

func NewLogPicChannel(maxSize int) *LogPicChannel {
	return &LogPicChannel{
		ch:      make(chan image.Image, maxSize),
		maxSize: maxSize,
	}
}

func (lc *LogPicChannel) Send(value image.Image) {
	select {
	case lc.ch <- value:
		// Value successfully sent to the channel
	default:
		// Channel is full, drop the oldest value
		select {
		case <-lc.ch: // Remove the oldest value from the channel
		default:
		}
		// Retry sending the new value
		lc.ch <- value
	}
}

func AddNewLogPicChannel(key string, maxSize int) {
	LogPicChannel_channel[key] = NewLogPicChannel(maxSize)
	mutexmap[key] = &sync.Mutex{}
}

func SendtoLogPicChannel(key string, value image.Image) error {
	// This function may cause picture misorder
	if mutexused[key] {
		return nil
	}
	BlockLogPicChannel(key)
	defer UnblockLogPicChannel(key)
	if LogPicChannel_channel[key] == nil {
		dut := dut_query.Get_dut_status(key)
		if dut.Machine_name == "null" {
			return errors.New("machine not found")
		}
		amount, _ := strconv.Atoi(config.Viper.GetString("log_img_amount"))
		amount = amount/2 + 1
		AddNewLogPicChannel(key, dut.Threshhold*12+amount)
	}
	LogPicChannel_channel[key].Send(value)
	return nil
}

func GetLogPicChannel(key string) *LogPicChannel {
	return LogPicChannel_channel[key]
}

func GetChannelContent(key string) image.Image {
	select {
	case value := <-LogPicChannel_channel[key].ch:
		return value
	default:
		return nil
	}
}

func DeleteLogPicChannel(key string) {
	delete(LogPicChannel_channel, key)
}

func BlockLogPicChannel(key string) {
	mutexmap[key].Lock()
	mutexused[key] = true
}

func UnblockLogPicChannel(key string) {
	mutexmap[key].Unlock()
	mutexused[key] = false
}

func RenewThreshold(key string) error {
	BlockLogPicChannel(key)
	dut := dut_query.Get_dut_status(key)
	if dut.Machine_name == "null" {
		return errors.New("machine not found")
	}
	amount, _ := strconv.Atoi(config.Viper.GetString("log_img_amount"))
	amount = amount/2 + 1
	LogPicChannel_channel[key] = NewLogPicChannel(dut.Threshhold*12 + amount)
	UnblockLogPicChannel(key)
	return nil
}
