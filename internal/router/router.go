package router

import (
	"recorder/config"

	dbghost_api "recorder/api/debug_host"
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

	// GET request
	//kvm
	router.GET("/api/kvm/list", kvm_api.Kvm_list)
	router.GET("/api/kvm/info", kvm_api.Kvm_info)
	router.GET("/api/kvm/search", kvm_api.Kvm_search)

	//debug_host
	router.GET("/api/dbg/list", dbghost_api.Dbghost_list)
	router.GET("/api/dbg/info", dbghost_api.Dbghost_info)
	router.GET("/api/dbg/search", dbghost_api.Dbghost_search)

	//dut
	router.GET("/api/dut/list", dut_api.Dut_list)
	router.GET("/api/dut/info", dut_api.Dut_info)
	router.GET("/api/dut/search", dut_api.Dut_search)

	router.POST("/api/kvm/kvm_mapping", kvm_api.Kvm_mapping)

	router.Run(":" + Port)
}
