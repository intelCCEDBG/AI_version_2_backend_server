package dut_api

import (
	"net/http"
	"recorder/internal/structure"
	apiservice "recorder/pkg/apiservice"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	"recorder/pkg/mariadb/method"

	"github.com/gin-gonic/gin"
)

type Dutlist_Response struct {
	Machine_name []string `json:"machines"`
}

type Dut struct {
	Machine_name    string `json:"machine"`
	Ssim            int    `json:"ssim"`
	Status          int    `json:"status"`
	Cycle_cnt       int    `json:"cycle_cnt"`
	Error_timestamp string `json:"error_timestamp"`
	Path            string `json:"path"`
	Threshold       int    `json:"threshold"`
}

func Dut_list(c *gin.Context) {
	extra := c.Query("extra")
	var Dut_list Dutlist_Response
	if extra == "empty" {
		rows, err := method.Query("SELECT machine_name FROM machine WHERE NOT EXISTS(SELECT 1 FROM debug_unit WHERE machine.machine_name=debug_unit.machine_name);")
		if err != nil {
			logger.Error("Query empty dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Dut_list.Machine_name = append(Dut_list.Machine_name, tmp)
		}
	} else {
		rows, err := method.Query("SELECT machine_name FROM machine")
		if err != nil {
			logger.Error("Query dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Dut_list.Machine_name = append(Dut_list.Machine_name, tmp)
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
		err = rows.Scan(&tmp.Machine_name, &tmp.Ssim, &tmp.Status, &tmp.Cycle_cnt, &tmp.Error_timestamp, &tmp.Path, &tmp.Threshold)
		if err != nil {
			logger.Error("Search all dut info error: " + err.Error())
		}
		Dut_info_list = append(Dut_info_list, tmp)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_info_list)
}

func Dut_info(c *gin.Context) {
	machine := c.Query("machine")
	rows := method.QueryRow("SELECT * FROM machine WHERE machine_name=?", machine)
	var tmp Dut
	err := rows.Scan(&tmp.Machine_name, &tmp.Ssim, &tmp.Status, &tmp.Cycle_cnt, &tmp.Error_timestamp, &tmp.Path, &tmp.Threshold)
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

func Dut_modify(c *gin.Context) {
	machine_name := c.Query("machine")
	ssim := c.Query("ssim")
	threshold := c.Query("threshold")
	_, err := method.Exec("UPDATE machine SET ssim=?, threshold=? WHERE machine_name=?", ssim, threshold, machine_name)
	if err != nil {
		logger.Error("Search dut mapping error" + err.Error())
		apiservice.ResponseWithJson(c.Writer, http.StatusNotFound, "")
		return
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Dut_lock_coord(c *gin.Context) {
	machine_name := c.Query("machine")
	result := dut_query.Get_AI_result(machine_name)
	coord_str := structure.Coord_f2s(result.Coords)
	dut_query.Update_lock_coord(machine_name, coord_str)
}
func Dut_unlock_coord(c *gin.Context) {
	machine_name := c.Query("machine")
	dut_query.Update_lock_coord(machine_name, "")
}
func Dut_status(c *gin.Context) {
	hostname := c.Query("hostname")
	status := c.Query("status")
	_, err := method.Exec("UPDATE machine SET status=? WHERE machine_name = (SELECT machine_name FROM debug_unit WHERE hostname=?)", status, hostname)
	if err != nil {
		logger.Error("update dut status error" + err.Error())
		apiservice.ResponseWithJson(c.Writer, http.StatusNotFound, "")
		return
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
