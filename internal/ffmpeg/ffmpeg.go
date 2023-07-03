package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"recorder/config"
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	"syscall"
	"time"
)

func init() {
	if _, err := os.Stat("/usr/bin/ffmpeg"); os.IsNotExist(err) {
		fmt.Println("ffmpeg not found")
	}
}

func Record(ch chan<- string, mh *kvm.Kvm, ctx context.Context) {
	hostname := mh.Hostname
	url := mh.Stream_url
	video_path := config.Viper.GetString("RECORDING_PATH") + hostname + "/"
	image_path := config.Viper.GetString("IMAGE_PATH") + hostname + "/"
	if _, err := os.Stat(video_path); os.IsNotExist(err) {
		err := os.Mkdir(video_path, 0777)
		if err != nil{
			logger.Error(err.Error())
		}
		// TODO: handle error
	}
	if _, err := os.Stat(image_path); os.IsNotExist(err) {
		err := os.Mkdir(image_path, 0777)
		if err != nil{
			logger.Error(err.Error())
		}
		// TODO: handle error
	}
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-i", url,
		"-c", "copy", "-f", "segment", "-segment_time", "10", "-segment_list", video_path+"all.m3u8", video_path+"%07d.ts",
		"-vf", "fps=0.2", "-update",image_path+hostname+".png")
	logger.Info("ffmpeg "+"-loglevel "+"quiet "+ "-i "+ url+
	" -c "+ "copy "+ "-f "+ "segment "+ "-segment_time "+ "10 "+ "-segment_list "+ video_path+"all.m3u8 "+ video_path+"%07d.ts "+
	"-vf "+ "fps=1 "+"-update "+ image_path+hostname+".png")
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
