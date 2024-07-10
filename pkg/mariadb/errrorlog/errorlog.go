package errorlog_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Get_all_error(machine_name string) []structure.Errorlog {
	var res []structure.Errorlog
	Result, err := method.Query("SELECT time, type, test_item, sku, image, bios, uuid FROM errorlog where machine_name = ?", machine_name)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var Log structure.Errorlog
	for Result.Next() {
		err := Result.Scan(&Log.Time, &Log.Type, &Log.Test_item, &Log.Sku, &Log.Image, &Log.Bios, &Log.Uuid)
		if err != nil {
			logger.Error(err.Error())
		}
		Log.Machine_name = machine_name
		res = append(res, Log)
	}
	return res
}
func Delete_all_error(machine_name string) int64 {
	Result, err := method.Exec("DELETE FROM errorlog where machine_name = ?", machine_name)
	if err != nil {
		logger.Error("Delete AI result error: " + err.Error())
	}
	affected, _ := Result.RowsAffected()
	return affected
}
func Set_error_record(errorlog structure.Errorlog) {
	_, err := method.Exec("INSERT INTO errorlog (time, type, test_item, sku, image, bios, uuid, machine_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", errorlog.Time, errorlog.Type, errorlog.Test_item, errorlog.Sku, errorlog.Image, errorlog.Bios, errorlog.Uuid, errorlog.Machine_name)
	if err != nil {
		logger.Error("Insert error log error" + err.Error())
	}
}
