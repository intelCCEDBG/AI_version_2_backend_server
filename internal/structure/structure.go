package structure

import (
	"fmt"
	"strconv"
	"strings"
)

type Kvm struct {
	Hostname        string `json:"hostname"`
	StreamUrl       string `json:"url"`
	StreamStatus    string `json:"status"`
	StreamInterface string `json:"interface"`
	StartRecordTime int64  `json:"start"`
}
type Target struct {
	Hostname string
	Status   string
	Ssim     int
	Wait     int
}
type ResultMessage struct {
	Hostname string    `json:"hostname"`
	Label    int64     `json:"label"`
	Coords   []float64 `json:"coords"`
	Created  int64     `json:"created"`
	Expire   int64     `json:"expire"`
}
type DUT struct {
	MachineName  string  `json:"machine_name"`
	Ssim         float64 `json:"ssim"`
	CycleCnt     int     `json:"cycle"`
	CycleCntHigh int     `json:"cycle_high"`
	Status       int     `json:"status"`
	Threshhold   int     `json:"threshold"`
	LockCoord    string  `json:"lock_coord"`
}
type Unit struct {
	Hostname    string
	Ip          string
	MachineName string
	Project     string
}
type UnitDetail struct {
	Hostname    Kvm    `json:"kvms"`
	Ip          string `json:"dbgs"`
	MachineName DUT    `json:"duts"`
	Project     string `json:"project"`
	TestItem    string `json:"test_item"`
	Sku         string `json:"sku"`
	Image       string `json:"image"`
	Bios        string `json:"bios"`
	Config      string `json:"config"`
}

// Status 0: BSOD 1: BLACK 2: RESTART 3: NORMAL 4: FREEZE
const (
	BSOD_LABEL    int = 0
	BLACK_LABEL   int = 1
	RESTART_LABEL int = 2
	NORMAL_LABEL  int = 3
	BSOD          int = 0
	BLACK         int = 1
	RESTART       int = 2
	FREEZE        int = 3
	NORMAL        int = 4
)

type AiResult struct {
	Hostname string    `json:"hostname"`
	Label    int       `json:"label"`
	Coords   []float64 `json:"coords"`
}

type Errorlog struct {
	MachineName string `json:"machine_name"`
	Time        string `json:"time"`
	Type        string `json:"type"`
	TestItem    string `json:"test_item"`
	Sku         string `json:"sku"`
	Image       string `json:"image"`
	Bios        string `json:"bios"`
	Uuid        string `json:"uuid"`
	Config      string `json:"config"`
}

type MachineStatus struct {
	MachineName string `json:"machine"`
	TestItem    string `json:"test_item"`
	Sku         string `json:"sku"`
	Image       string `json:"image"`
	Bios        string `json:"bios"`
	Config      string `json:"config"`
}

type SsimAndThreshold struct {
	Project string `json:"project"`
	Ssim    int    `json:"ssim"`
	Thresh  int    `json:"threshold"`
}

func CoordF2S(coord []float64) string {
	var coordStr string
	for i := 0; i < len(coord); i++ {
		coordStr += fmt.Sprintf("%f", coord[i])
		if i != len(coord)-1 {
			coordStr += ","
		}
	}
	return coordStr
}
func CoordS2F(coord string) []float64 {
	var coordF []float64
	if coord == "" {
		return coordF
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
			return coordF
		}
		coordF = append(coordF, floatValue)
	}
	return coordF
}

type ProjectSettingTemplate struct {
	ProjectName string          `json:"project_name"`
	ShortName   string          `json:"short_name"`
	Owner       string          `json:"owner"`
	EmailList   []EmailTemplate `json:"Email_list"`
}

type EmailTemplate struct {
	Account string `json:"account"`
	Report  bool   `json:"report"`
	Alert   bool   `json:"alert"`
}
