package dut_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	"strconv"
	"time"
)

func UpdateDutStatus(machineName string, status int) {
	_, err := method.Exec("UPDATE machine SET status = ? WHERE machine_name = ?", status, machineName)
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
	_, err := method.Exec("REPLACE INTO ai_result (machine_name, status, coords, time) VALUES (?, ?, ?, ?)", machine_name, status, coords_str, cur_time)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
}
func GetAIResult(machine_name string) (result structure.AiResult) {
	Result, err := method.Query("SELECT status, coords FROM ai_result WHERE machine_name = ?", machine_name)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var status int
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
		return structure.AiResult{
			Hostname: "null",
		}
	}
	var ai_result structure.AiResult
	ai_result.Hostname = machine_name
	ai_result.Label = status
	ai_result.Coords = structure.CoordS2F(coords_str)
	return ai_result
}
func GetProjectName(machine_name string) string {
	Result, err := method.Query("SELECT project_name FROM debug_unit JOIN project ON debug_unit.project=project.project_name where debug_unit.machine_name = ?", machine_name)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var Project string
	for Result.Next() {
		err := Result.Scan(&Project)
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return Project
}
func GetProjectCode(machine_name string) string {
	Result := method.QueryRow("SELECT short_name FROM debug_unit JOIN project ON debug_unit.project=project.project_name where debug_unit.machine_name = ?", machine_name)
	var Project string
	err := Result.Scan(&Project)
	if err != nil {
		logger.Error("Search debug_unit error" + err.Error())
	}
	return Project
}

func UpdateDutCycleCnt(machineName string, cnt int) {
	_, err := method.Exec("UPDATE machine SET cycle_cnt = ? WHERE machine_name = ?", cnt, machineName)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}

func UpdateDutCycleCntHigh(machineName string, cnt int) {
	_, err := method.Exec("UPDATE machine SET cycle_cnt_high = ? WHERE machine_name = ?", cnt, machineName)
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
func GetDutStatus(machine_name string) (dut_template structure.DUT) {
	KVM, err := method.Query("SELECT machine_name,ssim,cycle_cnt,cycle_cnt_high, status,threshold,lock_coord FROM machine where machine_name = " + "'" + machine_name + "'")
	if err != nil {
		logger.Error("Query DUT " + machine_name + " error: " + err.Error())
	}
	dut_template.MachineName = "null"
	for KVM.Next() {
		err := KVM.Scan(&dut_template.MachineName, &dut_template.Ssim, &dut_template.CycleCnt, &dut_template.CycleCntHigh, &dut_template.Status, &dut_template.Threshhold, &dut_template.LockCoord)
		if err != nil {
			logger.Error(err.Error())
			return dut_template
		}
	}
	return dut_template
}
func Get_all_dut_status() (dut_template []structure.DUT) {

	KVM, err := method.Query("SELECT machine_name,ssim,cycle_cnt,cycle_cnt_high,status,threshold,lock_coord FROM machine")
	if err != nil {
		logger.Error("Query All DUT error: " + err.Error())
	}
	for KVM.Next() {
		var dut structure.DUT
		err := KVM.Scan(&dut.MachineName, &dut.Ssim, &dut.CycleCnt, &dut.CycleCntHigh, &dut.Status, &dut.Threshhold, &dut.LockCoord)
		if err != nil {
			logger.Error(err.Error())
			return dut_template
		}
		dut_template = append(dut_template, dut)
	}
	return dut_template
}

func Clean_Cycle_Count(machine_name string) {
	logger.Info("Clean Cycle...")
	_, err := method.Exec("UPDATE machine SET cycle_cnt = ? WHERE machine_name = ?", 0, machine_name)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}

func Set_dut_status_from_kvm(status int, kvm structure.Kvm) {
	_, err := method.Exec("UPDATE machine SET status=? WHERE machine_name = (SELECT machine_name FROM debug_unit WHERE hostname=?)", status, kvm.Hostname)
	if err != nil {
		logger.Error("update dut status error" + err.Error())
		return
	}
}
func GetMachineStatus(machine_name string) (machine_status structure.MachineStatus) {
	KVM, err := method.Query("SELECT machine_name,test_item,sku,image,bios FROM machine_status where machine_name = " + "'" + machine_name + "'")
	if err != nil {
		logger.Error("Query DUT " + machine_name + " error: " + err.Error())
	}
	machine_status.MachineName = machine_name
	machine_status.Bios = "null"
	machine_status.Image = "null"
	machine_status.Sku = "null"
	machine_status.TestItem = "null"

	for KVM.Next() {
		err := KVM.Scan(&machine_status.MachineName, &machine_status.TestItem, &machine_status.Sku, &machine_status.Image, &machine_status.Bios)
		if err != nil {
			logger.Error(err.Error())
			return machine_status
		}
	}
	return machine_status
}
func Set_machine_status(machine_status structure.MachineStatus) {
	_, err := method.Exec("REPLACE INTO machine_status (machine_name,test_item,sku, image, bios, config) VALUES (?, ?, ?, ?, ?, ?)", machine_status.MachineName, machine_status.TestItem, machine_status.Sku, machine_status.Image, machine_status.Bios, machine_status.Config)
	if err != nil {
		logger.Error("update dut status error" + err.Error())
		return
	}
}
