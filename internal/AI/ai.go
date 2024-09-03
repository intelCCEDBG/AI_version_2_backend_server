package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"recorder/config"
	"recorder/internal/cropping"
	"recorder/internal/ffmpeg"
	"recorder/internal/logpicqueue"
	"recorder/internal/ssim"
	"recorder/internal/structure"
	videogen "recorder/internal/video_gen"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	email_query "recorder/pkg/mariadb/email"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
	kvm_query "recorder/pkg/mariadb/kvm"
	project_query "recorder/pkg/mariadb/project"
	unit_query "recorder/pkg/mariadb/unit"
	"recorder/pkg/rabbitmq"
	"recorder/pkg/redis"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var AI_list []string

type Message struct {
	Hostname    string    `json:"hostname"`
	MachineName string    `json:"machine_name"`
	Image       string    `json:"image"`
	Coord       []float64 `json:"coord"`
	Locked      int       `json:"locked"`
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

func ProcessAIResult(hostname string, machineName string) {
	// see if the kvm is holding
	key := redis.RedisGetByPattern("kvm:" + hostname + ":holding")
	if len(key) != 0 {
		dut_query.UpdateDutCycleCnt(machineName, 0)
		dut_query.UpdateDutCycleCntHigh(machineName, 0)
		dut_query.UpdateDutStatus(machineName, 4)
		return
	}
	// see if the kvm is checking (pressing windows key)
	key = redis.RedisGetByPattern("kvm:" + hostname + ":checking")
	if len(key) != 0 {
		return
	}
	KVM := kvm_query.GetKvmStatus(hostname)
	AIResult := dut_query.GetAiResult(machineName)
	if AIResult.Hostname == "null" { // if no hostname
		logger.Debug("Machine " + machineName + " not found in database. ") //This may happen when camera did not catch anything.
		return
	}
	if len(AIResult.Coords) == 0 { // if unlocked
		dut_query.UpdateDutStatus(machineName, 4)
		return
	}
	slowPath := config.Viper.GetString("slow_path")
	croppedPath := config.Viper.GetString("cropped_path")
	cropping.SwitchPictureIfExist(croppedPath + hostname + "_cropped.png")
	var croppedImage image.Image
	var err error
	croppedImage, err = cropping.CropImage(slowPath+hostname+".png", AIResult.Coords, croppedPath+hostname+"_cropped.png")
	if err != nil {
		logger.Error(err.Error())
	}
	err = logpicqueue.SendtoLogPicChannel(machineName, croppedImage)
	if err != nil {
		logger.Error(err.Error())
	}
	if !fileoperation.FileExists(croppedPath + hostname + "_cropped_old.png") {
		return
	}
	ssimResult, err := ssim.SsimCal(croppedPath+hostname+"_cropped.png", croppedPath+hostname+"_cropped_old.png")
	if err != nil {
		logger.Error(err.Error())
		return
	}
	dutInfo := dut_query.GetDutStatus(machineName)
	if len(dutInfo.LockCoord) == 0 {
		return
	}
	unit := unit_query.GetUnitByHostname(hostname)

	// update cycle count high (for cursor flash detection)
	upperBound := project_query.GetUpperBound(unit.Project)
	if ssimResult*100 >= float64(upperBound) {
		dut_query.UpdateDutCycleCntHigh(machineName, dutInfo.CycleCntHigh+1)
		dutInfo.CycleCntHigh++
	} else {
		dutInfo.CycleCntHigh -= 3
		if (dutInfo.CycleCntHigh) < 0 {
			dutInfo.CycleCntHigh = 0
		}
		dut_query.UpdateDutCycleCntHigh(machineName, dutInfo.CycleCntHigh)
	}
	// update cycle count (for normal freeze)
	if ssimResult*100 >= dutInfo.Ssim {
		dut_query.UpdateDutCycleCnt(machineName, dutInfo.CycleCnt+1)
		dutInfo.CycleCnt++
	} else {
		dut_query.UpdateDutCycleCnt(machineName, 0)
		dut_query.UpdateDutCycleCntHigh(machineName, 0)
		return
	}

	// Label 0: BSOD 1: BLACK 2: RESTART 3: NORMAL
	// Status 0: BSOD 1: BLACK 2: RESTART 3: FREEZE 4: NORMAL
	switch AIResult.Label {
	case structure.BSOD_LABEL:
		dut_query.UpdateDutStatus(machineName, structure.BSOD)
		if dutInfo.CycleCnt == dutInfo.Threshhold*12 {
			logger.Debug("Machine " + machineName + " Freeze Detected, Status: BSOD")
			freezeProcess(machineName, AIResult.Label, KVM, dutInfo.Threshhold)
		}
	case structure.BLACK_LABEL:
		dut_query.UpdateDutStatus(machineName, structure.BLACK)
		if dutInfo.CycleCnt == dutInfo.Threshhold*12 {
			logger.Debug("Machine " + machineName + " Freeze Detected, Status: BLACK")
			freezeProcess(machineName, AIResult.Label, KVM, dutInfo.Threshhold)
		}
	case structure.RESTART_LABEL:
		dut_query.UpdateDutStatus(machineName, structure.NORMAL)
		if dutInfo.CycleCnt == dutInfo.Threshhold*12 {
			logger.Debug("Machine " + machineName + " Freeze Detected, Status: RESTART")
			dut_query.UpdateDutStatus(machineName, structure.FREEZE)
			freezeProcess(machineName, structure.FREEZE, KVM, dutInfo.Threshhold)
		}
	case structure.NORMAL_LABEL:
		dut_query.UpdateDutStatus(machineName, structure.NORMAL)
		if dutInfo.CycleCnt == dutInfo.Threshhold*12 && dutInfo.CycleCntHigh > dutInfo.Threshhold*10 {
			logger.Warn("Machine " + machineName + " Windows Key Freeze Check!")
			freezed, err := FreezeCheck(machineName, AIResult.Coords, KVM)
			if err != nil {
				logger.Error("Error doing windows key freeze check: " + err.Error())
			}
			if freezed {
				logger.Debug("Machine " + machineName + " Freeze Detected, Status: NORMAL")
				dut_query.UpdateDutStatus(machineName, structure.FREEZE)
				freezeProcess(machineName, AIResult.Label, KVM, dutInfo.Threshhold)
			} else {
				logger.Warn("Machine " + machineName + " Recovered From Press Windows Key")
				dut_query.UpdateDutCycleCnt(machineName, 0)
				dut_query.UpdateDutCycleCntHigh(machineName, 0)
			}
		}
	}
}

func freezeProcess(machineName string, errortype int, kvm structure.Kvm, threshold int) {
	logger.Info("Machine " + machineName + " Failed!")
	machineStatus := dut_query.GetMachineStatus(machineName)
	errorRecord := createNewErrorRecord(machineName, strconv.Itoa(errortype), machineStatus)
	currentTime := time.Now()
	freezetime := currentTime.Add(time.Duration((threshold+2)*-1) * time.Minute)
	videogen.GenerateErrorVideo(freezetime.Hour(), freezetime.Minute(), 180, kvm.Hostname, machineName, errorRecord.Uuid)
	copyFileFromQueue(machineName)
	currentPicturePath := config.Viper.GetString("logimage_path") + machineName + "/current.png"
	ffmpeg.TakePhoto(currentPicturePath, kvm)
	code := dut_query.GetProjectCode(machineName)
	project := dut_query.GetProjectName(machineName)
	ffmpeg.RenewSmallPicture(kvm)
	SendAlertMail(machineName, code, errortype, project)
}

func SendAlertMail(machineName string, code string, errortype int, project string) {
	Ip := config.Viper.GetString("email_server_ip")
	Port := config.Viper.GetString("email_port")
	startTime := project_query.GetStartTime(project)
	status := email_query.GetEmailThreestrikeStatus(code)
	times := errorlog_query.GetErrorCountAfterTime(machineName, startTime)
	if (status == 1 && times < 4) || status == 0 {
		_, err := http.Get("http://" + Ip + ":" + Port + "/api/mail/send_alert?project=" + code + "&machine_name=" + machineName + "&type=" + strconv.Itoa(errortype))
		if err != nil {
			logger.Error(err.Error())
		}
	}
}

func copyFileFromQueue(machineName string) {
	logpicqueue.BlockLogPicChannel(machineName)
	defer logpicqueue.UnblockLogPicChannel(machineName)
	var index = 1
	for {
		image := logpicqueue.GetChannelContent(machineName)
		if image == nil {
			break
		}
		fileoperation.CreateFolderifNotExist(config.Viper.GetString("logimage_path") + machineName)
		outputImagePath := config.Viper.GetString("logimage_path") + machineName + "/" + strconv.Itoa(index) + ".png"
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

func SendToRabbitMQ(hostname string, machineName string, locked string, path string, expireTime string) (err error) {
	var message Message
	message.Hostname = hostname
	message.MachineName = machineName
	message.Locked = 0
	if locked != "" {
		message.Locked = 1
		message.Coord = structure.CoordS2F(locked)
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
	rabbitmq.PublishWithExpiration("AI_queue1", jsonMessage, expireTime)
	imageFile.Close()
	return nil
}

func createNewErrorRecord(machineName string, errortype string, machineStatus structure.MachineStatus) structure.Errorlog {
	var errorlog structure.Errorlog
	errorlog.MachineName = machineName
	errorlog.Time = time.Now().Format("2006-01-02 15:04:05")
	errorlog.Type = errortype
	errorlog.TestItem = machineStatus.TestItem
	errorlog.Sku = machineStatus.Sku
	errorlog.Image = machineStatus.Image
	errorlog.Bios = machineStatus.Bios
	errorlog.Uuid = uuid.New().String()
	errorlog_query.SetErrorRecord(errorlog)
	return errorlog
}

// func setErrorType(machineName string, errorType int) {
// 	if errorType == 2 || errorType == 4 {
// 		errorType = 3
// 	}
// 	dut_query.UpdateDutStatus(machineName, errorType)
// }
