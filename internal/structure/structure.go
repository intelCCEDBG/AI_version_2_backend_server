package structure

import (
	"fmt"
	"strconv"
	"strings"
)

type Kvm struct {
	Hostname          string
	Stream_url        string
	Stream_status     string
	Stream_interface  string
	Start_record_time int64
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
	Machine_name string
	Ssim         float64
	Cycle_cnt    int
	Status       string
	Threshhold   int
	Lock_coord   string
}

// Status 0: BSOD 1: BLACK 2: RESTART 3: NORMAL 4: FREEZE
type AI_result struct {
	Hostname string    `json:"hostname"`
	Label    int64     `json:"label"`
	Coords   []float64 `json:"coords"`
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
	values := strings.Split(coord, ",")
	for _, value := range values {
		// Convert the string to a float64
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println("Error converting string to float:", err)
			return coord_f
		}
		coord_f = append(coord_f, floatValue)
	}
	return coord_f
}
