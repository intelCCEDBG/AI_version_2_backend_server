package main

import (
	"fmt"
	"os"
	"os/exec"
	"recorder/config"
	"recorder/internal/ssim"
	"recorder/pkg/mariadb"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
	unit_query "recorder/pkg/mariadb/unit"
	"time"
)

func ssimtest(image1, image2 string) {
	ssimValue, err := ssim.SsimCal(image1, image2)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("SSIM Value:", ssimValue)
	}
}

func failcounttest() {
	times := errorlog_query.GetErrorCountAfterTime("18F_21", "2024-08-13 16:28:45")
	fmt.Println("Fail count:", times)
}

func capture(times int, streamUrl string, path string) {
	fmt.Println("Start capturing from", streamUrl)
	for i := 0; i < times; i++ {
		storedPath := path + fmt.Sprintf("%d", i) + ".png"
		cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", streamUrl, "-vframes", "1", storedPath)
		fmt.Println("Capturing to", storedPath)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error:", err)
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("Finish capturing from", streamUrl)
}

func main() {
	config.LoadConfig()
	err := mariadb.ConnectDB()
	if err != nil {
		fmt.Println("Connect to mariadb error:", err)
		return
	}
	datasetPath := "../one_on_one_plan/cursor/dataset/"
	duts := []string{"18F_26", "18F_30", "18F_31"}
	timeLength := 5 * time.Minute

	for _, dut := range duts {
		// make sure the dataset folder exists
		if _, err := os.Stat(datasetPath + dut); os.IsNotExist(err) {
			os.Mkdir(datasetPath+dut, 0777)
		}
		// get kvm ip
		unit := unit_query.GetByMachine(dut)
		ip := unit.Ip
		times := int(timeLength.Seconds())
		streamUrl := fmt.Sprintf("http://%s:8081", ip)
		// start capturing
		fmt.Println("Start capturing for dut", dut, "from", streamUrl)
		capture(times, streamUrl, datasetPath+dut+"/")
	}
}
