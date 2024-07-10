package ai

import (
	"context"
	"path/filepath"
	"recorder/config"
	"recorder/pkg/logger"
	unit_query "recorder/pkg/mariadb/unit"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	slowmutex       sync.Mutex
	slowdebounceMap = make(map[string]time.Time)
)

func slowdebounceEvent(eventName string, duration time.Duration, action func()) {
	slowmutex.Lock()
	defer slowmutex.Unlock()
	if _, ok := slowdebounceMap[eventName]; !ok {
		slowdebounceMap[eventName] = time.Now()
		go func() {
			time.Sleep(duration)
			slowmutex.Lock()
			delete(slowdebounceMap, eventName)
			slowmutex.Unlock()
			action()
		}()
	} else {
		slowdebounceMap[eventName] = time.Now()
	}
}

func FS_monitor_slow(ctx context.Context) {
	ramdisk_path := config.Viper.GetString("slow_path")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err.Error())
	}
	defer watcher.Close()

	err = watcher.Add(ramdisk_path)
	if err != nil {
		logger.Error(err.Error())
	}
	logger.Info("Start monitoring slow path!")
	for {
		select {
		case event, ok := <-watcher.Events:
			// logger.Info("Get event!")
			if !ok {
				logger.Error(err.Error())
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				filename := filepath.Base(event.Name)
				// logger.Info("modified file:" + filename)
				hostname := filename[:len(filename)-4]
				unit := unit_query.Get_unitbyhostname(hostname)
				slowdebounceEvent(hostname, 1200*time.Millisecond, func() {
					Process_AI_result(hostname, unit.Machine_name)
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				logger.Error(err.Error())
				return
			}
			logger.Error(err.Error())
		case <-ctx.Done():
			return
		}
	}
}
