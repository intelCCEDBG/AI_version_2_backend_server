package apiservice

import (
	"encoding/json"
	"net/http"
)

type ApiResponse struct {
	ResultCode    string
	ResultMessage interface{}
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