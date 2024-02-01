package unit_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Get_unitbyhostname(hostname string) structure.Unit {
	var unit_template structure.Unit
	UNIT, err := method.Query("SELECT machine_name,ip,hostname,project FROM debug_unit where hostname = " + "'" + hostname + "'")
	if err != nil {
		logger.Error("Query Unit " + hostname + " error: " + err.Error())
	}
	for UNIT.Next() {
		err := UNIT.Scan(&unit_template.Machine_name, &unit_template.Ip, &unit_template.Hostname, &unit_template.Project)
		if err != nil {
			logger.Error(err.Error())
			return unit_template
		}
	}
	return unit_template
}
