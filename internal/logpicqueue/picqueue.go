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
var mapsmutex *sync.Mutex
var channelmapsmutex *sync.Mutex
var usedmutex *sync.Mutex

func init() {

	LogPicChannel_channel = make(map[string]*LogPicChannel)
	mutexmap = make(map[string]*sync.Mutex)
	mutexused = make(map[string]bool)
	mapsmutex = &sync.Mutex{}
	usedmutex = &sync.Mutex{}
	channelmapsmutex = &sync.Mutex{}
}

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
	mapsmutex.Lock()
	defer mapsmutex.Unlock()
	channelmapsmutex.Lock()
	defer channelmapsmutex.Unlock()
	LogPicChannel_channel[key] = NewLogPicChannel(maxSize)
	mutexmap[key] = &sync.Mutex{}
}

func SendtoLogPicChannel(key string, value image.Image) error {
	// This function may cause picture misorder
	usedmutex.Lock()
	if mutexused[key] {
		return nil
	}
	usedmutex.Unlock()
	channel := GetLogPicChannel(key)
	if channel == nil {
		dut := dut_query.Get_dut_status(key)
		if dut.Machine_name == "null" {
			return errors.New("machine not found")
		}
		amount, _ := strconv.Atoi(config.Viper.GetString("log_img_amount"))
		amount = amount/2 + 1
		AddNewLogPicChannel(key, (dut.Threshhold*12)+amount)
	}
	channel = GetLogPicChannel(key)
	BlockLogPicChannel(key)
	defer UnblockLogPicChannel(key)
	channel.Send(value)
	return nil
}

func GetLogPicChannel(key string) *LogPicChannel {
	channelmapsmutex.Lock()
	defer channelmapsmutex.Unlock()
	return LogPicChannel_channel[key]
}

func GetChannelContent(key string) image.Image {
	channel := GetLogPicChannel(key)
	select {
	case value := <-channel.ch:
		return value
	default:
		return nil
	}
}

func DeleteLogPicChannel(key string) {
	channelmapsmutex.Lock()
	defer channelmapsmutex.Unlock()
	delete(LogPicChannel_channel, key)
}

func BlockLogPicChannel(key string) {
	mapsmutex.Lock()
	defer mapsmutex.Unlock()
	usedmutex.Lock()
	mutexmap[key].Lock()
	mutexused[key] = true
	usedmutex.Unlock()
}

func UnblockLogPicChannel(key string) {
	mapsmutex.Lock()
	defer mapsmutex.Unlock()
	mutexmap[key].Unlock()
	usedmutex.Lock()
	mutexused[key] = false
	usedmutex.Unlock()
}

func RenewThreshold(key string) error {
	BlockLogPicChannel(key)
	dut := dut_query.Get_dut_status(key)
	if dut.Machine_name == "null" {
		return errors.New("machine not found")
	}
	amount, _ := strconv.Atoi(config.Viper.GetString("log_img_amount"))
	amount = amount/2 + 1
	channelmapsmutex.Lock()
	defer channelmapsmutex.Unlock()
	LogPicChannel_channel[key] = NewLogPicChannel(dut.Threshhold*12 + amount)
	UnblockLogPicChannel(key)
	return nil
}

func ErrorImageOutput(key string) []image.Image {
	var rt []image.Image
	BlockLogPicChannel(key)
	defer UnblockLogPicChannel(key)
	for i := 1; i <= 3; i++ {
		rt = append(rt, GetChannelContent(key))
	}
	return rt
}
