package kvm_api

import (
	"fmt"

	"sort"

	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"strconv"
	"strings"

	"golang.org/x/exp/slices"

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
	Hostname []string `json:"hostnames"`
}

type Kvm_del_Req struct {
	Kvm_hostname string `json:"kvm_hostname"`
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

type Mapping_file struct {
	Hostname     string
	Ip           string
	Machine_name string
	Project      string
}

type Video_info struct {
	Hostname string `json:"kvm_hostname"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`
	Duration int    `json:"duration"`
}

type Kvm_state struct {
	Hostname      string `json:"kvm_hostname"`
	Stream_status string `json:"stream_status"`
}

func Kvm_list(c *gin.Context) {
	extra := c.Query("extra")
	var Kvm_list Kvmlist_Response
	if extra == "empty" {
		rows, err := method.Query("SELECT hostname FROM kvm WHERE NOT EXISTS(SELECT 1 FROM debug_unit WHERE kvm.hostname=debug_unit.hostname)")
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

func Kvm_all_info(c *gin.Context) {
	var Kvm_list []Kvm
	rows, err := method.Query("SELECT * FROM kvm;")
	if err != nil {
		logger.Error("Search all kvm info error: " + err.Error())
	}
	for rows.Next() {
		var tmp Kvm
		err = rows.Scan(&tmp.Hostname, &tmp.Ip, &tmp.Owner, &tmp.Status, &tmp.Version, &tmp.NAS_ip, &tmp.Stream_url, &tmp.Stream_status, &tmp.Stream_interface, &tmp.Start_record_time)
		if err != nil {
			logger.Error("Search all kvm info error: " + err.Error())
		}
		Kvm_list = append(Kvm_list, tmp)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Kvm_list)
}

func Kvm_info(c *gin.Context) {
	hostname := c.Query("hostname")
	rows := method.QueryRow("SELECT * FROM kvm WHERE hostname=?", hostname)
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
	row := method.QueryRow("SELECT hostname, ip, machine_name FROM debug_unit WHERE hostname=?", hostname)
	err := row.Scan(&res.Hostname, &res.Ip, &res.Machine_name)
	if err != nil {
		logger.Error("Search kvm mapping error" + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, res)
}

func Kvm_mapping(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Read kvm mapping request error: " + err.Error())
	}
	var Req apiservice.Debug_unit
	err = json.Unmarshal(body, &Req)
	if err != nil {
		logger.Error("Parse kvm mapping request error: " + err.Error())
	}
	uuid := uuid.New().String()
	row := method.QueryRow("SELECT count(*) FROM debug_unit WHERE hostname=?", Req.Hostname)
	var exist int
	err = row.Scan(&exist)
	if err != nil {
		logger.Error("update kvm mapping error: " + err.Error())
	}
	if exist != 0 {
		// _, err := method.Exec("DELETE FROM debug_unit WHERE hostname=?", Req.Hostname)
		apiservice.ResponseWithJson(c.Writer, http.StatusForbidden, "")
		return
	}
	row = method.QueryRow("SELECT count(*) FROM debug_unit WHERE ip=?", Req.Ip)
	var exist2 int
	err = row.Scan(&exist2)
	if err != nil {
		logger.Error("update kvm mapping error: " + err.Error())
	}
	if exist2 != 0 {
		// _, err := method.Exec("DELETE FROM debug_unit WHERE ip=?", Req.Ip)
		apiservice.ResponseWithJson(c.Writer, http.StatusForbidden, "")
		return
	}
	row = method.QueryRow("SELECT count(*) FROM debug_unit WHERE machine_name=?", Req.Machine_name)
	var exist3 int
	err = row.Scan(&exist3)
	if err != nil {
		logger.Error("update kvm mapping error: " + err.Error())
	}
	if exist3 != 0 {
		// _, err := method.Exec("DELETE FROM debug_unit WHERE machine_name=?", Req.Machine_name)
		apiservice.ResponseWithJson(c.Writer, http.StatusForbidden, "")
		return
	}
	_, err = method.Exec("INSERT INTO debug_unit ( uuid, hostname, ip, machine_name) VALUES (?, ?, ?, ?);", uuid, Req.Hostname, Req.Ip, Req.Machine_name)
	if err != nil {
		logger.Error("insert kvm mapping error: " + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func Kvm_delete(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Update kvm mapping error: " + err.Error())
	}
	var Req Kvm_del_Req
	err = json.Unmarshal(body, &Req)
	if err != nil {
		logger.Error("Update kvm mapping error: " + err.Error())
	}
	_, err = method.Exec("DELETE FROM debug_unit WHERE hostname=?", Req.Kvm_hostname)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func checkMappingFile(data [][]string) (bool, string) {
	var check_hostname []string
	var check_ip []string
	var check_machine []string
	for i, line := range data {
		if i > 0 { // omit header line
			for j, field := range line {
				if j == 0 {
					check_hostname = append(check_hostname, field)
				} else if j == 1 {
					check_ip = append(check_ip, field)
				} else if j == 2 {
					check_machine = append(check_machine, field)
				}
			}
		}
	}
	allKeys := make(map[string]bool)
	for _, item := range check_hostname {
		if allKeys[item] == true {
			var resp string = "Duplicate kvm mapping: " + item
			return false, resp
		}
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
		}
	}
	allKeys = make(map[string]bool)
	for _, item := range check_ip {
		if allKeys[item] == true {
			var resp string = "Duplicate dbghost mapping" + item
			return false, resp
		}
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
		}
	}
	allKeys = make(map[string]bool)
	for _, item := range check_machine {
		if allKeys[item] == true {
			var resp string = "Duplicate dut mapping" + item
			return false, resp
		}
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
		}
	}

	f, err := os.Open("./upload/kvm.csv")
	if err != nil {
		logger.Error(err.Error())
	}

	csvReader := csv.NewReader(f)
	kvm_data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	var hostnames []string
	for i, line := range kvm_data {
		if i > 0 { // omit header line
			var tmp0 string = line[0]
			hostnames = append(hostnames, tmp0)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}

	f, err = os.Open("./upload/dbg.csv")
	if err != nil {
		logger.Error(err.Error())
	}

	csvReader = csv.NewReader(f)
	dbg_data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	var ips []string
	for i, line := range dbg_data {
		if i > 0 { // omit header line
			var tmp0 string = line[1]
			ips = append(ips, tmp0)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}

	f, err = os.Open("./upload/dut.csv")
	if err != nil {
		logger.Error(err.Error())
	}
	csvReader = csv.NewReader(f)
	dut_data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	var machines []string
	for i, line := range dut_data {
		if i > 0 { // omit header line
			var tmp0 string = line[0]
			machines = append(machines, tmp0)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}

	for i, line := range data {
		if i > 0 {
			if contains := slices.Contains(hostnames, line[0]); !contains {
				var resp string = line[0] + " not found in kvm csv"
				return false, resp
			}
			if contains := slices.Contains(ips, line[1]); !contains {
				var resp string = line[1] + " not found in dbg csv"
				return false, resp
			}
			if contains := slices.Contains(machines, line[2]); !contains {
				var resp string = line[2] + " not found in dut csv"
				return false, resp
			}
		}
	}
	return true, ""
}

func Kvm_csv2db() {
	f, err := os.Open("./upload/kvm.csv")
	if err != nil {
		logger.Error(err.Error())
	}
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	_, err = method.Exec("DELETE FROM kvm")
	if err != nil {
		logger.Error(err.Error())
	}
	for i, line := range data {
		if i > 0 { // omit header line
			var tmp0 string = line[0]
			var tmp1 string = line[1]
			var tmp2 string = line[2]
			var tmp3 string = line[3]
			var tmp4 string = line[4]
			_, err = method.Exec("INSERT INTO kvm ( hostname, ip, owner, status, version, NAS_ip, stream_url, stream_status, stream_interface, start_record_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", tmp0, tmp1, tmp2, "null", tmp3, tmp4, "http://"+tmp1+":8081", "recording", "null", 0)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}

func Dbghost_csv2db() {
	f, err := os.Open("./upload/dbg.csv")
	if err != nil {
		logger.Error(err.Error())
	}
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	_, err = method.Exec("DELETE FROM debug_host")
	if err != nil {
		logger.Error(err.Error())
	}
	for i, line := range data {
		if i > 0 { // omit header line
			var tmp0 string = line[0]
			var tmp1 string = line[1]
			var tmp2 string = line[2]
			_, err = method.Exec("INSERT INTO debug_host ( ip, hostname, owner) VALUES (?, ?, ?);", tmp1, tmp0, tmp2)
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}

func Dut_csv2db() {
	f, err := os.Open("./upload/dut.csv")
	if err != nil {
		logger.Error(err.Error())
	}
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	_, err = method.Exec("DELETE FROM machine")
	if err != nil {
		logger.Error(err.Error())
	}
	for i, line := range data {
		if i > 0 { // omit header line
			var tmp0 string = line[0]
			var tmp1 string = line[1]
			var tmp2 string = line[2]
			_, err = method.Exec("INSERT INTO machine ( machine_name, ssim, status, cycle_cnt, error_timestamp, path, threshold) VALUES (?, ?, ?, ?, ?, ?, ?);", tmp0, tmp1, 1, 0, 0, "null", tmp2)
			if err != nil {
				logger.Error(err.Error())
			}

		}
	}
}

func Dbgunit_csv2db() {
	f, err := os.Open("./upload/map.csv")
	if err != nil {
		logger.Error(err.Error())
	}
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}
	for i, line := range data {
		if i > 0 { // omit header line
			uuid := uuid.New().String()
			var tmp0 string = line[0]
			var tmp1 string = line[1]
			var tmp2 string = line[2]
			var tmp3 string = line[3]
			_, err = method.Exec("INSERT INTO debug_unit (uuid, hostname, ip,  machine_name, project) VALUES (?, ?, ?, ?, ?);", uuid, tmp0, tmp1, tmp2, tmp3)
			if err != nil {
				logger.Error(err.Error())
			}

		}
	}
}
func createMappingList(data [][]string) []Mapping_file {
	var mappingList []Mapping_file
	for i, line := range data {
		if i > 0 { // omit header line
			var rec Mapping_file
			for j, field := range line {
				if j == 0 {
					rec.Hostname = field
				} else if j == 1 {
					rec.Ip = field
				} else if j == 2 {
					rec.Machine_name = field
				} else if j == 3 {
					rec.Project = field
				}
			}
			mappingList = append(mappingList, rec)
		}
	}
	return mappingList
}

func Kvm_csv_mapping(c *gin.Context) {
	file_list := [4]string{"dutfile", "kvmfile", "dbgfile", "mapfile"}
	for _, fileName := range file_list {
		file, header, err := c.Request.FormFile(fileName)
		filename := header.Filename
		out, err := os.Create("./upload/" + filename)
		if err != nil {
			logger.Error(err.Error())
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			logger.Error(err.Error())
		}
	}
	f, err := os.Open("./upload/map.csv")
	if err != nil {
		logger.Error(err.Error())
	}
	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		logger.Error(err.Error())
	}

	result, result_mesg := checkMappingFile(data)
	fmt.Printf(result_mesg)
	if !result {
		apiservice.ResponseWithJson(c.Writer, http.StatusBadRequest, result_mesg)
		return
	}
	_, err = method.Exec("DELETE FROM debug_unit")

	//import csv file to db
	Kvm_csv2db()
	Dbghost_csv2db()
	Dut_csv2db()
	Dbgunit_csv2db()

	// mappingList := createMappingList(data)

	defer f.Close()
	//_, err = method.Exec("INSERT INTO debug_unit ( uuid, hostname, ip, machine_name, project) VALUES (?, ?, ?, ?, ?);", uuid, Req.Hostname, Req.Ip, Req.Machine_name, Req.Project)

	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func Kvm_status(c *gin.Context) {
	action := c.Query("action")
	var Resp Kvm_state
	var Req Kvm_state
	body, _ := ioutil.ReadAll(c.Request.Body)
	_ = json.Unmarshal(body, &Req)
	if action == "search" {
		Resp.Hostname = Req.Hostname
		row := method.QueryRow("SELECT stream_status FROM kvm WHERE hostname=?", Req.Hostname)
		err := row.Scan(&Resp.Stream_status)
		if err != nil {
			logger.Error("search kvm status error" + err.Error())
		}
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
		return
	} else if action == "update" {
		_, err := method.Exec("UPDATE kvm SET stream_status=? WHERE hostname=?", Req.Stream_status, Req.Hostname)
		if err != nil {
			logger.Error("update kvm status error" + err.Error())
		}
		var ip string
		row := method.QueryRow("SELECT ip FROM kvm WHERE hostname=?", Req.Hostname)
		err = row.Scan(&ip)
		if err != nil {
			logger.Error(err.Error())
		}
		if Req.Stream_status == "recording" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			_, err := client.Get("https://" + ip + ":8443/api/switch_mode?mode=motion")
			if err != nil {
				logger.Error("update send switch mode request to kvm error" + err.Error())
			}
		} else if Req.Stream_status == "idle" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			_, err := client.Get("https://" + ip + ":8443/api/switch_mode?mode=kvmd")
			if err != nil {
				logger.Error("update send switch mode request to kvm error" + err.Error())
			}
		}

	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "update successfully")

}
func Kvm_genvideo(c *gin.Context) {
	body, _ := ioutil.ReadAll(c.Request.Body)
	var Req Video_info
	_ = json.Unmarshal(body, &Req)
	fmt.Println(Req.Duration)
	files, err := ioutil.ReadDir("/home/media/video/" + Req.Hostname + "/")
	if err != nil {
		logger.Error("List video dir fail: " + err.Error())
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	sort.Strings(filenames)
	f, err := os.OpenFile("/home/media/video/"+Req.Hostname+"/self-define.m3u8", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 644)
	if err != nil {
		logger.Error("open video file fail: " + err.Error())
		return
	}
	defer f.Close()
	f.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-ALLOW-CACHE:YES\n#EXT-X-TARGETDURATION:10\n")
	var h, m string
	if Req.Hour < 10 {
		h = strconv.Itoa(Req.Hour)
		h = "0" + h
	} else {
		h = strconv.Itoa(Req.Hour)
	}
	if Req.Minute < 10 {
		m = strconv.Itoa(Req.Minute)
		m = "0" + m
	} else {
		m = strconv.Itoa(Req.Minute)
	}
	for ii, filename := range filenames {
		if strings.Contains(filename, "2023-08-16_"+h+"-"+m) {
			for i := 0; i < Req.Duration/10; i++ {
				f.WriteString("#EXTINF:10.000000,\n")
				f.WriteString(filenames[ii+i] + "\n")
			}
			break
		}
	}
	f.WriteString("#EXT-X-ENDLIST")
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
