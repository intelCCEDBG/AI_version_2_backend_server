package dbgunit_api

import (
	"net/http"
	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	unit_query "recorder/pkg/mariadb/unit"

	"github.com/gin-gonic/gin"
	// "fmt"
)

type Project_list_response struct {
	Project []string `json:"projects"`
}

type Debug_unit_info struct {
	Hostname      string `json:"hostname"`
	Machine_name  string `json:"machine_name"`
	KVM_link      string `json:"stream_url"`
	Record_status string `json:"record_status"`
	Lock_coord    string `json:"lock_coord"`
	Status        string `json:"status"`
	Debug_host    string `json:"debug_host"`
	Last_fail     string `json:"last_fail"`
}

type Project_info_response struct {
	Duts []Debug_unit_info `json:"duts"`
}

func Project_list(c *gin.Context) {
	var Project_list Project_list_response
	rows, err := method.Query("SELECT project_name FROM project;")
	if err != nil {
		logger.Error("Query project list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			logger.Error("Query project list error: " + err.Error())
		}
		Project_list.Project = append(Project_list.Project, tmp)
	}

	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Project_list)
}

func Project_info(c *gin.Context) {
	var Resp Project_info_response
	project_name := c.Query("project")
	if project_name == "ALL" {
		var Resp2 Project_info_response
		var Project_list []string
		rows, err := method.Query("SELECT DISTINCT project FROM debug_unit;")
		if err != nil {
			logger.Error("Search all project info error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			if err != nil {
				logger.Error("Search all project info error: " + err.Error())
			}
			Project_list = append(Project_list, tmp)
		}
		for _, pj := range Project_list {
			rows, err := method.Query("SELECT hostname, machine_name, ip FROM debug_unit WHERE project=?;", pj)
			if err != nil {
				logger.Error("Search project info error: " + err.Error())
			}
			for rows.Next() {
				var tmp Debug_unit_info
				err = rows.Scan(&tmp.Hostname, &tmp.Machine_name, &tmp.Debug_host)
				if err != nil {
					logger.Error("Search project info error: " + err.Error())
				}
				row := method.QueryRow("SELECT status FROM machine WHERE machine_name=?;", &tmp.Machine_name)
				err = row.Scan(&tmp.Status)
				if err != nil {
					logger.Error("Search project info error: " + err.Error())
				}
				row = method.QueryRow("SELECT stream_status, stream_url FROM kvm WHERE hostname=?;", &tmp.Hostname)
				err = row.Scan(&tmp.Record_status, &tmp.KVM_link)
				if err != nil {
					logger.Error("Search project info error: " + err.Error())
				}
				row = method.QueryRow("select time from errorlog where machine_name=? order by time desc limit 1", &tmp.Machine_name)
				err = row.Scan(&tmp.Last_fail)
				if err != nil {
					tmp.Last_fail = "N/A"
				}
				Resp2.Duts = append(Resp2.Duts, tmp)
			}
			// Resp.Duts = append(Resp.Duts, tmp2)
		}
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp2)
		return
	} else {
		rows, err := method.Query("SELECT hostname, machine_name, ip FROM debug_unit WHERE project=?;", project_name)
		if err != nil {
			logger.Error("Search project info error: " + err.Error())
		}
		var tmp2 []Debug_unit_info
		for rows.Next() {
			var tmp Debug_unit_info
			err = rows.Scan(&tmp.Hostname, &tmp.Machine_name, &tmp.Debug_host)
			if err != nil {
				logger.Error("Search project info error: " + err.Error())
			}
			row := method.QueryRow("SELECT status, lock_coord FROM machine WHERE machine_name=?;", &tmp.Machine_name)
			err = row.Scan(&tmp.Status, &tmp.Lock_coord)
			if err != nil {
				logger.Error("Search project info error: " + err.Error())
			}
			row = method.QueryRow("SELECT stream_status, stream_url FROM kvm WHERE hostname=?;", &tmp.Hostname)
			err = row.Scan(&tmp.Record_status, &tmp.KVM_link)
			if err != nil {
				logger.Error("Search project info error: " + err.Error())
			}
			row = method.QueryRow("select time from errorlog where machine_name=? order by time desc limit 1", &tmp.Machine_name)
			err = row.Scan(&tmp.Last_fail)
			if err != nil {
				tmp.Last_fail = "N/A"
			}
			tmp2 = append(tmp2, tmp)
		}
		Resp.Duts = tmp2
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
}

func Save_csv(c *gin.Context) {
	unit_query.ExportAllToCsv()
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "OK")
}
