package project

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	emailfunction "recorder/internal/email_function"
	"recorder/internal/structure"
	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	project_query "recorder/pkg/mariadb/project"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
}
type In_Project_tamplate struct {
	Project_name string   `json:"project_name"`
	Short_name   string   `json:"short_name"`
	Owner        string   `json:"owner"`
	Report_list  []string `json:"email_list"`
	Alert_list   []string `json:"alert_list"`
}
type freeze_template struct {
	Project_name string `json:"project_name"`
	Switch       string `json:"switch"`
}
type Project_exec_template struct {
	Project_name string `json:"project_name"`
	Operation    string `json:"operation"`
}
type Project_status_respo struct {
	Project_name string `json:"project_name"`
	Status       int    `json:"status"`
}
type Project_floor struct {
	Floor  []string `json:"floor"`
	Amount []int    `json:"amount"`
}
type Report_State struct {
	Fail  []int `json:"fail"`
	Total int   `json:"total"`
}

func GetProjectSetting(c *gin.Context) {
	var Project_setting_out structure.ProjectTemplate
	project_name := c.Query("project_name")
	rows, err := method.Query("SELECT project_name,short_name,owner,email_list FROM project WHERE project_name=?;", project_name)
	if err != nil {
		logger.Error("Query project setting error: " + err.Error())
		Project_setting_out.ProjectName = "Not Found"
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, Project_setting_out)
		return
	}
	for rows.Next() {
		var Email_string string
		err = rows.Scan(&Project_setting_out.ProjectName, &Project_setting_out.ShortName, &Project_setting_out.Owner, &Email_string)
		if err != nil {
			logger.Error("Reading project setting error: " + err.Error())
		}
		Project_setting_out.EmailList = emailfunction.String_to_Email(Email_string)
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Project_setting_out)
}
func GetProjectSettingByCode(c *gin.Context) {
	var Out_Email []structure.EmailTemplate
	var Out_email_template structure.EmailTemplate
	var Project_setting_out structure.ProjectTemplate
	code := c.Query("short_name")
	rows, err := method.Query("SELECT project_name,short_name,owner,email_list FROM project WHERE short_name=?;", code)
	if err != nil {
		logger.Error("Query project setting error: " + err.Error())
		Project_setting_out.ProjectName = "Not Found"
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, Project_setting_out)
		return
	}
	for rows.Next() {
		var Email_string string
		err = rows.Scan(&Project_setting_out.ProjectName, &Project_setting_out.ShortName, &Project_setting_out.Owner, &Email_string)
		if err != nil {
			logger.Error("Reading project setting error: " + err.Error())
		}
		email_list := strings.Split(Email_string, " ")
		for _, S := range email_list {
			t := strings.Split(S, ",")
			if len(t) == 3 {
				Out_email_template.Account = t[0]
				Out_email_template.Report, _ = strconv.ParseBool(t[1])
				Out_email_template.Alert, _ = strconv.ParseBool(t[2])
				Out_Email = append(Out_Email, Out_email_template)
			}
		}
		Project_setting_out.EmailList = Out_Email
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Project_setting_out)
}
func SetProjectSetting(c *gin.Context) {
	var Project_setting_in In_Project_tamplate
	var Update_project_tamplate structure.ProjectTemplate
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Read project setting request error: " + err.Error())
	}
	err = json.Unmarshal(body, &Project_setting_in)
	if err != nil {
		logger.Error("Parse project setting request error: " + err.Error())
	}
	Email_string := emailfunction.Email_to_string(Email_list_merge(Project_setting_in))
	Update_project_tamplate.ProjectName = Project_setting_in.Project_name
	Update_project_tamplate.Owner = Project_setting_in.Owner
	Update_project_tamplate.ShortName = Project_setting_in.Short_name
	project_query.Update_project_setting(Update_project_tamplate, Email_string)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func Email_list_merge(In_pakage In_Project_tamplate) []structure.EmailTemplate {
	var List []structure.EmailTemplate
	var user structure.EmailTemplate
	Uni := Union(In_pakage.Report_list, In_pakage.Alert_list)
	for _, add := range Uni {
		user.Account = add
		user.Report = false
		user.Alert = false
		for _, rep := range In_pakage.Report_list {
			if add == rep {
				user.Report = true
			}
		}
		for _, rep := range In_pakage.Alert_list {
			if add == rep {
				user.Alert = true
			}
		}
		List = append(List, user)
	}
	return List
}
func Union(a, b []string) []string {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}
func FreezeDetection(c *gin.Context) {
	action := c.Query("action")
	project_name := c.Query("project_name")
	switches := c.Query("switch")
	var Resp freeze_template
	if action == "search" {
		Resp.Project_name = project_name
		row := method.QueryRow("SELECT freeze_detection FROM project WHERE project_name=?", project_name)
		err := row.Scan(&Resp.Switch)
		if err != nil {
			logger.Error("Reading freeze status error: " + err.Error())
		}
	} else if action == "update" {
		_, err := method.Exec("UPDATE project SET freeze_detection=? WHERE project_name=?", switches, project_name)
		if err != nil {
			logger.Error("Update feeze status error: " + err.Error())
		}
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
		return
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
}
func Control(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body) //io.LimitReader限制大小
	if err != nil {
		logger.Error("Reading freeze status error: " + err.Error())
	}
	var In_pakage Project_exec_template
	err = json.Unmarshal(body, &In_pakage)
	if err != nil {
		logger.Error("Reading freeze status error: " + err.Error())
	}
	if In_pakage.Operation == "stop" {
		log.Println("Stop Project:", In_pakage.Project_name)
		_, err = method.Exec("UPDATE project SET status=? WHERE project_name=?;", 0, In_pakage.Project_name)
		if err != nil {
			logger.Error("Reading freeze status error: " + err.Error())
		}
		//todo
	} else if In_pakage.Operation == "start" {
		log.Println("Lock Project:", In_pakage.Project_name)
		_, err = method.Exec("UPDATE project SET status=? WHERE project_name=?;", 2, In_pakage.Project_name)
		if err != nil {
			logger.Error("Reading freeze status error: " + err.Error())
		}
		//todo
	} else if In_pakage.Operation == "wake" {
		log.Println("Wake Project:", In_pakage.Project_name)
		_, err = method.Exec("UPDATE project SET status=? WHERE project_name=?;", 1, In_pakage.Project_name)
		if err != nil {
			logger.Error("Reading freeze status error: " + err.Error())
		}
		//todo
	} else if In_pakage.Operation == "reset" {
		log.Println("Reset Project:", In_pakage.Project_name)
		_, err = method.Exec("UPDATE project SET status=? WHERE project_name=?;", 3, In_pakage.Project_name)
		if err != nil {
			logger.Error("Reading freeze status error: " + err.Error())
		}
		//todo
	} else if In_pakage.Operation == "Search" {
		// log.Println("Search For Project Status:", In_pakage.Project_name)
		var statue_re int
		row := method.QueryRow("SELECT status FROM project WHERE project_name=?;", In_pakage.Project_name)
		err = row.Scan(&statue_re)
		if err != nil {
			logger.Error("Reading freeze status error: " + err.Error())
			response := ApiResponse{"500", "project cannot found in method"}
			apiservice.ResponseWithJson(c.Writer, http.StatusOK, response)
			return
		}
		var response Project_status_respo
		response.Project_name = In_pakage.Project_name
		response.Status = statue_re
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, response)
		return
	}
	response := ApiResponse{"200", ""}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, response)
}
func GetProjectFloor(c *gin.Context) {
	var Floor_out Project_floor
	Floor_map := project_query.Get_all_Floor()
	for floor := range Floor_map {
		Floor_out.Floor = append(Floor_out.Floor, floor)
		Floor_out.Amount = append(Floor_out.Amount, Floor_map[floor])
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Floor_out)
}
func GetReportState(c *gin.Context) {
	project := c.Query("project")
	Resp := project_query.Get_Units(project)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
}
func AddNewProject(c *gin.Context) {
	project := c.Query("project")
	codename := c.Query("codename")
	create_new_project(project, codename)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}
