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
	UNIT, err := method.Query("SELECT machine_name, ip, hostname, project FROM debug_unit where machine_name = " + "'" + machine_name + "'")
	if err != nil {
		logger.Error("Query Unit " + machine_name + " error: " + err.Error())
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

func Export_all_to_csv() {
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
