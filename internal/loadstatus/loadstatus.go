package loadstatus

import (
	"recorder/pkg/mariadb/query"
	"time"
)

func Loadstatus() {
	query.Get_kvm_list()
}

func Sync_with_mariadb() {
	for {
		query.Update_kvm_list()
		time.Sleep(5 * time.Second)
	}
}
