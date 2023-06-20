package loadstatus

import "recorder/pkg/mariadb/query"

func Loadstatus() {
	query.Get_machine_list()
}
