package debughost_query

import (
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func CheckDebugHostIPExist(ip string) (bool, error) {
	row := method.QueryRow("SELECT hostname FROM debug_host WHERE ip = ?", ip)
	var hostname string
	err := row.Scan(&hostname)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		logger.Error("Check debug host ip error: " + err.Error())
		return false, err
	}
	return true, nil
}

func CreateDebugHost(ip, hostname, owner string) error {
	query := `
		INSERT INTO debug_host (ip, hostname, owner)
		VALUES (?, ?, ?)
	`
	_, err := method.Exec(query, ip, hostname, owner)
	if err != nil {
		logger.Error("Insert debug host error: " + err.Error())
		return err
	}
	return nil
}

func CheckDebugHostFree(dbgIp string, kvmHostname string) (bool, error) {
	// check if the debug host is paired with the another kvm
	row := method.QueryRow("SELECT hostname FROM debug_unit WHERE ip = ?", dbgIp)
	var hostname string
	err := row.Scan(&hostname)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return true, nil
		}
		logger.Error("Check debug host free error: " + err.Error())
		return false, err
	}
	if hostname != kvmHostname {
		return false, nil
	}
	return true, nil
}
