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

type Pr
func Project_list(c *gin.Context) {
	var Project_list Projectj_list_response
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