package kvm

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"recorder/pkg/logger"
	kvm_query "recorder/pkg/mariadb/kvm"
	"time"

	"github.com/gorilla/websocket"
)

type auth_response struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

type Action struct {
	EventType string `json:"event_type"`
	Event     event  `json:"event"`
}

type event struct {
	Key   string `json:"key"`
	State bool   `json:"state"`
}

func RecordtoIdle(hostname string) {
	// Idle_kvm[hostname] = Recording_kvm[hostname]
	// delete(Recording_kvm, hostname)
	kvm_query.UpdateStatus(hostname, "idle")
}
func ErrortoIdle(hostname string) {
	// Idle_kvm[hostname] = Error_kvm[hostname]
	// delete(Error_kvm, hostname)
	kvm_query.UpdateStatus(hostname, "idle")
}
func ErrortoRecord(hostname string) {
	// Recording_kvm[hostname] = Error_kvm[hostname]
	// delete(Error_kvm, hostname)
	kvm_query.UpdateStatus(hostname, "recording")
}
func IdletoError(hostname string) {
	// Error_kvm[hostname] = Idle_kvm[hostname]
	// delete(Idle_kvm, hostname)
	kvm_query.UpdateStatus(hostname, "error")
}
func RecordtoError(hostname string) {
	// Error_kvm[hostname] = Recording_kvm[hostname]
	// delete(Recording_kvm, hostname)
	kvm_query.UpdateStatus(hostname, "error")
}
func IdletoRecord(hostname string) {
	// Recording_kvm[hostname] = Idle_kvm[hostname]
	// delete(Idle_kvm, hostname)
	kvm_query.UpdateStatus(hostname, "recording")
}

func SwitchStreamStatus(ip, status string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 7 * time.Second}
	url := fmt.Sprintf("https://%s:8443/api/switch_mode?mode=", ip)
	switch status {
	case "recording":
		url += "motion"
	case "idle":
		url += "kvmd"
	default:
		logger.Error("Invalid status: " + status)
		return
	}
	_, err := client.Get(url)
	if err != nil {
		logger.Error(err.Error())
	}
}

func Auth(ip string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 7 * time.Second}
	body := []byte(`user=admin&passwd=admin`)
	req, err := http.NewRequest("POST", "https://"+ip+"/api/auth/login", bytes.NewBuffer(body))
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}
	defer resp.Body.Close()
	var auth auth_response
	err = json.NewDecoder(resp.Body).Decode(&auth)
	if err != nil {
		logger.Error(err.Error())
		return "", err
	}
	if !auth.OK {
		logger.Error("auth failed")
		return "", err
	}
	// get cookie auth token
	auth_token := ""
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "auth_token" {
			auth_token = cookie.Value
		}
	}
	return auth_token, nil
}

func PressWindowsKey(ip string) error {
	// auth
	auth_token, err := Auth(ip)
	if err != nil {
		logger.Error("Error getting auth token: " + err.Error())
		return err
	}
	// connect to websocket
	ws_url := fmt.Sprintf("wss://%s/api/ws", ip)
	cookieHeader := fmt.Sprintf("auth_token=%s", auth_token)
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	header := make(http.Header)
	header.Add("Cookie", cookieHeader)
	conn, _, err := dialer.Dial(ws_url, header)
	if err != nil {
		logger.Error("Error connecting to kvm ws: " + err.Error())
		return err
	}
	defer conn.Close()
	// send key press action
	action := Action{
		EventType: "key",
		Event: event{
			Key:   "MetaLeft",
			State: true,
		},
	}
	err = conn.WriteJSON(action)
	if err != nil {
		logger.Error("Error sending key press action: " + err.Error())
		return err
	}
	time.Sleep(200 * time.Millisecond)
	action.Event.State = false
	err = conn.WriteJSON(action)
	if err != nil {
		logger.Error("Error sending key release action: " + err.Error())
		return err
	}
	// close websocket connection
	time.Sleep(200 * time.Millisecond)
	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		logger.Error("Error closing websocket connection: " + err.Error())
		return err
	}
	return nil
}
