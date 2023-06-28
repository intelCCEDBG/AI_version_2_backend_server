package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"recorder/config"
	"recorder/internal/machine"
	"recorder/pkg/logger"
	"syscall"
	"time"
)

func init() {
	if _, err := os.Stat("/usr/bin/ffmpeg"); os.IsNotExist(err) {
		fmt.Println("ffmpeg not found")
	}
}

func Record(ch chan<- string, mh *machine.Machine, ctx context.Context) {
	hostname := mh.Hostname
	url := mh.Stream_url + "&localaddr=" + mh.Stream_interface
	video_path := config.Viper.GetString("RECORDING_PATH") + hostname + "/"
	image_path := config.Viper.GetString("IMAGE_PATH") + hostname + "/"
	cmd := exec.Command("ffmpeg", "-loglevel", "quite", "-i", url,
		"-c", "copy", "-f", "segment", "-segment_time", "10", "-segment_list", video_path+"%07d.m3u8", video_path+"%07d.ts",
		"-vf", "fps=1", image_path+hostname+".png")
	_, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	err = cmd.Start()
	mh.Start_record_time = time.Now().Unix()
	if err != nil {
		logger.Error(err.Error())
	}
	cmdDone := make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()
	select {
	case <-ctx.Done():
		err := cmd.Process.Signal(syscall.SIGINT)
		if err != nil {
			logger.Error(err.Error())
		}
	case err := <-cmdDone:
		if err != nil {
			logger.Error(err.Error())
		}
	}
	ch <- hostname
}
