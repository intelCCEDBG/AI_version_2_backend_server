package structure

import (
	"fmt"
	"strconv"
	"strings"
)

type Kvm struct {
	Hostname          string `json:"hostname"`
	Stream_url        string `json:"url"`
	Stream_status     string `json:"status"`
	Stream_interface  string `json:"interface"`
	Start_record_time int64  `json:"start"`
}
type Target struct {
	Hostname string
	Status   string
	Ssim     int
	Wait     int
}
type Result_message struct {
	Hostname string    `json:"hostname"`
	Label    int64     `json:"label"`
	Coords   []float64 `json:"coords"`
	Created  int64     `json:"created"`
	Expire   int64     `json:"expire"`
}
type DUT struct {
	Machine_name   string  `json:"machine_name"`
	Ssim           float64 `json:"ssim"`
	Cycle_cnt      int     `json:"cycle"`
	Cycle_cnt_high int     `json:"cycle_high"`
	Status         int     `json:"status"`
	Threshhold     int     `json:"threshold"`
	Lock_coord     string  `json:"lock_coord"`
}
type Unit struct {
	Hostname     string
	Ip           string
	Machine_name string
	Project      string
}
type Unit_detail struct {
	Hostname     Kvm    `json:"kvms"`
	Ip           string `json:"dbgs"`
	Machine_name DUT    `json:"duts"`
	Project      string `json:"project"`
	Test_item    string `json:"test_item"`
	Sku          string `json:"sku"`
	Image        string `json:"image"`
	Bios         string `json:"bios"`
	Config       string `json:"config"`
}

// Status 0: BSOD 1: BLACK 2: RESTART 3: NORMAL 4: FREEZE
type AI_result struct {
	Hostname string    `json:"hostname"`
	Label    int       `json:"label"`
	Coords   []float64 `json:"coords"`
}

type Errorlog struct {
	Machine_name string `json:"machine_name"`
	Time         string `json:"time"`
	Type         string `json:"type"`
	Test_item    string `json:"test_item"`
	Sku          string `json:"sku"`
	Image        string `json:"image"`
	Bios         string `json:"bios"`
	Uuid         string `json:"uuid"`
	Config       string `json:"config"`
}

type Machine_status struct {
	Machine_name string `json:"machine"`
	Test_item    string `json:"test_item"`
	Sku          string `json:"sku"`
	Image        string `json:"image"`
	Bios         string `json:"bios"`
	Config       string `json:"config"`
}

type Ssim_and_threshold struct {
	Projet string `json:"project"`
	Ssim   int    `json:"ssim"`
	Thresh int    `json:"threshold"`
}

func Coord_f2s(coord []float64) string {
	var coord_str string
	for i := 0; i < len(coord); i++ {
		coord_str += fmt.Sprintf("%f", coord[i])
		if i != len(coord)-1 {
			coord_str += ","
		}
	}
	return coord_str
}
func Coord_s2f(coord string) []float64 {
	var coord_f []float64
	if coord == "" {
		return coord_f
	}
	values := strings.Split(coord, ",")
	for _, value := range values {
		// Convert the string to a float64
		if len(value) == 0 {
			break
		}
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println("Error converting string to float:", coord)
			return coord_f
		}
		coord_f = append(coord_f, floatValue)
	}
	return coord_f
}

type Project_setting_Tamplate struct {
	Project_name string `json:"project_name"`
	Short_name   string `json:"short_name"`
	// Host         []string         `json:"host"`
	Owner      string           `json:"owner"`
	Email_list []Email_tamplate `json:"Email_list"`
}

type Email_tamplate struct {
	Account string `json:"account"`
	Report  bool   `json:"report"`
	Alert   bool   `json:"alert"`
}
