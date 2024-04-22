package dbghost_api

import (
	"net/http"
	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"

	"github.com/gin-gonic/gin"
)

type Dbghostlist_Response struct {
	Ip []string `json:"ips"`
}

type Dbg_host struct {
	Ip       string `json:"ip"`
	Hostname string `json:"hostname"`
	Owner    string `json:"owner"`
}

func Dbghost_list(c *gin.Context) {
	extra := c.Query("extra")
	var Dbghost_list Dbghostlist_Response
	if extra == "empty" {
		rows, err := method.Query("SELECT ip FROM debug_host WHERE NOT EXISTS(SELECT 1 FROM debug_unit WHERE debug_host.ip=debug_unit.ip);")
		if err != nil {
			logger.Error("Query empty debug host list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Dbghost_list.Ip = append(Dbghost_list.Ip, tmp)
		}
	} else {
		rows, err := method.Query("SELECT ip FROM debug_host")
		if err != nil {
			logger.Error("Query debug host list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Dbghost_list.Ip = append(Dbghost_list.Ip, tmp)
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dbghost_list)
}

func Dbghost_freelist(c *gin.Context) {
	var Dbghost_list Dbghostlist_Response
	rows, err := method.Query("SELECT A.ip FROM debug_host A LEFT JOIN debug_unit C ON A.ip = C.ip WHERE C.ip IS NULL;")
	if err != nil {
		logger.Error("Query empty debug host list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		Dbghost_list.Ip = append(Dbghost_list.Ip, tmp)
	}
	if Dbghost_list.Ip == nil {
		tmp := []string{}
		Dbghost_list.Ip = tmp
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dbghost_list)
}

func Dbghost_all_info(c *gin.Context) {
	var Dbg_info_list []Dbg_host
	rows, err := method.Query("SELECT * FROM debug_host;")
	if err != nil {
		logger.Error("Search all debug host info error: " + err.Error())
	}
	for rows.Next() {
		var tmp Dbg_host
		err = rows.Scan(&tmp.Ip, &tmp.Hostname, &tmp.Owner)
		if err != nil {
			logger.Error("Search all debug host info error: " + err.Error())
		}
		Dbg_info_list = append(Dbg_info_list, tmp)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dbg_info_list)
}

func Dbghost_info(c *gin.Context) {
	ip := c.Query("ip")
	rows := method.QueryRow("SELECT * FROM debug_host WHERE ip=?", ip)
	var tmp Dbg_host
	err := rows.Scan(&tmp.Ip, &tmp.Hostname, &tmp.Owner)
	if err != nil {
		logger.Error("Search debug host info error: " + err.Error())
	}
	// response := ApiResponse{"200", Kvm_list}
	// c.JSON(http.StatusOK, response)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, tmp)
}

func Dbghost_search(c *gin.Context) {
	ip := c.Query("ip")
	var res apiservice.Debug_unit
	row := method.QueryRow("SELECT hostname, ip, machine_name FROM debug_unit WHERE ip=?", ip)
	err := row.Scan(&res.Hostname, &res.Ip, &res.Machine_name)
	if err != nil {
		logger.Error("Search dut mapping error" + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func Dbghost_modify(c *gin.Context) {
	ip := c.Query("ip")
	owner := c.Query("owner")
	_, err := method.Exec("UPDATE debug_host SET owner=? WHERE ip=?", owner, ip)
	if err != nil {
		logger.Error("Modify dbg_host info error" + err.Error())
		apiservice.ResponseWithJson(c.Writer, http.StatusNotFound, "")
		return
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
