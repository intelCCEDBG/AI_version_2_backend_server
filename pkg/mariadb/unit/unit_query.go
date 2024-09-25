package unit_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func GetUnitByHostname(hostname string) structure.Unit {
	var unit_template structure.Unit
	UNIT, err := method.Query("SELECT machine_name,ip,hostname,project FROM debug_unit where hostname = " + "'" + hostname + "'")
	if err != nil {
		logger.Error("Query Unit " + hostname + " error: " + err.Error())
	}
	for UNIT.Next() {
		err := UNIT.Scan(&unit_template.MachineName, &unit_template.Ip, &unit_template.Hostname, &unit_template.Project)
		if err != nil {
			logger.Error(err.Error())
			return unit_template
		}
	}
	return unit_template
}

func GetByMachine(machine_name string) structure.Unit {
	var unit_template structure.Unit
	unit, err := method.Query("SELECT machine_name, ip, hostname, project FROM debug_unit where machine_name = " + "'" + machine_name + "'")
	if err != nil {
		logger.Error("Query Unit " + machine_name + " error: " + err.Error())
	}
	for unit.Next() {
		err := unit.Scan(&unit_template.MachineName, &unit_template.Ip, &unit_template.Hostname, &unit_template.Project)
		if err != nil {
			logger.Error(err.Error())
			return unit_template
		}
	}
	return unit_template
}

func CheckKvmUnit(ip string) (string, structure.DebugUnitDetail, error) {
	status := "Unknown" // default status
	result := structure.DebugUnitDetail{}
	// get kvm hostname
	info := method.QueryRow("SELECT hostname, ip, stream_status FROM kvm where ip = ?", ip)
	if info == nil {
		return status, result, nil
	}
	err := info.Scan(&result.KvmHostName, &result.KvmIP, &result.KvmStatus)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return status, result, nil
		}
		logger.Error("Reading kvm hostname error: " + err.Error())
		return status, result, err
	}
	status = "Available"

	// get debug unit info
	info = method.QueryRow("SELECT machine_name, ip, hostname, project FROM debug_unit where hostname = ?", result.KvmHostName)
	if info == nil {
		return status, result, nil
	}
	err = info.Scan(&result.MachineName, &result.DebugHostIP, &result.KvmHostName, &result.ProjectName)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return status, result, nil
		}
		logger.Error("Reading debug unit info error: " + err.Error())
		return status, result, err
	}
	status = "In Use"

	// get debug host info
	info = method.QueryRow("SELECT hostname, ip FROM debug_host where ip = ?", result.DebugHostIP)
	if info == nil {
		return status, result, nil
	}
	err = info.Scan(&result.DebugHostName, &result.DebugHostIP)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return status, result, nil
		}
		logger.Error("Reading debug host info error: " + err.Error())
		return status, result, err
	}
	return status, result, nil
}

func ExportAllToCsv() {
	_, err := method.Query("SELECT * FROM debug_unit INTO OUTFILE '/home/jimmy/multi_streaming_recorder/cmd/backend/upload/debug_unit.csv' FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n';")
	if err != nil {
		logger.Error("Export all units to csv error: " + err.Error())
	}
	_, err = method.Query("SELECT * FROM kvm INTO OUTFILE 'upload/kvm.csv' FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n';")
	if err != nil {
		logger.Error("Export all units to csv error: " + err.Error())
	}
	_, err = method.Query("SELECT * FROM debug_host INTO OUTFILE 'upload/dbg.csv' FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n';")
	if err != nil {
		logger.Error("Export all units to csv error: " + err.Error())
	}
	_, err = method.Query("SELECT * FROM machine INTO OUTFILE 'upload/dut.csv' FIELDS TERMINATED BY ',' ENCLOSED BY '\"' LINES TERMINATED BY '\n';")
	if err != nil {
		logger.Error("Export all units to csv error: " + err.Error())
	}
}
