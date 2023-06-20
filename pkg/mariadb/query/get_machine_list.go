package query

import (
	"recorder/internal/machine"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Get_machine_list() {
	Idle_list, err := method.Query("SELECT * FROM machine where stream_status = 'idle'")
	if err != nil {
		logger.Error("Query idle machine error: " + err.Error())
	}
	for Idle_list.Next() {
		var tmp machine.Machine
		err := Idle_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_pid, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		machine.Idle_machine[tmp.Hostname] = tmp
		machine.All_machine[tmp.Hostname] = tmp
	}
	Recording_list, err := method.Query("SELECT * FROM machine where stream_status = 'recording'")
	if err != nil {
		logger.Error("Query idle machine error: " + err.Error())
	}
	for Recording_list.Next() {
		var tmp machine.Machine
		err := Recording_list.Scan(&tmp.Hostname, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_pid, &tmp.Stream_interface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		machine.Recording_machine[tmp.Hostname] = tmp
		machine.All_machine[tmp.Hostname] = tmp
	}
}
