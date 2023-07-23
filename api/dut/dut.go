package dut_api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"recorder/pkg/mariadb/method"
	"recorder/pkg/logger"
	"recorder/pkg/apiservice"

)

type Dutlist_Response struct {
	Machine_name			string	`json:"machine"`
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
	var Dut_list []Dutlist_Response
	if extra == "empty"{
		rows, err := method.Query("select machine_name from machine where not exists(select 1 from debug_unit where machine.machine_name=debug_unit.machine_name);")
		if err != nil {
			logger.Error("Query empty dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp Dutlist_Response
			err = rows.Scan(&tmp.Machine_name)
			Dut_list = append(Dut_list,tmp)
		}
	}else{
		rows, err := method.Query("SELECT machine_name from machine")
		if err != nil {
			logger.Error("Query dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp Dutlist_Response
			err = rows.Scan(&tmp.Machine_name)
			Dut_list = append(Dut_list,tmp)
		}		
	}

	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_list)
}

func Dut_info(c *gin.Context) {
	var Dut_list []Dut
	rows, err := method.Query("SELECT * from machine")
	if err != nil {
		logger.Error("Query dut list error: " + err.Error())
	}
	for rows.Next() {
		var tmp Dut
		err = rows.Scan(&tmp.Machine_name, &tmp.Ssim, &tmp.Status, &tmp.Cycle_cnt, &tmp.Error_timestamp, &tmp.Path)
		Dut_list = append(Dut_list,tmp)
	}
	// response := ApiResponse{"200", Kvm_list}
	// c.JSON(http.StatusOK, response)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_list)
}

func Dut_search(c *gin.Context) {
	machine_name := c.Query("machine")
	target := c.Query("target")
	var res string
	if target == "kvm" {
		row := method.QueryRow("select hostname from debug_unit where machine_name=?", machine_name)
		err := row.Scan(&res)
		if err != nil {
			res = "null"
		}
	}else if target == "dbghost" {
		row := method.QueryRow("select ip from debug_unit where machine_name=?", machine_name)
		err := row.Scan(&res)
		if err != nil {
			res = "null"
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}