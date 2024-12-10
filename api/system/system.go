package system

import (
	"io"
	"net/http"
	"recorder/config"

	"github.com/gin-gonic/gin"
)

func GetCPUStatus(c *gin.Context) {
	query := `/api/v1/query?query=100%20*%20(1%20-%20avg%20by%20(instance)%20(rate(node_cpu_seconds_total{mode="idle"}[1m])))`
	ip := config.Viper.GetString("PROMETHEUS_IP")
	port := config.Viper.GetString("PROMETHEUS_PORT")
	url := "http://" + ip + ":" + port + query
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get response from Prometheus"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}
	c.Data(resp.StatusCode, "application/json", body)
}
