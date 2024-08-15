package main

import (
	"fmt"
	"recorder/internal/ssim"
	"recorder/pkg/mariadb"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
)

func ssimtest() {
	image1 := "./image1.png"
	image2 := "./image2.png"
	ssimValue, err := ssim.Ssim_cal(image1, image2)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("SSIM Value:", ssimValue)
	}
}

func failcounttest() {
	times := errorlog_query.Get_error_count_after_time("18F_21", "2024-08-13 16:28:45")
	fmt.Println("Fail count:", times)
}
func main() {
	err := mariadb.ConnectDB()
	if err != nil {

		return
	}
	failcounttest()
}
