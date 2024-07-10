package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"os"
	"recorder/config"
	"recorder/internal/cropping"
	emailfunction "recorder/internal/email_function"
	"recorder/internal/ffmpeg"
	"recorder/internal/logpicqueue"
	"recorder/internal/ssim"
	"recorder/internal/structure"
	videogen "recorder/internal/video_gen"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/rabbitmq"
	"recorder/pkg/redis"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var AI_list []string

type Message struct {
	Hostname     string    `json:"hostname"`
	Machine_name string    `json:"machine_name"`
	Image        string    `json:"image"`
	Coord        []float64 `json:"coord"`
	Locked       int       `json:"locked"`
	// Path     string `json:"path"`
}

func Start_ai_monitoring(ctx context.Context) {
	_, err := rabbitmq.Declare("AI_queue1")
	if err != nil {
		logger.Error("Declare to rabbit error: " + err.Error())
		return
	}
	go FS_monitor_ramdisk(ctx)

	<-ctx.Done()

}

func Process_AI_result(hostname string, machine_name string) {
	// unit := unit_query.Get_unitbyhostname(hostname)
	// sta := dut_query.Get_dut_status(unit.Machine_name)
	key := redis.Redis_get_by_pattern("kvm:" + hostname + ":holding")
	if len(key) != 0 {
		return
	}
	KVM := kvm_query.Get_kvm_status(hostname)
	Ai_result := dut_query.Get_AI_result(machine_name)
	if Ai_result.Hostname == "null" {
		logger.Debug("Machine " + machine_name + " not found in database. ") //This may happens when camera did not catch anything.
		return
	}
	if len(Ai_result.Coords) == 0 {
		return
	}
	slow_path := config.Viper.GetString("slow_path")
	cropped_path := config.Viper.GetString("cropped_path")
	cropping.Switch_picture_if_exist(cropped_path + hostname + "_cropped.png")
	var cropped_image image.Image
	var err error
	cropped_image, err = cropping.Crop_image(slow_path+hostname+".png", Ai_result.Coords, cropped_path+hostname+"_cropped.png")
	if err != nil {
		logger.Error(err.Error())
	}
	err = logpicqueue.SendtoLogPicChannel(machine_name, cropped_image)
	if err != nil {
		logger.Error(err.Error())
	}
	// if Ai_result.Label == 0 {
	// 	dut_query.Update_dut_status(machine_name, 0)
	// 	dut_query.Update_dut_cnt(machine_name, 0)
	// } else {
	if !fileoperation.FileExists(cropped_path + hostname + "_cropped_old.png") {
		return
	}
	ssim_result, err := ssim.Ssim_cal(cropped_path+hostname+"_cropped.png", cropped_path+hostname+"_cropped_old.png")

	if err != nil {
		logger.Error(err.Error())
		return
	}
	dut_info := dut_query.Get_dut_status(machine_name)
	if ssim_result*100 >= dut_info.Ssim {
		// logger.Info("Freeze: " + hostname)
		dut_query.Update_dut_cnt(machine_name, dut_info.Cycle_cnt+1)
		dut_info.Cycle_cnt++
	} else {
		dut_query.Update_dut_cnt(machine_name, 1)
		dut_query.Update_dut_status(hostname, 4)
	}
	if dut_info.Cycle_cnt == dut_info.Threshhold*12 {
		// dut_query.Update_dut_status(hostname, 4)
		freeze_process(machine_name, Ai_result.Label, KVM, dut_info.Threshhold)
	}
	// logger.Debug("SSIM result: " + strconv.FormatFloat(ssim_result, 'f', 6, 64))
	// }
	if Ai_result.Label == 2 {
		//todo: handle restart type
	}
}
func freeze_process(machine_name string, errortype int, kvm structure.Kvm, threshold int) {
	logger.Info("Machine " + machine_name + " Fail !")
	Machine_status := dut_query.Get_machine_status(machine_name)
	error_record := create_new_error_record(machine_name, strconv.Itoa(errortype), Machine_status)
	currentTime := time.Now()
	// threshold*=-1
	freezetime := currentTime.Add(time.Duration((threshold+2)*-1) * time.Minute)
	logger.Info(freezetime.String())
	videogen.GenerateErrorVideo(freezetime.Hour(), freezetime.Minute(), 180, kvm.Hostname, machine_name, error_record.Uuid)
	copyFileFromQueue(machine_name)
	current_picture_path := config.Viper.GetString("logimage_path") + machine_name + "/current.png"
	ffmpeg.Take_photo(current_picture_path, kvm)
	project := dut_query.Get_project_code(machine_name)
	dut_query.Set_dut_status_from_kvm(errortype, kvm)
	emailfunction.Send_alert_mail(machine_name, project, errortype)
}
func copyFileFromQueue(machine_name string) {
	logpicqueue.BlockLogPicChannel(machine_name)
	defer logpicqueue.UnblockLogPicChannel(machine_name)
	var index = 1
	for {
		image := logpicqueue.GetChannelContent(machine_name)
		if image == nil {
			break
		}
		fileoperation.CreateFolderifNotExist(config.Viper.GetString("logimage_path") + machine_name)
		outputImagePath := config.Viper.GetString("logimage_path") + machine_name + "/" + strconv.Itoa(index) + ".png"
		outputFile, err := os.Create(outputImagePath)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		err = png.Encode(outputFile, image)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		outputFile.Close()
		index++
		if index == 4 {
			break
		}
	}

}
func Send_to_rabbitMQ(hostname string, machine_name string, locked string, path string, expire_time string) (err error) {
	var message Message
	message.Hostname = hostname
	message.Machine_name = machine_name
	message.Locked = 0
	if locked != "" {
		message.Locked = 1
		message.Coord = structure.Coord_s2f(locked)
	}
	time.Sleep(100 * time.Millisecond)
	// logger.Info(path)
	imageFile, err := os.Open(path)
	if err != nil {
		return err
	}
	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		return err
	}
	message.Image = base64.StdEncoding.EncodeToString(imageData)
	jsonMessage, _ := json.Marshal(message)
	rabbitmq.Publish_with_expiration("AI_queue1", jsonMessage, expire_time)
	imageFile.Close()
	return nil
}
func create_new_error_record(machine_name string, errortype string, machine_status structure.Machine_status) structure.Errorlog {
	var errorlog structure.Errorlog
	errorlog.Machine_name = machine_name
	errorlog.Time = time.Now().Format("2006-01-02 15:04:05")
	errorlog.Type = errortype
	errorlog.Test_item = machine_status.Test_item
	errorlog.Sku = machine_status.Sku
	errorlog.Image = machine_status.Image
	errorlog.Bios = machine_status.Bios
	errorlog.Uuid = uuid.New().String()
	errorlog_query.Set_error_record(errorlog)
	return errorlog
}
