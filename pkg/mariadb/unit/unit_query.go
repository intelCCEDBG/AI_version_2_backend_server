package unit_query

import (
	"fmt"
	"recorder/config"
	"recorder/internal/structure"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb"
	debughost_query "recorder/pkg/mariadb/debughost"
	dut_query "recorder/pkg/mariadb/dut"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/mariadb/method"
	project_query "recorder/pkg/mariadb/project"
	"time"

	"github.com/google/uuid"
)

func GetUnitsInfoByProject(projectName string) ([]structure.DebugUnitInfo, error) {
	unitsInfo := []structure.DebugUnitInfo{}
	rows, err := method.Query("SELECT machine_name, ip, hostname FROM debug_unit WHERE project = ?", projectName)
	if err != nil {
		logger.Error("Query units info by project error: " + err.Error())
		return unitsInfo, err
	}
	for rows.Next() {
		var unitInfo structure.DebugUnitInfo
		err := rows.Scan(&unitInfo.MachineName, &unitInfo.DebugHost, &unitInfo.Hostname)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				return unitsInfo, nil
			}
			logger.Error("Scan units info by project error: " + err.Error())
			return unitsInfo, err
		}
		row := method.QueryRow("SELECT status, lock_coord, high_frame_rate FROM machine WHERE machine_name = ?", unitInfo.MachineName)
		err = row.Scan(&unitInfo.Status, &unitInfo.LockCoord, &unitInfo.HighFrameRate)
		if err != nil {
			logger.Error("Scan machine status error: " + err.Error())
			return unitsInfo, err
		}
		row = method.QueryRow("SELECT stream_status, stream_url FROM kvm WHERE hostname = ?", unitInfo.Hostname)
		err = row.Scan(&unitInfo.RecordStatus, &unitInfo.KvmLink)
		if err != nil {
			logger.Error("Scan kvm status error: " + err.Error())
			return unitsInfo, err
		}
		row = method.QueryRow("SELECT time FROM errorlog WHERE machine_name = ? ORDER BY time DESC LIMIT 1", unitInfo.MachineName)
		err = row.Scan(&unitInfo.LastFail)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				unitInfo.LastFail = "N/A"
			} else {
				logger.Error("Scan last fail error: " + err.Error())
				return unitsInfo, err
			}
		}
		unitsInfo = append(unitsInfo, unitInfo)
	}
	return unitsInfo, nil
}

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

func GetByMachine(machineName string) structure.Unit {
	var unitTemplate structure.Unit
	unit, err := method.Query("SELECT machine_name, ip, hostname, project FROM debug_unit where machine_name = " + "'" + machineName + "'")
	if err != nil {
		logger.Error("Query Unit " + machineName + " error: " + err.Error())
	}
	for unit.Next() {
		err := unit.Scan(&unitTemplate.MachineName, &unitTemplate.Ip, &unitTemplate.Hostname, &unitTemplate.Project)
		if err != nil {
			logger.Error(err.Error())
			return unitTemplate
		}
	}
	return unitTemplate
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
	tr, err := mariadb.DB.Begin()
	if err != nil {
		return "Create transaction error: " + err.Error(), err
	}
	// check if project code exists
	exists, err := project_query.ProjectCodeExists(req.ProjectCode)
	if err != nil {
		return "Check project code error: " + err.Error(), err
	}
	if exists {
		return "Project code already exists", nil
	}
	// create project
	query := `
		INSERT INTO project (project_name, short_name, owner, email_list, status, freeze_detection)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = tr.Exec(query, req.ProjectName, req.ProjectCode, "", "", 0, "open")
	if err != nil {
		tr.Rollback()
		return "Create project error: " + err.Error(), err
	}
	// create kvm unit
	if req.KvmOwner != "" {
		// check if kvm hostname exists
		exists, err := kvm_query.KvmHostnameExists(req.KvmHostname)
		if err != nil {
			tr.Rollback()
			return "Check kvm hostname error: " + err.Error(), err
		}
		if exists {
			tr.Rollback()
			return "Kvm hostname already exists", nil
		}
		// create kvm
		query = `
			INSERT INTO kvm (hostname, ip, owner, status, version, NAS_ip, stream_url, stream_status, stream_interface, start_record_time)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err = tr.Exec(query, req.KvmHostname, req.KvmIp, req.KvmOwner, "null", 1, "null", "http://"+req.KvmIp+":8081", "idle", "null", 0)
		if err != nil {
			tr.Rollback()
			return "Create kvm error: " + err.Error(), err
		}
	}
	// check dut
	exists, err = dut_query.CheckDutExist(req.DutName)
	if err != nil {
		tr.Rollback()
		return "Check dut name error: " + err.Error(), err
	}
	if !exists {
		tr.Rollback()
		return "Dut name not exists", nil
	}
	// check debug host
	exists, err = debughost_query.CheckDebugHostIPExist(req.DebugHostIP)
	if err != nil {
		tr.Rollback()
		return "Check debug host ip error: " + err.Error(), err
	}
	if !exists {
		tr.Rollback()
		return "Debug host ip not exists", nil
	}
	// create debug unit
	uuid := uuid.New().String()
	query = `DELETE FROM debug_unit WHERE hostname = ? OR ip = ? OR machine_name = ?`
	_, err = tr.Exec(query, req.KvmHostname, req.DebugHostIP, req.DutName)
	if err != nil {
		tr.Rollback()
		return "Delete debug unit error: " + err.Error(), err
	}
	query = `
		INSERT INTO debug_unit (uuid, machine_name, ip, hostname, project)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err = tr.Exec(query, uuid, req.DutName, req.DebugHostIP, req.KvmHostname, req.ProjectName)
	if err != nil {
		tr.Rollback()
		return "Create debug unit error: " + err.Error(), err
	}
	tr.Commit()
	return "Successfully Created, UUID: " + uuid, nil
}

func LeaveProject(dutName string) error {
	_, err := method.Exec("UPDATE debug_unit SET project = 'null' WHERE machine_name = ?", dutName)
	if err != nil {
		logger.Error("Leave project error: " + err.Error())
		return err
	}
	return nil
}

func JoinProject(projectName string, duts []string) error {
	tr, err := mariadb.DB.Begin()
	if err != nil {
		return err
	}
	for _, dut := range duts {
		_, err := tr.Exec("UPDATE debug_unit SET project = ? WHERE machine_name = ?", projectName, dut)
		if err != nil {
			tr.Rollback()
			logger.Error("Join project error: " + err.Error())
			return err
		}
	}
	tr.Commit()
	return nil
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
	result_url := fmt.Sprintf("https://%s:%s/csv/%s_mapping.zip", config.Viper.GetString("SERVER_IP"), config.Viper.GetString("SERVER_SOURCE_PORT"), currentDate)
	return result_url, nil
}
