package main

import (
	"fmt"
	"os/exec"
	"recorder/internal/ssim"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
	"time"

	"github.com/miekg/dns"
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
	// config.LoadConfig()
	// err := mariadb.ConnectDB()
	// if err != nil {
	// 	fmt.Println("Connect to mariadb error:", err)
	// 	return
	// }
	// err = unit_query.ExportAllToCsv()
	// if err != nil {
	// 	fmt.Println("Export all units to csv error:", err)
	// 	return
	// }

	// Set the IP address you want to resolve
	ip := "10.5.238.12"
	ptrIP, err := dns.ReverseAddr(ip)
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(ptrIP, dns.TypePTR)
	r, t, err := c.Exchange(m, "10.248.2.1:53")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Query time: %v\n", t)
	if len(r.Answer) == 0 {
		fmt.Println("No PTR records found")
		return
	}
	for _, ans := range r.Answer {
		if ptr, ok := ans.(*dns.PTR); ok {
			fmt.Println("Hostname:", ptr.Ptr)
		}
	}

	// datasetPath := "../one_on_one_plan/cursor/dataset/"
	// // streams := []string{"http://10.5.238.52:8081", "http://10.5.238.80:8081"}
	// duts := []string{"18F_24"}
	// timeLength := 5 * time.Minute

	// for i, stream := range streams {
	// 	machine := "21F_R7B2"
	// 	if i == 0 {
	// 		machine = "21F_R5A2"
	// 	}
	// 	capture(int(timeLength.Seconds()), stream, datasetPath+machine+"/")
	// }

	// for _, dut := range duts {
	// 	// make sure the dataset folder exists
	// 	if _, err := os.Stat(datasetPath + dut); os.IsNotExist(err) {
	// 		os.Mkdir(datasetPath+dut, 0777)
	// 	}
	// 	// get kvm ip
	// 	unit := unit_query.GetByMachine(dut)
	// 	ip := unit.Ip
	// 	times := int(timeLength.Seconds())
	// 	streamUrl := fmt.Sprintf("http://%s:8081", ip)
	// 	// start capturing
	// 	fmt.Println("Start capturing for dut", dut, "from", streamUrl)
	// 	capture(times, streamUrl, datasetPath+dut+"/")
	// }

	// err := kvm.PressWindowsKey("10.5.254.155")
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// }
}
