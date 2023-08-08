package dut_api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"recorder/pkg/mariadb/method"
	"recorder/pkg/logger"
	apiservice "recorder/pkg/apiservice"

)

type Dutlist_Response struct {
	Machine_name		[]string	`json:"machines"`
}

type Dut struct{
	Machine_name		string	`json:"machine"`
	Ssim 				string	`json:"ssim"`
	Status				string	`json:"status"`
	Cycle_cnt			string	`json:"cycle_cnt"`
	Error_timestamp 	string	`json:"error_timestamp"`
	Path           		string	`json:"path"`
}

func Dut_list(c *gin.Context) {
	extra := c.Query("extra")
	var Dut_list Dutlist_Response
	if extra == "empty"{
		rows, err := method.Query("SELECT machine_name FROM machine WHERE NOT EXISTS(SELECT 1 FROM debug_unit WHERE machine.machine_name=debug_unit.machine_name);")
		if err != nil {
			logger.Error("Query empty dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Dut_list.Machine_name = append(Dut_list.Machine_name,tmp)
		}
	}else{
		rows, err := method.Query("SELECT machine_name FROM machine")
		if err != nil {
			logger.Error("Query dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Dut_list.Machine_name = append(Dut_list.Machine_name,tmp)
		}		
	}

	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_list)
}

func Dut_all_info(c *gin.Context) {
	var Dut_info_list []Dut
	rows, err := method.Query("SELECT * FROM machine;")
	if err != nil {
		logger.Error("Search all dut info error: " + err.Error())
	}
	for rows.Next() {
		var tmp Dut
		err = rows.Scan(&tmp.Machine_name, &tmp.Ssim, &tmp.Status, &tmp.Cycle_cnt, &tmp.Error_timestamp, &tmp.Path)
		if err != nil {
			logger.Error("Search all dut info error: " + err.Error())
		}
		Dut_info_list = append(Dut_info_list, tmp)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_info_list)
}

func Dut_info(c *gin.Context) {
	machine := c.Query("machine")
	rows := method.QueryRow("SELECT * FROM machine WHERE machine_name=?",machine)
	var tmp Dut
	err := rows.Scan(&tmp.Machine_name, &tmp.Ssim, &tmp.Status, &tmp.Cycle_cnt, &tmp.Error_timestamp, &tmp.Path)
	if err != nil {
		logger.Error("Search dut information error: " + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, tmp)
}

func Dut_search(c *gin.Context) {
	machine_name := c.Query("machine")
	var res apiservice.Debug_unit
	row := method.QueryRow("SELECT hostname, ip, machine_name FROM debug_unit WHERE machine_name=?", machine_name)
	err := row.Scan(&res.Hostname, &res.Ip, &res.Machine_name)
	if err != nil {
		logger.Error("Search dut mapping error" + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}