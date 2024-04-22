package ai

import (
	"context"
	"path/filepath"
	"recorder/config"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	unit_query "recorder/pkg/mariadb/unit"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

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
				unit := unit_query.Get_unitbyhostname(hostname)
				sta := dut_query.Get_dut_status(unit.Machine_name)
				debounceEvent(hostname, 500*time.Millisecond, func() {
					Send_to_rabbitMQ(unit.Hostname, unit.Machine_name, sta.Lock_coord, ramdisk_path+filename, "2000")
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
