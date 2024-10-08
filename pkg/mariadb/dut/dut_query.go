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

func UpdateAIResult(machineName string, status int64, coords []float64) {
	coordsStr := ""
	for i := 0; i < len(coords); i++ {
		coordsStr += strconv.FormatFloat(coords[i], 'f', 6, 64)
		if i != len(coords)-1 {
			coordsStr += ","
		}
	}
	curTime := time.Now()
	_, err := method.Exec("REPLACE INTO ai_result (machine_name, status, coords, time) VALUES (?, ?, ?, ?)", machineName, status, coordsStr, curTime)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
}

func GetAiResult(machineName string) (result structure.AiResult) {
	Result, err := method.Query("SELECT status, coords FROM ai_result WHERE machine_name = ?", machineName)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var status int
	var coordsStr string
	dataExist := false
	for Result.Next() {
		err := Result.Scan(&status, &coordsStr)
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
	var aiResult structure.AiResult
	aiResult.Hostname = machineName
	aiResult.Label = status
	aiResult.Coords = structure.CoordS2F(coordsStr)
	return aiResult
}

func GetProjectName(machineName string) string {
	result, err := method.Query("SELECT project_name FROM debug_unit JOIN project ON debug_unit.project=project.project_name where debug_unit.machine_name = ?", machineName)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var Project string
	for result.Next() {
		err := result.Scan(&Project)
		if err != nil {
			logger.Error(err.Error())
		}
	}
	return Project
}

func GetProjectCode(machineName string) string {
	Result := method.QueryRow("SELECT short_name FROM debug_unit JOIN project ON debug_unit.project=project.project_name where debug_unit.machine_name = ?", machineName)
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

func UpdateLockCoord(machineName string, coord string) {
	_, err := method.Exec("UPDATE machine SET lock_coord = ? WHERE machine_name = ?", coord, machineName)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}

func GetDutStatus(machineName string) (dut_template structure.DUT) {
	KVM, err := method.Query("SELECT machine_name,ssim,cycle_cnt,cycle_cnt_high, status,threshold,lock_coord FROM machine where machine_name =" + "'" + machineName + "'")
	if err != nil {
		logger.Error("Query DUT " + machineName + " error: " + err.Error())
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

func GetAllDutStatus() (dutTemplate []structure.DUT) {
	KVM, err := method.Query("SELECT machine_name,ssim,cycle_cnt,cycle_cnt_high,status,threshold,lock_coord FROM machine")
	if err != nil {
		logger.Error("Query All DUT error: " + err.Error())
	}
	for KVM.Next() {
		var dut structure.DUT
		err := KVM.Scan(&dut.MachineName, &dut.Ssim, &dut.CycleCnt, &dut.CycleCntHigh, &dut.Status, &dut.Threshhold, &dut.LockCoord)
		if err != nil {
			logger.Error(err.Error())
			return dutTemplate
		}
		dutTemplate = append(dutTemplate, dut)
	}
	return dutTemplate
}

func Clean_Cycle_Count(machineName string) {
	logger.Info("Clean Cycle...")
	_, err := method.Exec("UPDATE machine SET cycle_cnt = ? WHERE machine_name = ?", 0, machineName)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}

func SetDutStatusFromKvm(status int, kvm structure.Kvm) {
	_, err := method.Exec("UPDATE machine SET status=? WHERE machine_name = (SELECT machine_name FROM debug_unit WHERE hostname=?)", status, kvm.Hostname)
	if err != nil {
		logger.Error("update dut status error" + err.Error())
		return
	}
}

func GetMachineStatus(machineName string) (machineStatus structure.MachineStatus) {
	KVM, err := method.Query("SELECT machine_name,test_item,sku,image,bios FROM machine_status where machine_name =" + "'" + machineName + "'")
	if err != nil {
		logger.Error("Query DUT " + machineName + " error: " + err.Error())
	}
	machineStatus.MachineName = machineName
	machineStatus.Bios = "null"
	machineStatus.Image = "null"
	machineStatus.Sku = "null"
	machineStatus.TestItem = "null"

	for KVM.Next() {
		err := KVM.Scan(&machineStatus.MachineName, &machineStatus.TestItem, &machineStatus.Sku, &machineStatus.Image, &machineStatus.Bios)
		if err != nil {
			logger.Error(err.Error())
			return machineStatus
		}
	}
	return machineStatus
}

func SetMachineStatus(machineStatus structure.MachineStatus) {
	_, err := method.Exec("REPLACE INTO machine_status (machine_name,test_item,sku, image, bios, config) VALUES (?, ?, ?, ?, ?, ?)", machineStatus.MachineName, machineStatus.TestItem, machineStatus.Sku, machineStatus.Image, machineStatus.Bios, machineStatus.Config)
	if err != nil {
		logger.Error("update dut status error" + err.Error())
		return
	}
}

func CheckDutExist(machineName string) (bool, error) {
	query := "SELECT machine_name FROM machine WHERE machine_name = ?"
	row := method.QueryRow(query, machineName)
	var name string
	err := row.Scan(&name)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		logger.Error("Check dut exist error: " + err.Error())
		return false, err
	}
	return true, nil
}

func CreateDut(machineName string) (err error) {
	query := `
		INSERT INTO machine (machine_name, ssim, status, cycle_cnt, cycle_cnt_high, error_timestamp, path, threshold, lock_coord)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = method.Exec(query, machineName, 80, 1, 0, 0, 0, "null", 15, "")
	if err != nil {
		logger.Error("Insert dut error: " + err.Error())
		return err
	}
	return nil
}
