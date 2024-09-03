package main

import (
	"fmt"
	"os"

	// "os/exec"
	"recorder/internal/kvm"
	"recorder/internal/ssim"

	// "recorder/pkg/mariadb"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
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

func main() {
	ip := os.Args[1]
	// streamUrl := fmt.Sprintf("http://%s:8081", ip)
	// cmd := exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", streamUrl, "-vframes", "1", "./pressed1.png")
	// err := cmd.Run()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	err := kvm.PressWindowsKey(ip)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// cmd = exec.Command("ffmpeg", "-loglevel", "quiet", "-y", "-i", streamUrl, "-vframes", "1", "./pressed2.png")
	// err = cmd.Run()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// err = kvm.PressWindowsKey(ip)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
	// ssimtest("./pressed1.png", "./pressed2.png")
}
