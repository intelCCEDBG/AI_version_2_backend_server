package aikvm_api

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	unit_query "recorder/pkg/mariadb/unit"
	"strings"

	"github.com/gin-gonic/gin"
)

type CheckKvmResponse struct {
	Status        string `json:"status"`
	ProjectName   string `json:"project"`
	KvmHostnName  string `json:"kvm_location"`
	KvmIP         string `json:"kvm_ip"`
	KvmStatus     string `json:"kvm_status"`
	MachineName   string `json:"machine_name"`
	DebugHostIP   string `json:"dbg_ip"`
	DebugHostName string `json:"dbg_hostname"`
}

func CheckKvmUnit(c *gin.Context) {
	ip := c.Query("ip")
	status, unit, err := unit_query.CheckKvmUnit(ip)
	if err != nil {
		logger.Error("Check kvm unit error: " + err.Error())
		c.JSON(500, gin.H{
			"error": "Check kvm unit error: " + err.Error(),
		})
	}
	result := CheckKvmResponse{
		Status:        status,
		ProjectName:   unit.ProjectName,
		KvmHostnName:  unit.KvmHostName,
		KvmIP:         unit.KvmIP,
		KvmStatus:     unit.KvmStatus,
		MachineName:   unit.MachineName,
		DebugHostIP:   unit.DebugHostIP,
		DebugHostName: unit.DebugHostName,
	}
	c.JSON(200, result)
}

func CreateKvmUnit(c *gin.Context) {
	var req structure.CreateKvmUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Create kvm unit error: " + err.Error())
		c.JSON(500, gin.H{
			"error": "Create kvm unit error: " + err.Error(),
		})
		return
	}
	message, err := unit_query.CreateKvmUnit(req)
	if err != nil {
		logger.Error("Create kvm unit error: " + message)
		c.JSON(500, gin.H{
			"error": message,
		})
		return
	}
	if !strings.Contains(message, "Successfully") {
		c.JSON(400, gin.H{
			"error": message,
		})
		return
	}
	c.JSON(200, gin.H{
		"message": message,
	})
}
