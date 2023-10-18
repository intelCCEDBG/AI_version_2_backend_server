package query

import (
	"recorder/internal/kvm"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func init() {
	kvm.Recording_kvm = make(map[string]*kvm.Kvm)
	kvm.Idle_kvm = make(map[string]*kvm.Kvm)
	kvm.All_kvm = make(map[string]*kvm.Kvm)
}
func Get_kvm_list() {
	Idle_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'idle'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for Idle_list.Next() {
		var tmp = new(kvm.Kvm)
		err := Idle_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvm.Idle_kvm[tmp.Hostname] = tmp
		kvm.All_kvm[tmp.Hostname] = tmp
	}
	Recording_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'recording'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for Recording_list.Next() {
		var tmp = new(kvm.Kvm)
		err := Recording_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvm.Recording_kvm[tmp.Hostname] = tmp
		kvm.All_kvm[tmp.Hostname] = tmp
	}
}

func Update_kvm_list() {
	for _, element := range kvm.All_kvm {
		All_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where hostname = " + "'" + element.Hostname + "'")
		if err != nil {
			logger.Error("Query kvm " + element.Hostname + " error: " + err.Error())
		}
		for All_list.Next() {
			var tmp = kvm.Kvm{}
			err := All_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			if tmp == *element {
				continue
			} else {
				kvm.All_kvm[tmp.Hostname].Stream_status = tmp.Stream_status
				kvm.All_kvm[tmp.Hostname].Stream_url = tmp.Stream_url
				kvm.All_kvm[tmp.Hostname].Stream_interface = tmp.Stream_interface
				if tmp.Stream_status == "idle" {
					kvm.Idle_kvm[tmp.Hostname] = element
					delete(kvm.Recording_kvm, tmp.Hostname)
				} else if tmp.Stream_status == "recording" {
					kvm.Recording_kvm[tmp.Hostname] = element
					delete(kvm.Idle_kvm, tmp.Hostname)
				}
			}
		}
	}
}

func Update_kvm_status(hostname string, status string) {
	_, err := method.Exec("UPDATE kvm SET stream_status = ? WHERE hostname = ?", status, hostname)
	if err != nil {
		logger.Error("Update kvm status error: " + err.Error())
	}
}
