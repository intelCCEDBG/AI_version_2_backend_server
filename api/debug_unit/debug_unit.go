package dbgunit_api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"recorder/pkg/mariadb/method"
	"recorder/pkg/logger"
	"recorder/pkg/apiservice"

)

type Project_list_response struct {
	Project		[]string	`json:"projects"`
}

type Debug_unit_info struct{
	Hostname		string	`json:"hostname"`
	Machine_name	string	`json:"machine_name"`
	Ip				string 	`json:"ip"`
}

type Project_info_response struct{
	Project		string				`json:"project"`
	Info		[]Debug_unit_info	`json:"info"`
}
// {
// 	[{"project":"ODINW", "info":[{"hstname":"10.x.x.","machine_name":"ODINWXXX"},{"10.x.y", "ODINWXXY"}]}]
// }
func Project_list(c *gin.Context) {
	var Project_list Project_list_response
	rows, err := method.Query("SELECT DISTINCT project FROM debug_unit;")
	if err != nil {
		logger.Error("Query project list error: " + err.Error())
	}
	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		Project_list.Project = append(Project_list.Project,tmp)
	}	

	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Project_list)
}

func Project_info(c *gin.Context) {
	var Resp Project_info_response
	project_name := c.Query("project")
	if project_name == "ALL" {
		var Resp2 []Project_info_response
		var Project_list []string
		rows, err := method.Query("SELECT DISTINCT project FROM debug_unit;")
		if err != nil {
			logger.Error("Search all project info error: " + err.Error())
		}
		for rows.Next() {
			var tmp string
			err = rows.Scan(&tmp)
			Project_list = append(Project_list,tmp)
		}
		for _, pj := range Project_list {
			rows, err := method.Query("SELECT hostname, machine_name, ip FROM debug_unit WHERE project=?;",pj)
			if err != nil {
				logger.Error("Search project info error: " + err.Error())
			}
			var tmp2 []Debug_unit_info
			for rows.Next() {
				var tmp Debug_unit_info
				err = rows.Scan(&tmp.Hostname, &tmp.Machine_name, &tmp.Ip)
				if err != nil {
					logger.Error("Search project info error: " + err.Error())
				}
				tmp2 = append(tmp2, tmp)
			}
			Resp.Project = pj
			Resp.Info = tmp2
			Resp2 = append(Resp2,Resp)
		}	
		apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp2)
		return
	}else{
		rows, err := method.Query("SELECT hostname, machine_name, ip FROM debug_unit WHERE project=?;",project_name)
		if err != nil {
			logger.Error("Search project info error: " + err.Error())
		}
		var tmp2 []Debug_unit_info
		for rows.Next() {
			var tmp Debug_unit_info
			err = rows.Scan(&tmp.Hostname, &tmp.Machine_name, &tmp.Ip)
			if err != nil {
				logger.Error("Search project info error: " + err.Error())
			}
			tmp2 = append(tmp2, tmp)
		}
		Resp.Project = project_name
		Resp.Info = tmp2
	}
	apiservice.ResponseWithJson(c.Writer, http.StatusOK, Resp)
}