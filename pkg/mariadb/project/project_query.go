package project_query

import (
	emailfunction "recorder/internal/email_function"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	"strings"
)

func Update_project_setting(setting structure.Project_setting_Tamplate, Email_string string) {
	_, err := method.Exec("UPDATE project SET short_name = ?, owner=?, email_list=? WHERE project_name = ?", setting.Short_name, setting.Owner, Email_string, setting.Project_name)
	if err != nil {
		logger.Error("Update project status error: " + err.Error())
	}
}
func Get_all_projects_setting() []structure.Project_setting_Tamplate {
	Project, err := method.Query("SELECT project_name,short_name,owner,email_list FROM project")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
	}
	var project_list []structure.Project_setting_Tamplate
	for Project.Next() {
		var project_setting structure.Project_setting_Tamplate
		var email_string string
		err := Project.Scan(&project_setting.Project_name, &project_setting.Short_name, &project_setting.Owner, &email_string)
		if err != nil {
			logger.Error(err.Error())
			return project_list
		}
		project_setting.Email_list = emailfunction.String_to_Email(email_string)
		project_list = append(project_list, project_setting)
	}
	return project_list
}
func Get_all_Floor() map[string]int {
	Floors := make(map[string]int)
	Projects, err := method.Query("SELECT project_name FROM project")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
		return Floors
	}
	for Projects.Next() {
		var Floor string
		err = Projects.Scan(&Floor)
		if err != nil {
			logger.Error(err.Error())
			return Floors
		}
		String_list := strings.Split(Floor, "_")
		_, ok := Floors[String_list[0]]
		if ok {
			Floors[String_list[0]]++
		} else {
			Floors[String_list[0]] = 1
		}
	}
	return Floors
}
