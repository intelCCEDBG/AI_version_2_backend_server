package dut_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	"strconv"
	"time"
)

func Update_dut_status(machine_name string, status int) {
	_, err := method.Exec("UPDATE machine SET status = ? WHERE machine_name = ?", status, machine_name)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}
func Update_AI_result(machine_name string, status int64, coords []float64) {
	coords_str := ""
	for i := 0; i < len(coords); i++ {
		coords_str += strconv.FormatFloat(coords[i], 'f', 6, 64)
		if i != len(coords)-1 {
			coords_str += ","
		}
	}
	cur_time := time.Now()
	_, err := method.Exec("INSERT INTO ai_result (machine_name, status, coords, time) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE status = VALUES(status), coords = VALUES(coords),time = VALUES(time);", machine_name, status, coords_str, cur_time)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
}
func Get_AI_result(machine_name string) (result structure.AI_result) {
	Result, err := method.Query("SELECT status, coords FROM ai_result WHERE machine_name = ?", machine_name)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var status int64
	var coords_str string
	dataExist := false
	for Result.Next() {
		err := Result.Scan(&status, &coords_str)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		dataExist = true
	}
	if !dataExist {
		return structure.AI_result{
			Hostname: "null",
		}
	}
	var ai_result structure.AI_result
	ai_result.Hostname = machine_name
	ai_result.Label = status
	ai_result.Coords = structure.Coord_s2f(coords_str)
	return ai_result
}
func Update_dut_cnt(machine_name string, cnt int) {
	_, err := method.Exec("UPDATE machine SET cycle_cnt = ? WHERE machine_name = ?", cnt, machine_name)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}
func Update_lock_coord(machine_name string, coord string) {
	_, err := method.Exec("UPDATE machine SET lock_coord = ? WHERE machine_name = ?", coord, machine_name)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}
func Get_dut_status(machine_name string) (dut_template structure.DUT) {
	KVM, err := method.Query("SELECT machine_name,ssim,cycle_cnt,status,threshold,lock_coord FROM machine where machine_name = " + "'" + machine_name + "'")
	if err != nil {
		logger.Error("Query DUT " + machine_name + " error: " + err.Error())
	}
	dut_template.Machine_name = "null"
	for KVM.Next() {
		err := KVM.Scan(&dut_template.Machine_name, &dut_template.Ssim, &dut_template.Cycle_cnt, &dut_template.Status, &dut_template.Threshhold, &dut_template.Lock_coord)
		if err != nil {
			logger.Error(err.Error())
			return dut_template
		}
	}
	return dut_template
}