func create_new_project(project_name string, codename string) {
	_, err := method.Exec("INSERT INTO project (project_name, short_name, owner,  email_list, status, freeze_detection) VALUES (?, ?, ?, ?, ?, ?);", project_name, codename, "", "", 0, "open")
	if err != nil {
		logger.Error("INSERT project error: " + err.Error())
	}
}
func GetSsimAndThreshold(c *gin.Context) {
	project := c.Query("project")
	Resp := project_query.Get_ssim_and_threshold(project)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
}
func UpdateSsimAndThreshold(c *gin.Context) {
	var Resp structure.SsimAndThreshold
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Read ssim and threshold request error: " + err.Error())
	}
	err = json.Unmarshal(body, &Resp)
	if err != nil {
		logger.Error("Parse ssim and threshold request error: " + err.Error())
	}
	project_query.Update_ssim_and_threshold(Resp)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func DeleteProject(c *gin.Context) {
	project := c.Query("project")
	project_query.Delete_project(project)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, "")
}

func GetStartTime(c *gin.Context) {
	project := c.Query("project")
	Resp := project_query.GetStartTime(project)
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
}

func CheckProject(c *gin.Context) {
	checkType := c.Query("type")
	switch checkType {
	case "name":
		projectName := c.Query("project")
		exists, err := project_query.ProjectNameExists(projectName)
		if err != nil {
			logger.Error("Check project name error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"exists": exists})
	case "code":
		projectCode := c.Query("code")
		exists, err := project_query.ProjectCodeExists(projectCode)
		if err != nil {
			logger.Error("Check project code error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"exists": exists})
	}
}
