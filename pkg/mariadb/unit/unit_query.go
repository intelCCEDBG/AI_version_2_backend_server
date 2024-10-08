package unit_query

import (
	"fmt"
	"recorder/config"
	"recorder/internal/structure"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	debughost_query "recorder/pkg/mariadb/debughost"
	dut_query "recorder/pkg/mariadb/dut"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/mariadb/method"
	project_query "recorder/pkg/mariadb/project"
	"time"

	"github.com/google/uuid"
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

func CreateUnit(unit structure.Unit) (string, error) {
	uuid := uuid.New().String()
	query := `
		INSERT INTO debug_unit (uuid, machine_name, ip, hostname, project)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := method.Exec(query, uuid, unit.MachineName, unit.Ip, unit.Hostname, unit.Project)
	if err != nil {
		logger.Error("Insert unit error: " + err.Error())
		return "Insert unit error", err
	}
	return uuid, nil
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

func CreateKvmUnit(req structure.CreateKvmUnitRequest) (string, error) {
	// check if project code exists
	exists, err := project_query.ProjectCodeExists(req.ProjectCode)
	if err != nil {
		return "Check project code error: " + err.Error(), err
	}
	if exists {
		return "Project code already exists", nil
	}
	// create project
	err = project_query.CreateProject(req.ProjectName, req.ProjectCode)
	if err != nil {
		return "Create project error: " + err.Error(), err
	}
	// create kvm unit
	if req.KvmOwner != "" {
		// check if kvm hostname exists
		exists, err := kvm_query.KvmHostnameExists(req.KvmHostname)
		if err != nil {
			return "Check kvm hostname error: " + err.Error(), err
		}
		if exists {
			return "Kvm hostname already exists", nil
		}
		// create kvm
		err = kvm_query.CreateKvm(req.KvmIp, req.KvmHostname, req.KvmOwner)
		if err != nil {
			return "Create kvm error: " + err.Error(), err
		}
	}
	// create dut
	exists, err = dut_query.CheckDutExist(req.DutName)
	if err != nil {
		return "Check dut name error: " + err.Error(), err
	}
	if !exists {
		err = dut_query.CreateDut(req.DutName)
		if err != nil {
			return "Create dut error: " + err.Error(), err
		}
	}
	// create debug host
	if req.DebugHostOwner != "" {
		// check if debug host ip exists
		exists, err := debughost_query.CheckDebugHostIP(req.DebugHostIP)
		if err != nil {
			return "Check debug host ip error: " + err.Error(), err
		}
		if exists {
			return "Debug host ip already exists", nil
		}
		// create debug host
		err = debughost_query.CreateDebugHost(req.DebugHostIP, req.DebugHost, req.DebugHostOwner)
		if err != nil {
			return "Create debug host error: " + err.Error(), err
		}
	}
	// create debug unit
	uuid, err := CreateUnit(structure.Unit{
		MachineName: req.DutName,
		Ip:          req.DebugHostIP,
		Hostname:    req.KvmHostname,
		Project:     req.ProjectName,
	})
	if err != nil {
		return "Create debug unit error: " + err.Error(), err
	}
	return uuid, nil
}

func ExportAllToCsv() (string, error) {
	serverCsvPath := config.Viper.GetString("SERVER_CSV_PATH")
	queryCsvPath := config.Viper.GetString("CSV_PATH")
	// remove old csv files
	files := []string{"kvm.csv", "dbg.csv", "dut.csv", "map.csv"}
	for _, file := range files {
		err := fileoperation.DeleteFiles(serverCsvPath + file)
		if err != nil {
			logger.Error("Delete old csv files error: " + err.Error())
			return "", err
		}
	}
	query := fmt.Sprintf(`
		(
			SELECT 'hostname', 'ip', 'owner', 'version', 'NAS_ip'
			UNION ALL
			SELECT hostname, ip, owner, version, NAS_ip
			FROM kvm
		)
		INTO OUTFILE '%s'
		FIELDS TERMINATED BY ',' 
		LINES TERMINATED BY '\n';
	`, queryCsvPath+"kvm.csv")
	_, err := method.Exec(query)
	if err != nil {
		logger.Error("Export all kvms to csv error: " + err.Error())
		return "", err
	}
	query = fmt.Sprintf(`
		(
			SELECT 'owner', 'ip', 'hostname'
			UNION ALL
			SELECT owner, ip, hostname
			FROM debug_host
		)
		INTO OUTFILE '%s'
		FIELDS TERMINATED BY ','
		LINES TERMINATED BY '\n';
	`, queryCsvPath+"dbg.csv")
	_, err = method.Exec(query)
	if err != nil {
		logger.Error("Export all debug hosts to csv error: " + err.Error())
		return "", err
	}
	query = fmt.Sprintf(`
		(
			SELECT 'machine_name', 'ssim', 'threshold'
			UNION ALL
			SELECT machine_name, ssim, threshold
			FROM machine
		)
		INTO OUTFILE '%s'
		FIELDS TERMINATED BY ','
		LINES TERMINATED BY '\n';
	`, queryCsvPath+"dut.csv")
	_, err = method.Exec(query)
	if err != nil {
		logger.Error("Export all duts to csv error: " + err.Error())
		return "", err
	}
	query = fmt.Sprintf(`
		(
			SELECT 'hostname(KVM)', 'ip(debug host)', 'machine_name', 'project'
			UNION ALL
			SELECT hostname, ip, machine_name, project
			FROM debug_unit
		)
		INTO OUTFILE '%s'
		FIELDS TERMINATED BY ','
		LINES TERMINATED BY '\n';
	`, queryCsvPath+"map.csv")
	_, err = method.Exec(query)
	if err != nil {
		logger.Error("Export all units to csv error: " + err.Error())
		return "", err
	}
	// zip all csv files
	currentDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Format("2006-01-02")
	result := fmt.Sprintf("%s%s_mapping.zip", serverCsvPath, currentDate)
	err = fileoperation.ZipFiles(result, serverCsvPath+"kvm.csv", serverCsvPath+"dbg.csv", serverCsvPath+"dut.csv", serverCsvPath+"map.csv")
	if err != nil {
		logger.Error("Zip all csv files error: " + err.Error())
		return "", err
	}
	return "https://10.227.106.11:8000/csv/" + currentDate + "_mapping.zip", nil
}
