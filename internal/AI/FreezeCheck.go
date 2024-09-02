package ai

import (
	"fmt"
	"recorder/config"
	"recorder/internal/cropping"
	"recorder/internal/ffmpeg"
	"recorder/internal/kvm"
	"recorder/internal/ssim"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/redis"
	"time"
)

func FreezeCheck(machineName string, coords []float64, KVM structure.Kvm) (bool, error) {
	logger.Info("Performing windows key freeze check ...")
	freezed := true
	ip := kvm_query.GetIP(KVM.Hostname)
	// press windows key
	redis.Redis_set("kvm:"+KVM.Hostname+":holding", KVM.Hostname)
	defer redis.Redis_del("kvm:" + KVM.Hostname + ":holding")
	// take a picture
	unpressedImagePath := config.Viper.GetString("LOGIMAGE_PATH") + machineName + "/unpressed.png"
	ffmpeg.Take_photo(unpressedImagePath, KVM)
	// press windows key
	err := kvm.PressWindowsKey(ip)
	if err != nil {
		logger.Error("Error pressing windows key: " + err.Error())
		return freezed, err
	}
	time.Sleep(1 * time.Second)
	// take a picture
	pressedImagePath := config.Viper.GetString("LOGIMAGE_PATH") + machineName + "/pressed.png"
	ffmpeg.Take_photo(pressedImagePath, KVM)
	// calculate ssim
	_, err = cropping.Crop_image(unpressedImagePath, coords, unpressedImagePath)
	if err != nil {
		logger.Error("Error cropping image: " + err.Error())
		return freezed, err
	}
	_, err = cropping.Crop_image(pressedImagePath, coords, pressedImagePath)
	if err != nil {
		logger.Error("Error cropping image: " + err.Error())
		return freezed, err
	}
	ssimResult, err := ssim.Ssim_cal(pressedImagePath, unpressedImagePath)
	if err != nil {
		logger.Error("Error calculating SSIM: " + err.Error())
		return freezed, err
	}
	logger.Info("SSIM: " + fmt.Sprintf("%f", ssimResult))
	if ssimResult < 0.8 {
		freezed = false
		// press windows key back
		err = kvm.PressWindowsKey(ip)
		if err != nil {
			logger.Error("Error pressing windows key: " + err.Error())
			return freezed, err
		}
	}
	logger.Info("FreezeCheck result: " + fmt.Sprintf("%t", freezed))
	return freezed, nil
}