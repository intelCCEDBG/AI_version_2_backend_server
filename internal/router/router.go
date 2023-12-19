package router

import (
	"recorder/config"

	dbghost_api "recorder/api/debug_host"
	dbgunit_api "recorder/api/debug_unit"
	dut_api "recorder/api/dut"
	kvm_api "recorder/api/kvm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start_backend() {
	// init setup
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"*"}
	corsConfig.AllowHeaders = []string{"*"}
	router.RedirectFixedPath = true
	router.Use(cors.New(corsConfig))

	Port := config.Viper.GetString("SERVER_PORT")

	router.POST("/api/upload", kvm_api.Kvm_csv_mapping)

	//kvm
	router.GET("/api/kvm/list", kvm_api.Kvm_list)
	router.GET("/api/kvm/info", kvm_api.Kvm_info)
	router.GET("/api/kvm/all_info", kvm_api.Kvm_all_info)
	router.GET("/api/kvm/search", kvm_api.Kvm_search)
	router.POST("/api/kvm/kvm_mapping", kvm_api.Kvm_mapping)
	router.POST("/api/kvm/delete_mapping", kvm_api.Kvm_delete)
	router.POST("/api/kvm/project_status", kvm_api.Project_status)
	router.OPTIONS("/api/kvm/project_status", kvm_api.Project_status)
	router.POST("/api/kvm/stream_status", kvm_api.Kvm_status)
	router.GET("/api/kvm/modify", kvm_api.Kvm_modify)

	//debug_host
	router.GET("/api/dbg/list", dbghost_api.Dbghost_list)
	router.GET("/api/dbg/info", dbghost_api.Dbghost_info)
	router.GET("/api/dbg/all_info", dbghost_api.Dbghost_all_info)
	router.GET("/api/dbg/search", dbghost_api.Dbghost_search)
	router.GET("/api/dbg/modify", dbghost_api.Dbghost_modify)

	//dut
	router.GET("/api/dut/list", dut_api.Dut_list)
	router.GET("/api/dut/info", dut_api.Dut_info)
	router.GET("/api/dut/all_info", dut_api.Dut_all_info)
	router.GET("/api/dut/search", dut_api.Dut_search)
	router.GET("/api/dut/modify", dut_api.Dut_modify)
	router.GET("/api/dut/status", dut_api.Dut_status)


	router.GET("/api/kvm/gen_video", kvm_api.Kvm_genvideo)
	router.OPTIONS("/api/kvm/gen_video", kvm_api.Kvm_genvideo)
	//debug_unit
	router.GET("/api/dbgunit/project_list", dbgunit_api.Project_list)
	router.GET("/api/dbgunit/project_info", dbgunit_api.Project_info)

	router.Run(":" + Port)
}
