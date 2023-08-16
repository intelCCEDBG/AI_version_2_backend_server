package kvm_api

import (
	"fmt"
	"sort"
	"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
}

type Kvmlist_Response struct {
	Hostname string `json:"hostname"`
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

type Debug_unit struct {
	Hostname     string `json:"kvm_hostname"`
	Ip           string `json:"dbghost_ip"`
	Machine_name string `json:"dut_machine"`
	Project      string `json:"project"`
}
type Video_info struct {
	Hostname     string `json:"kvm_hostname"`
	Hour     int `json:"hour"`
	Minute           int `json:"minute"`
	Duration int `json:"duration"`
}
func Kvm_list(c *gin.Context) {
	extra := c.Query("extra")
	var Kvm_list []Kvmlist_Response
	if extra == "empty" {
		rows, err := method.Query("select hostname from kvm where not exists(select 1 from debug_unit where kvm.hostname=debug_unit.hostname)")
		if err != nil {
			logger.Error("Query empty kvm list error: " + err.Error())
		}
		for rows.Next() {
			var tmp Kvmlist_Response
			err = rows.Scan(&tmp.Hostname)
			Kvm_list = append(Kvm_list, tmp)
		}
	} else {
		rows, err := method.Query("SELECT hostname FROM kvm")
		if err != nil {
			logger.Error("Query kvm list error: " + err.Error())
		}
		for rows.Next() {
			var tmp Kvmlist_Response
			err = rows.Scan(&tmp.Hostname)
			Kvm_list = append(Kvm_list, tmp)
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Kvm_list)
}

func Kvm_info(c *gin.Context) {
	var Kvm_list []Kvm
	rows, err := method.Query("SELECT * FROM kvm")
	if err != nil {
		logger.Error("Query kvm info error: " + err.Error())
	}
	for rows.Next() {
		var tmp Kvm
		err = rows.Scan(&tmp.Hostname, &tmp.Ip, &tmp.Owner, &tmp.Status, &tmp.Version, &tmp.NAS_ip, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface, &tmp.Start_record_time)
		if err != nil {
			logger.Error("Query kvm info error: " + err.Error())
		}
		Kvm_list = append(Kvm_list, tmp)
	}
	// response := ApiResponse{"200", Kvm_list}
	// c.JSON(http.StatusOK, response)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Kvm_list)
}

func Kvm_search(c *gin.Context) {
	hostname := c.Query("hostname")
	target := c.Query("target")
	var res string
	if target == "dut" {
		row := method.QueryRow("select machine_name from debug_unit where hostname=?", hostname)
		err := row.Scan(&res)
		if err != nil {
			res = "null"
		}
	} else if target == "dbghost" {
		row := method.QueryRow("select ip from debug_unit where hostname=?", hostname)
		err := row.Scan(&res)
		if err != nil {
			res = "null"
		}
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func Kvm_mapping(c *gin.Context) {
	// var Kvm_mapping_list	[]Debug_unit
	body, err := ioutil.ReadAll(c.Request.Body)
	var Req Debug_unit
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
func Kvm_genvideo(c *gin.Context){
	body, _ := ioutil.ReadAll(c.Request.Body)
	var Req Video_info
	_ = json.Unmarshal(body, &Req)
	fmt.Println(Req.Duration)
	files, err := ioutil.ReadDir("/home/media/video/"+Req.Hostname+"/")
    if err != nil {
        logger.Error("List video dir fail: " + err.Error())
    }
	
	var filenames []string
	for _, file := range files {
		filenames = append(filenames,file.Name())
    }
	sort.Strings(filenames)
	f, err := os.OpenFile("/home/media/video/"+Req.Hostname+"/self-define.m3u8",os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 644)
	if err != nil {
        logger.Error("open video file fail: " + err.Error())
		return
    }
	defer f.Close()
	f.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-ALLOW-CACHE:YES\n#EXT-X-TARGETDURATION:10\n")
	var h,m string
	if (Req.Hour<10){
		h = strconv.Itoa(Req.Hour)
		h = "0"+h
	}else{
		h = strconv.Itoa(Req.Hour)
	}
	if (Req.Minute<10){
		m = strconv.Itoa(Req.Minute)
		m = "0"+m
	}else{
		m = strconv.Itoa(Req.Minute)
	}
	for ii, filename := range filenames {
		if (strings.Contains(filename,"2023-08-15_"+h+"-"+m)){
			for i:=0; i<Req.Duration/10;i++{
				f.WriteString("#EXTINF:10.000000,\n")
				f.WriteString(filenames[ii+i]+"\n")
			}
			break
		}
    }
	f.WriteString("#EXT-X-ENDLIST")
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}