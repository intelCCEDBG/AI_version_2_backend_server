package query

import (
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Update_kvm_status(hostname string, status string) {
	_, err := method.Exec("UPDATE kvm SET stream_status = ? WHERE hostname = ?", status, hostname)
	if err != nil {
		logger.Error("Update kvm status error: " + err.Error())
	}
}
