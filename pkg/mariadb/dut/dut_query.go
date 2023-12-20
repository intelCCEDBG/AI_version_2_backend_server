package dut_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Update_dut_status(machine_name string, status string) {
	_, err := method.Exec("UPDATE machine SET status = ? WHERE machine_name = ?", status, machine_name)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}
func Update_dut_cnt(machine_name string, cnt int) {
	_, err := method.Exec("UPDATE machine SET cycle_cnt = ? WHERE machine_name = ?", cnt, machine_name)
	if err != nil {
		logger.Error("Update DUT status error: " + err.Error())
	}
}
func Get_dut_status(machine_name string) (dut_template structure.DUT) {
	KVM, err := method.Query("SELECT machine_name,ssim,cycle_cnt,status,threshhold FROM machine where machine_name = " + "'" + machine_name + "'")
	if err != nil {
		logger.Error("Query DUT " + machine_name + " error: " + err.Error())
	}
	for KVM.Next() {
		err := KVM.Scan(&dut_template.Machine_name, &dut_template.Ssim, &dut_template.Cycle_cnt, &dut_template.Status, &dut_template.Threshhold)
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}
	return dut_template
}
