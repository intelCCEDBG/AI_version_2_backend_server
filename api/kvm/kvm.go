package kvm_api

import (
	// "fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apiservice "recorder/pkg/apiservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
}

type Kvmlist_Response struct {
	Hostname []string `json:"hostnames"`
}

type Kvm struct {
	Hostname          string `json:"hostname"`
	Ip                string `json:"ip"`
	Owner             string `json:"owner"`
	Status            string `json:"status"`
	Version           string `json:"version"`
	NAS_ip            string `json:"nas_ip"`
	Stream_url        string `json:"stream_url"`
	Stream_status     string `json:"stream_status"`
	Stream_interface  string `json:"stream_interface"`
	Start_record_time string `json:"start_record_time"`
}

func Kvm_list(c *gin.Context) {
	extra := c.Query("extra")
	var Kvm_list Kvmlist_Response
	if extra == "empty" {
		rows, err := method.Query("select hostname from kvm where not exists(select 1 from debug_unit where kvm.hostname=debug_unit.hostname)")
		if err != nil {
			logger.Error("Query empty kvm list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Kvm_list.Hostname = append(Kvm_list.Hostname, tmp)
		}
	} else {
		rows, err := method.Query("SELECT hostname FROM kvm")
		if err != nil {
			logger.Error("Query kvm list error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Kvm_list.Hostname = append(Kvm_list.Hostname, tmp)
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Kvm_list)
}

func Kvm_info(c *gin.Context) {
	hostname := c.Query("hostname")
	rows := method.QueryRow("SELECT * FROM kvm WHERE hostname=?",hostname)
	var tmp Kvm
	err := rows.Scan(&tmp.Hostname, &tmp.Ip, &tmp.Owner, &tmp.Status, &tmp.Version, &tmp.NAS_ip, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface, &tmp.Start_record_time)
	if err != nil {
		logger.Error("Search kvm info error: " + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, tmp)
}

func Kvm_search(c *gin.Context) {
	hostname := c.Query("hostname")
	var res apiservice.Debug_unit
	row := method.QueryRow("select hostname, ip, machine_name, project from debug_unit where hostname=?", hostname)
	err := row.Scan(&res.Hostname, &res.Ip, &res.Machine_name, &res.Project)
	if err != nil {
		logger.Error("Search kvm mapping error" + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func Kvm_mapping(c *gin.Context) {
	// var Kvm_mapping_list	[]Debug_unit
	body, err := ioutil.ReadAll(c.Request.Body)
	var Req apiservice.Debug_unit
	_ = json.Unmarshal(body, &Req)
	uuid := uuid.New().String()
	row := method.QueryRow("SELECT count(*) FROM debug_unit where hostname=?", Req.Hostname)
	var exist int
	err = row.Scan(&exist)
	if err != nil {
		logger.Error("update kvm mapping error: " + err.Error())
	}
	if exist != 0 {
		_, err := method.Exec("DELETE FROM debug_unit WHERE hostname=?", Req.Hostname)
		if err != nil {
			logger.Error("update kvm mapping error: " + err.Error())
		}
	}
	row = method.QueryRow("SELECT count(*) FROM debug_unit where ip=?", Req.Ip)
	var exist2 int
	err = row.Scan(&exist2)
	if err != nil {
		logger.Error("update kvm mapping error: " + err.Error())
	}
	if exist2 != 0 {
		_, err := method.Exec("DELETE FROM debug_unit WHERE ip=?", Req.Ip)
		if err != nil {
			logger.Error("update kvm mapping error: " + err.Error())
		}
	}
	row = method.QueryRow("SELECT count(*) FROM debug_unit where machine_name=?", Req.Machine_name)
	var exist3 int
	err = row.Scan(&exist3)
	if err != nil {
		logger.Error("update kvm mapping error: " + err.Error())
	}
	if exist3 != 0 {
		_, err := method.Exec("DELETE FROM debug_unit WHERE machine_name=?", Req.Machine_name)
		if err != nil {
			logger.Error("update kvm mapping error: " + err.Error())
		}
	}
	_, err = method.Exec("INSERT INTO debug_unit ( uuid, hostname, ip, machine_name, project) VALUES (?, ?, ?, ?, ?);", uuid, Req.Hostname, Req.Ip, Req.Machine_name, Req.Project)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

// func Kvm_csv_mapping(c *gin.Context) {
// 	c.Request.ParseMultipartForm(10 << 20)
// 	file, handler, err := c.Request.FormFile("file")
// 	if err != nil{
// 		logger.Error("Error Data retrieving " + err.Error())
// 	}
// 	fileName  := handler.Filename
// 	logger.Info("Uploaded File : %+v\n", fileName)
// 	tempFile , err := ioutil.TempFile("csv/temp-pdfs", "upload.csv")
// 	if err != nil{
// 		logger.Error(err.Error())
// 	}
// 	defer tempFile.Close()
// 	logger.Info("File name %s\n", tempFile.Name())
// 	var Req Debug_unit
// 	// _ = json.Unmarshal(body, &Req)
// 	uuid := uuid.New().String()
// 	row := method.QueryRow("SELECT count(*) FROM debug_unit where hostname=?", Req.Hostname)
// 	var exist int
// 	err = row.Scan(&exist)
// 	if err != nil {
// 		logger.Error("update kvm mapping error: " + err.Error())
// 	}
// 	if exist != 0 {
// 		_, err := method.Exec("DELETE FROM debug_unit WHERE hostname=?", Req.Hostname)
// 		if err != nil {
// 			logger.Error("update kvm mapping error: " + err.Error())
// 		}
// 	}
// 	row = method.QueryRow("SELECT count(*) FROM debug_unit where ip=?", Req.Ip)
// 	var exist2 int
// 	err = row.Scan(&exist2)
// 	if err != nil {
// 		logger.Error("update kvm mapping error: " + err.Error())
// 	}
// 	if exist2 != 0 {
// 		_, err := method.Exec("DELETE FROM debug_unit WHERE ip=?", Req.Ip)
// 		if err != nil {
// 			logger.Error("update kvm mapping error: " + err.Error())
// 		}
// 	}
// 	row = method.QueryRow("SELECT count(*) FROM debug_unit where machine_name=?", Req.Machine_name)
// 	var exist3 int
// 	err = row.Scan(&exist3)
// 	if err != nil {
// 		logger.Error("update kvm mapping error: " + err.Error())
// 	}
// 	if exist3 != 0 {
// 		_, err := method.Exec("DELETE FROM debug_unit WHERE machine_name=?", Req.Machine_name)
// 		if err != nil {
// 			logger.Error("update kvm mapping error: " + err.Error())
// 		}
// 	}
// 	_, err = method.Exec("INSERT INTO debug_unit ( uuid, hostname, ip, machine_name, project) VALUES (?, ?, ?, ?, ?);", uuid, Req.Hostname, Req.Ip, Req.Machine_name, Req.Project)
// 	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
// }
