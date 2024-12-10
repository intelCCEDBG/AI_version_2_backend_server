package router

import (
	"io"
	"os"
	"recorder/config"
	"recorder/pkg/logger"

	aikvm_api "recorder/api/aikvm"
	dbghost_api "recorder/api/debug_host"
	dbgunit_api "recorder/api/debug_unit"
	dut_api "recorder/api/dut"
	email "recorder/api/email"
	kvm_api "recorder/api/kvm"
	"recorder/api/project"
	"recorder/api/system"
	user_api "recorder/api/user"

	"github.com/gin-gonic/gin"
)

func Start_backend() {

	// init setup
	gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create(config.Viper.GetString("API_GIN_LOG_FILE"))
	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()
	router.RedirectFixedPath = true
	router.Use(corsMiddleware())
	router.Use(logger.GinLog())

	// user
	router.GET("/api/user/list", user_api.List)
	router.POST("/api/user/add", user_api.Add)
	router.POST("/api/user/check", user_api.Check)
	router.POST("/api/user/edit", user_api.Edit)
	router.POST("/api/user/delete", user_api.Delete)

	// kvm
	router.POST("/api/kvm", kvm_api.AddKvm)
	router.GET("/api/kvm/exists", kvm_api.CheckKvmHostname)
	router.DELETE("/api/kvm", kvm_api.DeleteKvm)
	router.GET("/api/kvm/list", kvm_api.KvmList)
	router.GET("/api/kvm/info", kvm_api.KvmInfo)
	router.GET("/api/kvm/all_info", kvm_api.KvmAllInfo)
	router.GET("/api/kvm/search", kvm_api.KvmSearch)
	router.POST("/api/kvm/kvm_mapping", kvm_api.KvmMapping)
	router.POST("/api/kvm/delete_mapping", kvm_api.KvmDelete)
	router.POST("/api/kvm/project_status", kvm_api.ProjectStatus)
	router.POST("/api/kvm/insert_message", kvm_api.InsertMessage)
	router.POST("/api/kvm/delete_message", kvm_api.DeleteMessage)
	router.POST("/api/kvm/stream_status", kvm_api.KvmStatus)
	router.GET("/api/kvm/modify", kvm_api.KvmModify)
	router.GET("/api/kvm/floor", kvm_api.GetKVMFloor)
	router.GET("/api/kvm/hostnames", kvm_api.GetHostnamesByFloor)
	router.GET("/api/kvm/get_message", kvm_api.GetKvmMessage)
	router.GET("/api/kvm/gen_video", kvm_api.KvmGenVideo)
	router.GET("/api/kvm/gen_minute_video", kvm_api.KvmGenMinuteVideo)
	router.GET("/api/kvm/set_high_fps", kvm_api.SetHighFPS)

	// debug_host
	router.POST("/api/dbg", dbghost_api.AddDbghost)
	router.GET("/api/dbg/free", dbghost_api.CheckDbghostFree)
	router.GET("/api/dbg/list", dbghost_api.DbgHostList)
	router.GET("/api/dbg/freelist", dbghost_api.DbghostFreeList)
	router.GET("/api/dbg/info", dbghost_api.Dbghost_info)
	router.GET("/api/dbg/all_info", dbghost_api.DbghostAllInfo)
	router.GET("/api/dbg/search", dbghost_api.DbghostSearch)
	router.GET("/api/dbg/modify", dbghost_api.DbghostModify)

	// dut
	router.POST("/api/dut", dut_api.AddDut)
	router.GET("/api/dut/free", dut_api.CheckDutFree)
	router.GET("/api/dut/list", dut_api.DutList)
	router.GET("/api/dut/listbyproject", dut_api.ProjectDutList)
	router.GET("/api/dut/freelist", dut_api.DutFreeList)
	router.GET("/api/dut/info", dut_api.DutInfo)
	router.GET("/api/dut/all_info", dut_api.DutAllInfo)
	router.GET("/api/dut/search", dut_api.DutSearch)
	router.GET("/api/dut/modify", dut_api.DutModify)
	router.GET("/api/dut/status", dut_api.DutStatus)
	router.GET("/api/dut/lockframe", dut_api.DutLockCoord)
	router.GET("/api/dut/unlockframe", dut_api.DutUnlockCoord)
	router.GET("/api/dut/islocked", dut_api.DutIslocked)
	router.GET("/api/dut/errorlog", dut_api.DutErrorlog)
	router.GET("/api/dut/seterrorlog", dut_api.SetDutErrorlog)
	router.POST("/api/dut/setmachinestatus", dut_api.SetDutMachineStatus)
	router.GET("/api/dut/deleteErrorlog", dut_api.DutDeleteErrorlog)
	router.GET("/api/dut/deleteErrorlogbyproject", dut_api.DutDeleteErrorlogProject)

	// debug_unit
	router.POST("/api/dbgunit", dbgunit_api.AddDebugUnit)
	router.GET("/api/dbgunit/project_list", dbgunit_api.ProjectList)
	router.GET("/api/dbgunit/project_info", dbgunit_api.ProjectInfo)
	router.GET("/api/dbgunit/leave_project", dbgunit_api.LeaveProject)
	router.POST("/api/dbgunit/join_project", dbgunit_api.JoinProject)

	// email
	// router.POST("/api/email/command_in", email.Email_command_in)
	// router.GET("/api/email/command_out", email.Email_command_get)
	router.GET("/api/report/state", email.Report)
	router.GET("/api/email/enable_mail_constraint", email.EnableMailConstraint)

	// project
	router.GET("/api/project/exists", project.CheckProject)
	router.GET("/api/project/get_project_setting", project.GetProjectSetting)
	router.GET("/api/project/project_code", project.GetProjectSettingByCode)
	router.POST("/api/project/set_project_setting", project.SetProjectSetting)
	router.OPTIONS("/api/project/set_project_setting", project.SetProjectSetting)
	router.POST("/api/project/project_exec", project.Control)
	router.OPTIONS("/api/project/project_exec", project.Control)
	router.GET("/api/project/freeze_switch", project.FreezeDetection)
	router.GET("/api/project/floor", project.GetProjectFloor)
	router.GET("/api/project/reportstate", project.GetReportState)
	router.GET("/api/project/addnewproject", project.AddNewProject)
	router.GET("/api/project/deleteproject", project.DeleteProject)
	router.GET("/api/project/ssim_threshold", project.GetSsimAndThreshold)
	router.POST("/api/project/ssim_threshold", project.UpdateSsimAndThreshold)
	router.OPTIONS("/api/project/ssim_threshold", project.UpdateSsimAndThreshold)
	router.GET("/api/project/getstarttime", project.GetStartTime)

	router.GET("/api/system/CPU", system.GetCPUStatus)
	router.POST("/api/export", dbgunit_api.DownloadCsv)
	router.POST("/api/upload", kvm_api.KvmCsvMapping)
	router.POST("/api/freezecheck", dut_api.FreezeCheck)

	// AI-KVM
	router.GET("/api/aikvm/check", aikvm_api.CheckKvmUnit)
	router.POST("/api/aikvm/create", aikvm_api.CreateKvmUnit)

	port := config.Viper.GetString("SERVER_PORT")
	router.Run(":" + port)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
