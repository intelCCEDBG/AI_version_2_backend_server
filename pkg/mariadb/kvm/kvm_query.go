package kvm_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Update_kvm_status(hostname string, status string) {
	_, err := method.Exec("UPDATE kvm SET stream_status = ? WHERE hostname = ?", status, hostname)
	if err != nil {
		logger.Error("Update kvm status error: " + err.Error())
	}
}

func Get_kvm_status(hostname string) (kvm_template structure.Kvm) {
	KVM, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where hostname = " + "'" + hostname + "'")
	if err != nil {
		logger.Error("Query kvm " + hostname + " error: " + err.Error())
	}
	for KVM.Next() {
		err := KVM.Scan(&kvm_template.Hostname, &kvm_template.Stream_url, &kvm_template.Stream_status, &kvm_template.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}
	return kvm_template
}

func Get_recording_kvms() (kvms []structure.Kvm) {
	Recording_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'recording'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for Recording_list.Next() {
		var tmp = structure.Kvm{}
		err := Recording_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func Get_idle_kvms() (kvms []structure.Kvm) {
	Idle_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'idle'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for Idle_list.Next() {
		var tmp = structure.Kvm{}
		err := Idle_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func Get_all_kvms() (kvms []structure.Kvm) {
	All_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for All_list.Next() {
		var tmp = structure.Kvm{}
		err := All_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}
