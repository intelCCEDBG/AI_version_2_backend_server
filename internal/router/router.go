package router

import (
	"io"
	"os"
	"recorder/config"
	"recorder/pkg/logger"

	dbghost_api "recorder/api/debug_host"
	dbgunit_api "recorder/api/debug_unit"
	dut_api "recorder/api/dut"
	email "recorder/api/email"
	kvm_api "recorder/api/kvm"
	"recorder/api/project"
	"recorder/api/system"

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

	router.POST("/api/upload", kvm_api.Kvm_csv_mapping)

	//kvm
	router.GET("/api/kvm/list", kvm_api.Kvm_list)
	router.GET("/api/kvm/info", kvm_api.Kvm_info)
	router.GET("/api/kvm/all_info", kvm_api.Kvm_all_info)
	router.GET("/api/kvm/search", kvm_api.Kvm_search)
	router.POST("/api/kvm/kvm_mapping", kvm_api.Kvm_mapping)
	router.OPTIONS("/api/kvm/kvm_mapping", kvm_api.Kvm_mapping)
	router.POST("/api/kvm/delete_mapping", kvm_api.Kvm_delete)
	router.OPTIONS("/api/kvm/delete_mapping", kvm_api.Kvm_delete)
	router.POST("/api/kvm/project_status", kvm_api.Project_status)
	router.OPTIONS("/api/kvm/project_status", kvm_api.Project_status)
	router.POST("/api/kvm/insert_message", kvm_api.Insert_message)
	router.OPTIONS("/api/kvm/insert_message", kvm_api.Insert_message)
	router.POST("/api/kvm/delete_message", kvm_api.Delete_message)
	router.OPTIONS("/api/kvm/delete_message", kvm_api.Delete_message)
	router.POST("/api/kvm/stream_status", kvm_api.Kvm_status)
	router.OPTIONS("/api/kvm/stream_status", kvm_api.Kvm_status)
	router.GET("/api/kvm/modify", kvm_api.Kvm_modify)
	router.GET("/api/kvm/floor", kvm_api.Get_KVM_Floor)
	router.GET("/api/kvm/hostnames", kvm_api.Get_hostnames_by_floor)
	router.GET("/api/kvm/get_message", kvm_api.Get_kvm_message)

	//debug_host
	router.GET("/api/dbg/list", dbghost_api.DbgHostList)
	router.GET("/api/dbg/freelist", dbghost_api.Dbghost_freelist)
	router.GET("/api/dbg/info", dbghost_api.Dbghost_info)
	router.GET("/api/dbg/all_info", dbghost_api.Dbghost_all_info)
	router.GET("/api/dbg/search", dbghost_api.Dbghost_search)
	router.GET("/api/dbg/modify", dbghost_api.Dbghost_modify)

	//dut
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
	router.OPTIONS("/api/dut/setmachinestatus", dut_api.SetDutMachineStatus)
	router.GET("/api/dut/deleteErrorlog", dut_api.DutDeleteErrorlog)
	router.GET("/api/dut/deleteErrorlogbyproject", dut_api.DutDeleteErrorlogProject)

	router.GET("/api/kvm/gen_video", kvm_api.Kvm_genvideo)
	router.GET("/api/kvm/gen_minute_video", kvm_api.Kvm_genminutevideo)
	router.OPTIONS("/api/kvm/gen_video", kvm_api.Kvm_genvideo)
	//debug_unit
	router.GET("/api/dbgunit/project_list", dbgunit_api.Project_list)
	router.GET("/api/dbgunit/project_info", dbgunit_api.Project_info)

	// email
	// router.POST("/api/email/command_in", email.Email_command_in)
	// router.GET("/api/email/command_out", email.Email_command_get)
	router.GET("/api/report/state", email.Report)
	router.GET("/api/email/enable_mail_constraint", email.Enable_mail_constraint)
	//project
	router.GET("/api/project/get_project_setting", project.Get_Project_setting)
	router.GET("/api/project/project_code", project.Get_Project_setting_by_code)
	router.POST("/api/project/set_project_setting", project.Set_Project_setting)
	router.OPTIONS("/api/project/set_project_setting", project.Set_Project_setting)
	router.POST("/api/project/project_exec", project.Control)
	router.OPTIONS("/api/project/project_exec", project.Control)
	router.GET("/api/project/freeze_switch", project.Freeze_detection)
	router.GET("/api/project/floor", project.Get_Project_Floor)
	router.GET("/api/project/reportstate", project.Get_Report_State)
	router.GET("/api/project/addnewproject", project.Add_new_project)
	router.GET("/api/project/deleteproject", project.Delete_project)
	router.GET("/api/project/ssim_threshold", project.Get_ssim_and_threshold)
	router.POST("/api/project/ssim_threshold", project.Update_ssim_and_threshold)
	router.OPTIONS("/api/project/ssim_threshold", project.Update_ssim_and_threshold)
	router.GET("/api/project/getstarttime", project.Getstarttime)

	router.GET("/api/system/CPU", system.Get_CPU_status)
	router.POST("/api/export/all", dbgunit_api.Save_csv)
	router.POST("/api/freezecheck", dut_api.FreezeCheck)

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

// sudo docker compose up -d --build --force-recreate backend
