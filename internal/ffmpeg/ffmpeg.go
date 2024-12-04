package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"recorder/config"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	"syscall"
	"time"
)

func init() {
	if _, err := os.Stat("/usr/bin/ffmpeg"); os.IsNotExist(err) {
		fmt.Println("ffmpeg not found")
	}
}

func Record(ch chan string, mh structure.Kvm, ctx context.Context) {
	hostname := mh.Hostname
	url := mh.StreamUrl
	video_path := config.Viper.GetString("RECORDING_PATH") + hostname + "/"
	image_path := config.Viper.GetString("IMAGE_PATH") + hostname + "/"
	ramdisk_path := config.Viper.GetString("RAMDISK_PATH")
	slow_path := config.Viper.GetString("SLOW_PATH")
	err := os.RemoveAll(video_path)
	if err != nil {
		logger.Error(err.Error())
	}
	if _, err := os.Stat(video_path); os.IsNotExist(err) {
		syscall.Umask(0)
		err := os.Mkdir(video_path, 0777)
		if err != nil {
			logger.Error(err.Error())
		}
	}
	err = os.RemoveAll(image_path)
	if err != nil {
		logger.Error(err.Error())
	}
	if _, err := os.Stat(image_path); os.IsNotExist(err) {
		syscall.Umask(0)
		err := os.Mkdir(image_path, 0777)
		if err != nil {
			logger.Error(err.Error())
		}
	}
	dut_query.SetDutStatusFromKvm(4, mh)
	reso := "scale=320:180"
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-r", "3", "-y", "-i", url,
		"-codec", "libx264", "-preset", "ultrafast", "-g", "30", "-f", "hls", "-hls_time", "10", "-hls_list_size", "0", "-strftime", "1", "-hls_segment_filename", video_path+"%Y-%m-%d_%H-%M-%S.ts", video_path+"all.m3u8",
		"-r", "0.3", "-update", "1", ramdisk_path+hostname+".png", "-r", "0.2", "-update", "1", slow_path+hostname+".png", "-vf", reso, "-r", "1", "-update", "1", image_path+hostname+"_low.png")
	cmd2 := exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", url, "-vframes", "1", image_path+"cover.png")
	in, err := cmd.StdinPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		logger.Error(err.Error())
	}
	err = cmd.Start()
	cmd2.Start()
	mh.StartRecordTime = time.Now().Unix()
	if err != nil {
		logger.Error(err.Error())
	}
	cmdDone := make(chan error)
	go func() {
		cmdDone <- cmd.Wait()
	}()
	select {
	case <-ctx.Done():
		{
			fmt.Println("send exit signal")
			io.WriteString(in, "q")
			if err != nil {
				logger.Error(err.Error())
			}
		}
	case err := <-cmdDone:
		if err != nil {
			logger.Error(err.Error())
		}
	}
	ch <- hostname
}

func Get_picture_from_video(ts string, frame_index string, image_path string) error {
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", ts, "-vf", "select='eq(n\\,"+frame_index+")'", "-vsync", "vfr", image_path)
	err := cmd.Run()
	if err != nil {
		logger.Error(err.Error())
	}
	return err
}

func TakePhoto(save_path string, target structure.Kvm) error {
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", target.StreamUrl, "-vframes", "1", save_path)
	err := cmd.Run()
	if err != nil {
		logger.Error(err.Error())
	}
	return err
}

func RenewSmallPicture(target structure.Kvm) error {
	image_path := config.Viper.GetString("IMAGE_PATH") + target.Hostname + "/"
	cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", target.StreamUrl, "-vf", "scale=320:180", "-vframes", "1", image_path+"cover.png")
	err := cmd.Run()
	if err != nil {
		logger.Error(err.Error())
	}
	return err
}
