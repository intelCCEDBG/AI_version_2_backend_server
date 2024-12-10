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

type DutListResponse struct {
	MachineName []string `json:"machines"`
}
type LockResponse struct {
	Locked int       `json:"locked"`
	Coord  []float64 `json:"coord"`
}
type Dut struct {
	MachineName    string `json:"machine"`
	Ssim           int    `json:"ssim"`
	Status         int    `json:"status"`
	CycleCnt       int    `json:"cycle_cnt"`
	ErrorTimestamp string `json:"error_timestamp"`
	Path           string `json:"path"`
	Threshold      int    `json:"threshold"`
}

func AddDut(c *gin.Context) {
	machineName := c.Query("dut")
	exists, err := dut_query.CheckDutExist(machineName)
	if err != nil {
		logger.Error("Check dut exists error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dut already exists"})
		return
	}
	err = dut_query.CreateDut(machineName)
	if err != nil {
		logger.Error("Create dut error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func DutList(c *gin.Context) {
	extra := c.Query("extra")
	var DutList DutListResponse
	if extra == "empty" {
		rows, err := method.Query("SELECT machine_name FROM machine WHERE NOT EXISTS(SELECT 1 FROM debug_unit WHERE machine.machine_name=debug_unit.machine_name);")
		if err != nil {
			logger.Error("Query dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			if err != nil {
				logger.Error("Query dut list error: " + err.Error())
			}
			DutList.MachineName = append(DutList.MachineName, tmp)
		}
	} else {
		rows, err := method.Query("SELECT machine_name FROM machine")
		if err != nil {
			logger.Error("Query dut list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			if err != nil {
				logger.Error("Query dut list error: " + err.Error())
			}
			DutList.MachineName = append(DutList.MachineName, tmp)
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, DutList)
}

func DutFreeList(c *gin.Context) {
	var DutList DutListResponse
	rows, err := method.Query("SELECT A.machine_name FROM machine A LEFT JOIN debug_unit C ON A.machine_name = C.machine_name WHERE C.machine_name IS NULL;")
	if err != nil {
		logger.Error("Query empty dut list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			logger.Error("Query empty dut list error: " + err.Error())
		}
		DutList.MachineName = append(DutList.MachineName, tmp)
	}
	if DutList.MachineName == nil {
		tmp := []string{}
		DutList.MachineName = tmp
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, DutList)
}

func DutAllInfo(c *gin.Context) {
	DutInfoList := dut_query.GetAllDutStatus()
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, DutInfoList)
}

func DutInfo(c *gin.Context) {
	machine := c.Query("machine")
	tmp := dut_query.GetDutStatus(machine)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, tmp)
}

func DutSearch(c *gin.Context) {
	machineName := c.Query("machine")
	var res apiservice.DebugUnit
	row := method.QueryRow("SELECT hostname, ip, machine_name FROM debug_unit WHERE machine_name=?", machineName)
	err := row.Scan(&res.Hostname, &res.Ip, &res.MachineName)
	if err != nil {
		logger.Error("Search dut mapping error" + err.Error())
	}
	res.Project = dut_query.GetProjectName(machineName)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func DutModify(c *gin.Context) {
	machineName := c.Query("machine")
	ssim := c.Query("ssim")
	threshold := c.Query("threshold")
	_, err := method.Exec("UPDATE machine SET ssim=?, threshold=? WHERE machine_name=?", ssim, threshold, machineName)
	if err != nil {
		logger.Error("Search dut mapping error" + err.Error())
		apiservice.ResponseWithJson(c.Writer, http.StatusNotFound, "")
		return
	}
	logpicqueue.RenewThreshold(machineName)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func DutLockCoord(c *gin.Context) {
	machineName := c.Query("machine")
	result := dut_query.GetAiResult(machineName)
	coordStr := structure.CoordF2S(result.Coords)
	logger.Debug(coordStr)
	dut_query.UpdateLockCoord(machineName, coordStr)
}

func DutUnlockCoord(c *gin.Context) {
	machineName := c.Query("machine")
	dut_query.UpdateLockCoord(machineName, "")
}

func DutIslocked(c *gin.Context) {
	machineName := c.Query("machine")
	dut := dut_query.GetDutStatus(machineName)
	var tmp LockResponse
	if dut.LockCoord == "" {
		tmp.Locked = 0
	} else {
		tmp.Locked = 1
		tmp.Coord = structure.CoordS2F(dut.LockCoord)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, tmp)
}

func DutStatus(c *gin.Context) {
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

func DutErrorlog(c *gin.Context) {
	machineName := c.Query("machine_name")
	res := errorlog_query.GetAllError(machineName)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func SetDutErrorlog(c *gin.Context) {
	var errorlog structure.Errorlog
	c.BindJSON(&errorlog)
	errorlog_query.SetErrorRecord(errorlog)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func DutDeleteErrorlog(c *gin.Context) {
	machineName := c.Query("machine_name")
	path := config.Viper.GetString("ERROR_VIDEO_PATH")
	res := errorlog_query.DeleteAllError(machineName)
	err := fileoperation.DeleteFiles(path + machineName + "/")
	if err != nil {
		logger.Error("Delete error video error: " + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func DutDeleteErrorlogProject(c *gin.Context) {
	project := c.Query("project")
	path := config.Viper.GetString("ERROR_VIDEO_PATH")
	duts := project_query.GetDuts(project)
	var rows int64
	rows = 0
	for _, dut := range duts {
		res := errorlog_query.DeleteAllError(dut.MachineName)
		rows += res
		err := fileoperation.DeleteFiles(path + dut.MachineName + "/")
		if err != nil {
			logger.Error("Delete error video error: " + err.Error())
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, rows)
}

func ProjectDutList(c *gin.Context) {
	var DutList DutListResponse
	projectName := c.Query("project")
	rows, err := method.Query("SELECT machine_name FROM debug_unit WHERE project=?;", projectName)
	if err != nil {
		logger.Error("Query dut list by project error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			logger.Error("Query dut list by project error: " + err.Error())
		}
		DutList.MachineName = append(DutList.MachineName, tmp)
	}
	if DutList.MachineName == nil {
		tmp := []string{}
		DutList.MachineName = tmp
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, DutList)
}

func SetDutMachineStatus(c *gin.Context) {
	var machineStatus structure.MachineStatus
	c.ShouldBind(&machineStatus)
	dut_query.SetMachineStatus(machineStatus)
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
	kvm := kvm_query.GetStatus(req.Hostname)
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

func CheckDutFree(c *gin.Context) {
	machineName := c.Query("dut")
	kvmHostname := c.Query("kvm")
	free, err := dut_query.CheckDutFree(machineName, kvmHostname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"free": free})
}
