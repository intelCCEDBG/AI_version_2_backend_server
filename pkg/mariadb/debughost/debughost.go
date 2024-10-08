package debughost_query

import (
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func CheckDebugHostIP(ip string) (bool, error) {
	row := method.QueryRow("SELECT hostname FROM debug_unit WHERE ip = ?", ip)
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
