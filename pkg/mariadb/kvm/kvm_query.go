package kvm_query

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

func Get_kvm_status(hostname string) {
	KVM, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where hostname = " + "'" + hostname + "'")
	if err != nil {
		logger.Error("Query kvm " + hostname + " error: " + err.Error())
	}
	for KVM.Next() {

	}
}
