package apiservice

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
}

type Debug_unit struct {
	Hostname     string `json:"kvm_hostname"`
	Ip           string `json:"dbghost_ip"`
	Machine_name string `json:"dut_machine"`
	Project      string `json:"project"`
}

func ResponseWithJson(w http.ResponseWriter, status_code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.WriteHeader(status_code)
	w.Write(response)
}
