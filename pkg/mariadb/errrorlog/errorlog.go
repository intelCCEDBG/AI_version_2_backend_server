package errorlog_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Get_all_error(machine_name string) []structure.Errorlog {
	var res []structure.Errorlog
	Result, err := method.Query("SELECT time, type, test_item, sku, image, bios FROM errorlog where machine_name = ?", machine_name)
	if err != nil {
		logger.Error("Update AI result error: " + err.Error())
	}
	var Log structure.Errorlog
	for Result.Next() {
		err := Result.Scan(&Log.Time, &Log.Type, &Log.Test_item, &Log.Sku, &Log.Image, &Log.Bios)
		if err != nil {
			logger.Error(err.Error())
		}
		res = append(res, Log)
	}
	return res
}
