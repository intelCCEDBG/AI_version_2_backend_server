package email

import (
	"encoding/json"
	"io"
	"net/http"
	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"

	"strconv"
	"time"

	// "os"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
}
type Email_command_in_s struct {
	Host    int    `json:"host"`
	User    string `json:"user"`
	Context string `json:"context"`
}
type Report_out struct {
	Status []Machine_status `json:"state"`
}
type Email_command_out_s struct {
	User    string `json:"user"`
	Context string `json:"context"`
	Time    string `json""time"`
}
type Email_command_out struct {
	List []Email_command_out_s `json:"list"`
}
type Machine_status struct {
	Machine_name string `json:"machine_name"`
	Status       int    `json:"status"`
	Type         int    `json:"type"`
}
type Porject_threestrike struct {
	Project string `json:"project"`
	Status  int    `json:"status"`
}

// stop using
func Email_command_in(c *gin.Context) {
	var In_pakage Email_command_in_s
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1024))
	logger.Error("Read email command body error: " + err.Error())
	// error.Error_log(err)
	_ = json.Unmarshal(body, &In_pakage)
	defer c.Request.Body.Close()
	currant_time := time.Now().Unix()
	_, err = method.Exec("INSERT INTO email_command (host_id, time,user, context) VALUES (?,?,?,?);", In_pakage.Host, currant_time, In_pakage.User, In_pakage.Context)
	logger.Error("Insert into email_command error: " + err.Error())
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

// stop using
func Email_command_get(c *gin.Context) {
	Host := c.Query("host")

	host, _ := strconv.Atoi(Host)
	var Response Email_command_out
	rows, err := method.Query("SELECT user, context, time FROM email_command WHERE host_id=? ORDER BY time;", host)
	defer rows.Close()
	logger.Error("Search in email_command error: " + err.Error())
	for rows.Next() {
		var Command Email_command_out_s
		var Time int64
		err = rows.Scan(&Command.User, &Command.Context, &Time)
		logger.Error("Scan error: " + err.Error())
		if Time != 0 {
			Timestamp := time.Unix(Time, 0)
			Command.Time = Timestamp.String()
		}
		Response.List = append(Response.List, Command)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Response)
}
func Report(c *gin.Context) {
	Host := c.Query("project")
	host, _ := strconv.Atoi(Host)
	Time := c.Query("time")
	time, _ := strconv.Atoi(Time)

	var Report Report_out
	rows, err := method.Query("SELECT machine_name, status, type FROM status_record WHERE host_id=? AND time=?;", host, time)
	defer rows.Close()
	logger.Error("Search in status_record error: " + err.Error())
	for rows.Next() {
		var Machine Machine_status
		var Type string
		err = rows.Scan(&Machine.Machine_name, &Machine.Status, &Type)
		logger.Error("Scan error: " + err.Error())
		Machine.Type, _ = strconv.Atoi(Type)
		Report.Status = append(Report.Status, Machine)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Report)
}
func Enable_mail_constraint(c *gin.Context) {
	Project := c.Query("project")
	Enable := c.Query("enable")
	enable, _ := strconv.Atoi(Enable)
	if enable == -1 {
		var ret Porject_threestrike
		ret.Project = Project
		// var ret int
		row := method.QueryRow("SELECT three_mail_constraint FROM project WHERE project_name=?;", Project)
		row.Scan(&ret.Status)
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, ret)
		return
	}
	_, err := method.Exec("UPDATE project SET three_mail_constraint=? WHERE project_name=?;", enable, Project)
	if err != nil {
		logger.Error("Update to host_info error: " + err.Error())
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
