package dut_api

import (
	"fmt"
	"net/http"
	"recorder/config"
	ai "recorder/internal/AI"
	"recorder/internal/logpicqueue"
	"recorder/internal/structure"
	apiservice "recorder/pkg/apiservice"
	"recorder/pkg/fileoperation"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	errorlog_query "recorder/pkg/mariadb/errrorlog"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/mariadb/method"
	project_query "recorder/pkg/mariadb/project"

	"github.com/gin-gonic/gin"
)

type Dutlist_Response struct {
	Machine_name []string `json:"machines"`
}
type Lock_Response struct {
	Locked int       `json:"locked"`
	Coord  []float64 `json:"coord"`
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
func Dut_freelist(c *gin.Context) {
	var Dut_list Dutlist_Response
	rows, err := method.Query("SELECT A.machine_name FROM machine A LEFT JOIN debug_unit C ON A.machine_name = C.machine_name WHERE C.machine_name IS NULL;")
	if err != nil {
		logger.Error("Query empty dut list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		Dut_list.Machine_name = append(Dut_list.Machine_name, tmp)
	}
	if Dut_list.Machine_name == nil {
		tmp := []string{}
		Dut_list.Machine_name = tmp
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_list)
}
func Dut_all_info(c *gin.Context) {
	Dut_info_list := dut_query.Get_all_dut_status()
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_info_list)
}

func Dut_info(c *gin.Context) {
	machine := c.Query("machine")
	tmp := dut_query.GetDutStatus(machine)
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
	res.Project = dut_query.GetProjectName(machine_name)
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
	logpicqueue.RenewThreshold(machine_name)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Dut_lock_coord(c *gin.Context) {
	machine_name := c.Query("machine")
	result := dut_query.GetAIResult(machine_name)
	coord_str := structure.CoordF2S(result.Coords)
	logger.Debug(coord_str)
	dut_query.Update_lock_coord(machine_name, coord_str)
}
func Dut_unlock_coord(c *gin.Context) {
	machine_name := c.Query("machine")
	dut_query.Update_lock_coord(machine_name, "")
}
func Dut_islocked(c *gin.Context) {
	machine_name := c.Query("machine")
	dut := dut_query.GetDutStatus(machine_name)
	var tmp Lock_Response
	if dut.LockCoord == "" {
		tmp.Locked = 0
	} else {
		tmp.Locked = 1
		tmp.Coord = structure.CoordS2F(dut.LockCoord)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, tmp)
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
func Dut_errorlog(c *gin.Context) {
	machine_name := c.Query("machine_name")
	res := errorlog_query.Get_all_error(machine_name)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}
func Set_dut_errorlog(c *gin.Context) {
	var errorlog structure.Errorlog
	c.BindJSON(&errorlog)
	errorlog_query.Set_error_record(errorlog)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Dut_deleteerrorlog(c *gin.Context) {
	machine_name := c.Query("machine_name")
	path := config.Viper.GetString("ERROR_VIDEO_PATH")
	res := errorlog_query.Delete_all_error(machine_name)
	err := fileoperation.DeleteFiles(path + machine_name + "/")
	if err != nil {
		logger.Error("Delete error video error: " + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}
func Dut_deleteerrorlog_project(c *gin.Context) {
	project := c.Query("project")
	path := config.Viper.GetString("ERROR_VIDEO_PATH")
	duts := project_query.Get_duts(project)
	var rows int64
	rows = 0
	for _, dut := range duts {
		res := errorlog_query.Delete_all_error(dut.MachineName)
		rows += res
		err := fileoperation.DeleteFiles(path + dut.MachineName + "/")
		if err != nil {
			logger.Error("Delete error video error: " + err.Error())
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, rows)
}
func Project_dut_list(c *gin.Context) {
	var Dut_list Dutlist_Response
	project_name := c.Query("project")
	rows, err := method.Query("SELECT machine_name FROM debug_unit WHERE project=?;", project_name)
	if err != nil {
		logger.Error("Query dut list by project error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		Dut_list.Machine_name = append(Dut_list.Machine_name, tmp)
	}
	if Dut_list.Machine_name == nil {
		tmp := []string{}
		Dut_list.Machine_name = tmp
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Dut_list)
}
func Set_dut_machine_status(c *gin.Context) {
	var machine_status structure.MachineStatus
	c.ShouldBind(&machine_status)
	dut_query.Set_machine_status(machine_status)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

type freezeCheckRequest struct {
	MachineName string    `json:"dut"`
	Hostname    string    `json:"kvm"`
	Coords      []float64 `json:"coords"`
}

func FreezeCheck(c *gin.Context) {
	var req freezeCheckRequest
	c.BindJSON(&req)
	logger.Info("FreezeCheck request: " + fmt.Sprintf("%+v", req))
	kvm := kvm_query.GetKvmStatus(req.Hostname)
	freezed, err := ai.FreezeCheck(req.MachineName, req.Coords, kvm)
	if err != nil {
		apiservice.ResponseWithJson(c.Writer, http.StatusInternalServerError, err.Error())
		return
	}
	msg := ""
	if freezed {
		msg = "Freezed"
	} else {
		msg = "Not Freezed"
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, msg)
}
