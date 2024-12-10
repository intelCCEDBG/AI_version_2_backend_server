package dbgunit_api

import (
	"net/http"
	"recorder/internal/structure"
	"recorder/pkg/apiservice"
	"recorder/pkg/logger"
	debughost_query "recorder/pkg/mariadb/debughost"
	dut_query "recorder/pkg/mariadb/dut"
	"recorder/pkg/mariadb/method"
	project_query "recorder/pkg/mariadb/project"
	unit_query "recorder/pkg/mariadb/unit"

	"github.com/gin-gonic/gin"
	// "fmt"
)

type addUnitRequest struct {
	ProjectName string `json:"project"`
	MachineName string `json:"dut"`
	HostName    string `json:"kvm"`
	DbgIp       string `json:"dbh"`
}

type joinRequest struct {
	ProjectName string   `json:"project"`
	Machines    []string `json:"duts"`
}

type Project_list_response struct {
	Project []string `json:"projects"`
}

type ProjectInfoResponse struct {
	Duts []structure.DebugUnitInfo `json:"duts"`
}

func AddDebugUnit(c *gin.Context) {
	var req addUnitRequest
	err := c.BindJSON(&req)
	if err != nil {
		logger.Error("Bind json error: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// check if the project exists
	if req.ProjectName != "" {
		exists, err := project_query.ProjectNameExists(req.ProjectName)
		if err != nil {
			logger.Error("Check project name error: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if !exists {
			logger.Error("Project name not exists")
			c.JSON(http.StatusBadRequest, gin.H{"error": "project name not exist"})
		}
	}
	// check if the debug host ip exists
	exists, err := debughost_query.CheckDebugHostIPExist(req.DbgIp)
	if err != nil {
		logger.Error("Check debug host ip error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		logger.Error("Debug host ip not exists")
		c.JSON(http.StatusBadRequest, gin.H{"error": "debug host ip not exist"})
		return
	}
	// check if the debug host is free
	free, err := debughost_query.CheckDebugHostFree(req.DbgIp, req.HostName)
	if err != nil {
		logger.Error("Check debug host free error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !free {
		logger.Error("Debug host is not free")
		c.JSON(http.StatusBadRequest, gin.H{"error": "debug host is not free"})
		return
	}
	// check if the machine name exists
	exists, err = dut_query.CheckDutExist(req.MachineName)
	if err != nil {
		logger.Error("Check dut exists error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !exists {
		logger.Error("Dut not exists")
		c.JSON(http.StatusBadRequest, gin.H{"error": "dut not exist"})
		return
	}
	// check if the machine name is free
	free, err = dut_query.CheckDutFree(req.MachineName, req.HostName)
	if err != nil {
		logger.Error("Check dut free error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !free {
		logger.Error("Dut is not free")
		c.JSON(http.StatusBadRequest, gin.H{"error": "dut is not free"})
		return
	}
	// insert the debug unit
	unit := structure.Unit{
		Project:     req.ProjectName,
		MachineName: req.MachineName,
		Hostname:    req.HostName,
		Ip:          req.DbgIp,
	}
	uuid, err := unit_query.CreateUnit(unit)
	if err != nil {
		logger.Error("Create unit error: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uuid": uuid})
}

func ProjectList(c *gin.Context) {
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

func ProjectInfo(c *gin.Context) {
	res := ProjectInfoResponse{}
	projectName := c.Query("project")
	if projectName == "ALL" {
		projectList, err := project_query.GetAllProjects()
		if err != nil {
			logger.Error("Get all project list error: " + err.Error())
			c.JSON(500, gin.H{
				"error": "Get all project list error: " + err.Error(),
			})
		}
		for _, pj := range projectList {
			unitsInfo, err := unit_query.GetUnitsInfoByProject(pj)
			if err != nil {
				logger.Error("Get units info by project error: " + err.Error())
				c.JSON(500, gin.H{
					"error": "Get units info by project error: " + err.Error(),
				})
			}
			res.Duts = append(res.Duts, unitsInfo...)
		}
	} else {
		unitsInfo, err := unit_query.GetUnitsInfoByProject(projectName)
		if err != nil {
			logger.Error("Get units info by project error: " + err.Error())
			c.JSON(500, gin.H{
				"error": "Get units info by project error: " + err.Error(),
			})
		}
		res.Duts = unitsInfo
	}
	c.JSON(200, res)
}

func DownloadCsv(c *gin.Context) {
	url, err := unit_query.ExportAllToCsv()
	if err != nil {
		logger.Error("Export all units to csv error: " + err.Error())
		c.JSON(500, gin.H{
			"error": "Export all units to csv error: " + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"url": url,
	})
}

func LeaveProject(c *gin.Context) {
	machine_name := c.Query("machine_name")
	err := unit_query.LeaveProject(machine_name)
	if err != nil {
		logger.Error("Leave project error: " + err.Error())
		c.JSON(500, gin.H{
			"error": "Leave project error: " + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Leave project success",
	})
}

func JoinProject(c *gin.Context) {
	var req joinRequest
	err := c.BindJSON(&req)
	if err != nil {
		logger.Error("Bind json error: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = unit_query.JoinProject(req.ProjectName, req.Machines)
	if err != nil {
		logger.Error("Join project error: " + err.Error())
		c.JSON(500, gin.H{
			"error": "Join project error: " + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
	})
}
