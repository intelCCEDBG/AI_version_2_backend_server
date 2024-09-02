package kvm_api

import (
	"bytes"
	"fmt"
	"sync"

	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"strconv"
	"time"

	"golang.org/x/exp/slices"

	"recorder/config"
	videogen "recorder/internal/video_gen"
	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/mariadb/method"
	project_query "recorder/pkg/mariadb/project"
	unit_query "recorder/pkg/mariadb/unit"
	"recorder/pkg/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type Proj_exec struct {
	Project   string `json:"project"`
	Operation string `json:"operation`
}

type KVM_floor struct {
	Floor  []string `json:"floor"`
	Amount []int    `json:"amount"`
}

type Hostnames struct {
	Hostnames    []string `json:"hostnames"`
	Islinked     []bool   `json:"islinked"`
	Messagecount []int    `json:"messagecount"`
}
type Message struct {
	Hostname string `json:"hostname"`
	Message  string `json:"message"`
}
type Messages struct {
	Hostname string   `json:"hostname"`
	Messages []string `json:"message"`
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
	body, err := io.ReadAll(c.Request.Body)
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
	_, err = method.Exec("INSERT INTO debug_unit ( uuid, hostname, ip, machine_name, project) VALUES (?, ?, ?, ?, ?);", uuid, Req.Hostname, Req.Ip, Req.Machine_name, Req.Project)
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
		if allKeys[item] {
			var resp string = "Duplicate kvm mapping: " + item
			return false, resp
		}
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
		}
	}
	allKeys = make(map[string]bool)
	for _, item := range check_ip {
		if allKeys[item] {
			var resp string = "Duplicate dbghost mapping" + item
			return false, resp
		}
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
		}
	}
	allKeys = make(map[string]bool)
	for _, item := range check_machine {
		if allKeys[item] {
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
	_, err = method.Exec("DELETE FROM kvm_message")
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
			_, err = method.Exec("INSERT INTO machine ( machine_name, ssim, status, cycle_cnt,cycle_cnt_high,  error_timestamp, path, threshold, lock_coord) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);", tmp0, tmp1, 1, 0, 0, 0, "null", tmp2, "")
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
	Project_list := make(map[string]bool)
	rows, err := method.Query("SELECT project_name FROM project;")
	if err != nil {
		logger.Error("Query project list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		Project_list[tmp] = true
	}
	rows, err = method.Query("SELECT DISTINCT project FROM debug_unit;")
	if err != nil {
		logger.Error("Query project list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if _, exists := Project_list[tmp]; !exists {
			create_new_project(tmp)
		}
	}
}
func create_new_project(project_name string) {
	_, err := method.Exec("INSERT INTO project (project_name, short_name, owner,  email_list, status, freeze_detection) VALUES (?, ?, ?, ?, ?,?);", project_name, project_name, "", "", 0, "open")
	if err != nil {
		logger.Error("INSERT project list error: " + err.Error())
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
		if err != nil {
			logger.Error(err.Error())
		}
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
	fmt.Println(result_mesg)
	if !result {
		apiservice.ResponseWithJson(c.Writer, http.StatusBadRequest, result_mesg)
		return
	}
	_, err = method.Exec("DELETE FROM debug_unit")
	if err != nil {
		logger.Error(err.Error())
	}

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
	body, _ := io.ReadAll(c.Request.Body)
	_ = json.Unmarshal(body, &Req)
	if action == "search" {
		Resp.Hostname = Req.Hostname
		row := method.QueryRow("SELECT stream_status FROM kvm WHERE hostname=?", Req.Hostname)
		err := row.Scan(&Resp.Stream_status)
		if err != nil {
			logger.Error("search kvm status error, " + err.Error())
		}
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
		return
	} else if action == "update" {
		var ip string
		row := method.QueryRow("SELECT ip FROM kvm WHERE hostname=?", Req.Hostname)
		err := row.Scan(&ip)
		if err != nil {
			logger.Error(err.Error())
		}
		if Req.Stream_status == "recording" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr, Timeout: 7 * time.Second}
			_, err := client.Get("https://" + ip + ":8443/api/switch_mode?mode=motion")
			if err != nil {
				logger.Error(err.Error())
			}
			fmt.Println("Enter Redis")
			redis.Redis_set("kvm:"+Req.Hostname+":recording", Req.Hostname)
			fmt.Println("finish writing")
			for {
				row = method.QueryRow("SELECT stream_status FROM kvm WHERE hostname=?", Req.Hostname)
				var stream_status string
				_ = row.Scan(&stream_status)
				//todo: handle fail case
				if stream_status == "recording" {
					break
				} else {
					time.Sleep(1 * time.Second)
				}
			}
			if err != nil {
				logger.Error("update send switch mode request to kvm error" + err.Error())
			}
		} else if Req.Stream_status == "idle" {
			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			client := &http.Client{Transport: tr}
			_, err := client.Get("https://" + ip + ":8443/api/switch_mode?mode=kvmd")
			redis.Redis_set("kvm:"+Req.Hostname+":stop", Req.Hostname)
			fmt.Println("finish writing", Req.Hostname)
			for {
				row = method.QueryRow("SELECT stream_status FROM kvm WHERE hostname=?", Req.Hostname)
				var stream_status string
				_ = row.Scan(&stream_status)
				if stream_status == "idle" {
					break
				} else {
					time.Sleep(1 * time.Second)
				}
			}
			if err != nil {
				logger.Error("update send switch mode request to kvm error" + err.Error())
			}
		} else if Req.Stream_status == "error" {
			redis.Redis_set("kvm:"+Req.Hostname+":error", Req.Hostname)
			for {
				row = method.QueryRow("SELECT stream_status FROM kvm WHERE hostname=?", Req.Hostname)
				var stream_status string
				_ = row.Scan(&stream_status)
				if stream_status == "error" {
					break
				} else {
					time.Sleep(1 * time.Second)
				}
			}
			if err != nil {
				logger.Error("update send switch mode request to kvm error" + err.Error())
			}
		}

	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "update successfully")
}
func Kvm_genvideo(c *gin.Context) {
	var Req Video_info
	Req.Hostname = c.Query("kvm_hostname")
	Req.Hour, _ = strconv.Atoi(c.Query("hour"))
	Req.Duration, _ = strconv.Atoi(c.Query("duration"))
	Req.Minute, _ = strconv.Atoi(c.Query("minute"))
	// fmt.Println(Req.Duration)
	// fmt.Println(Req.Hostname)
	videogen.GenerateVideo(Req.Hour, Req.Minute, Req.Duration, Req.Hostname, "self-define")
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Kvm_genminutevideo(c *gin.Context) {
	var Req Video_info
	Req.Hostname = c.Query("kvm_hostname")
	Req.Hour, _ = strconv.Atoi(c.Query("hour"))
	Req.Duration, _ = strconv.Atoi(c.Query("duration"))
	Req.Minute, _ = strconv.Atoi(c.Query("minute"))
	fmt.Println(Req.Hour)
	fmt.Println(Req.Minute)
	videogen.GenerateVideo(Req.Hour, Req.Minute, Req.Duration, Req.Hostname, c.Query("duration")+"M")
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func Project_status(c *gin.Context) { //start entry point
	body, _ := io.ReadAll(c.Request.Body)
	var Req Proj_exec
	_ = json.Unmarshal(body, &Req)
	if Req.Operation == "start" {
		rows, err := method.Query("SELECT hostname FROM debug_unit WHERE project=?", Req.Project)
		if err != nil {
			logger.Error(err.Error())
		}
		var wg sync.WaitGroup
		for rows.Next() {
			wg.Add(1)
			var req Kvm_state
			err = rows.Scan(&req.Hostname)
			if err != nil {
				logger.Error(err.Error())
			}
			req.Stream_status = "recording"
			json_data, err := json.Marshal(req)
			// logger.Info("Enter...")
			unit := unit_query.Get_unitbyhostname(req.Hostname)
			dut_query.Clean_Cycle_Count(unit.Machine_name)
			if err != nil {
				logger.Error(err.Error())
			}
			Port := config.Viper.GetString("SERVER_PORT")
			// http.DefaultClient.Timeout = time.Second * 20
			go func(Port string, json_data []byte, Hostname string) {
				defer wg.Done()
				http.Post("http://127.0.0.1:"+Port+"/api/kvm/stream_status?action=update", "application/json", bytes.NewBuffer(json_data))
				// fmt.Println(Hostname)
			}(Port, json_data, req.Hostname)
		}
		project_query.Add_start_time(Req.Project)
		wg.Wait()
	} else if Req.Operation == "stop" {
		rows, err := method.Query("SELECT hostname FROM debug_unit WHERE project=?", Req.Project)
		if err != nil {
			logger.Error(err.Error())
		}
		var wg sync.WaitGroup
		for rows.Next() {
			wg.Add(1)
			var req Kvm_state
			err = rows.Scan(&req.Hostname)
			if err != nil {
				logger.Error(err.Error())
			}
			req.Stream_status = "idle"
			json_data, err := json.Marshal(req)
			if err != nil {
				logger.Error(err.Error())
			}
			Port := config.Viper.GetString("SERVER_PORT")
			http.DefaultClient.Timeout = time.Second * 20
			go func(Port string, json_data []byte) {
				defer wg.Done()
				http.Post("http://127.0.0.1:"+Port+"/api/kvm/stream_status?action=update", "application/json", bytes.NewBuffer(json_data))
			}(Port, json_data)
		}
		wg.Wait()
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Kvm_modify(c *gin.Context) {
	hostname := c.Query("hostname")
	ip := c.Query("ip")
	nas_ip := c.Query("nas_ip")
	owner := c.Query("owner")
	_, err := method.Exec("UPDATE kvm SET ip=?, owner=?, nas_ip=?, stream_url=? WHERE hostname=?", ip, owner, nas_ip, "http://"+ip+":8081", hostname)
	if err != nil {
		logger.Error("Modify kvm info error" + err.Error())
		apiservice.ResponseWithJson(c.Writer, http.StatusNotFound, "")
		return
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Get_KVM_Floor(c *gin.Context) {
	var Floor_out KVM_floor
	Floor_map := kvm_query.Get_all_Floor_from_hostname()
	for floor := range Floor_map {
		Floor_out.Floor = append(Floor_out.Floor, floor)
		Floor_out.Amount = append(Floor_out.Amount, Floor_map[floor])
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Floor_out)
}
func Get_hostnames_by_floor(c *gin.Context) {
	var Hostnames_out Hostnames
	floor := c.Query("floor")
	Hostnames_out.Hostnames = kvm_query.Get_hostnames_by_floor(floor)
	Hostnames_out.Islinked = kvm_query.Get_link_status(Hostnames_out.Hostnames)
	Hostnames_out.Messagecount = kvm_query.Get_messagecount(Hostnames_out.Hostnames)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Hostnames_out)
}
func Insert_message(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var Req Message
	_ = json.Unmarshal(body, &Req)
	kvm_query.Insert_message(Req.Hostname, Req.Message)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Get_kvm_message(c *gin.Context) {
	var Messages_out Messages
	hostname := c.Query("hostname")
	Messages_out.Hostname = hostname
	Messages_out.Messages = kvm_query.Get_kvm_message(hostname)
	if Messages_out.Messages == nil {
		empty := []string{}
		Messages_out.Messages = empty
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Messages_out)
}
func Delete_message(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var Req Message
	_ = json.Unmarshal(body, &Req)
	kvm_query.Delete_message(Req.Hostname, Req.Message)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
