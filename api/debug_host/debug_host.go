package dbghost_api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"recorder/pkg/mariadb/method"
	"recorder/pkg/logger"
	"recorder/pkg/apiservice"

)

type Dbghostlist_Response struct {
	Ip			string	`json:"ip"`
}

type Dbg_host struct{
	Ip			string	`json:"ip"`
	Hostname 	string	`json:"hostname"`
	Owner		string	`json:"owner"`
}

func Dbghost_list(c *gin.Context) {
	extra := c.Query("extra")
	var Dbghost_list []Dbghostlist_Response
	if extra == "empty"{
		rows, err := method.Query("select ip from debug_host where not exists(select 1 from debug_unit where debug_host.ip=debug_unit.ip);")
		if err != nil {
			logger.Error("Query empty debug host list error: " + err.Error())
		}
		for rows.Next() {
			var tmp Dbghostlist_Response
			err = rows.Scan(&tmp.Ip)
			Dbghost_list = append(Dbghost_list,tmp)
		}
	}else{
		rows, err := method.Query("SELECT ip from debug_host")
		if err != nil {
			logger.Error("Query debug host list error: " + err.Error())
		}
		for rows.Next() {
			var tmp Dbghostlist_Response
			err = rows.Scan(&tmp.Ip)
			Dbghost_list = append(Dbghost_list,tmp)
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dbghost_list)
}

func Dbghost_info(c *gin.Context) {
	var Dbg_host_list []Dbg_host
	rows, err := method.Query("SELECT * from debug_host")
	if err != nil {
		logger.Error("Query debug host list error: " + err.Error())
	}
	for rows.Next() {
		var tmp Dbg_host
		err = rows.Scan(&tmp.Ip, &tmp.Hostname, &tmp.Owner)
		if err != nil {
			logger.Error("Query debug host list error: " + err.Error())
		}
		Dbg_host_list = append(Dbg_host_list,tmp)
	}
	// response := ApiResponse{"200", Kvm_list}
	// c.JSON(http.StatusOK, response)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dbg_host_list)
}

func Dbghost_search(c *gin.Context) {
	ip := c.Query("ip")
	target := c.Query("target")
	var res string
	if target == "kvm" {
		row := method.QueryRow("select hostname from debug_unit where ip=?", ip)
		err := row.Scan(&res)
		if err != nil {
			res = "null"
		}
	}else if target == "dut" {
		row := method.QueryRow("select machine_name from debug_unit where ip=?", ip)
		err := row.Scan(&res)
		if err != nil {
			res = "null"
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}