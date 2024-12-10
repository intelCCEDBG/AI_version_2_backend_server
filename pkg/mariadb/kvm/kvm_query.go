package kvm_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	"strings"
)

func KvmHostnameExists(hostname string) (bool, error) {
	// see if the kvm hostname exists
	row := method.QueryRow("SELECT ip FROM kvm WHERE hostname = ?", hostname)
	var ip string
	err := row.Scan(&ip)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		logger.Error("Check kvm hostname error: " + err.Error())
		return false, err
	}
	return true, nil
}

func CreateKvm(ip, location, owner string) error {
	query := `
		INSERT INTO kvm (hostname, ip, owner, status, version, NAS_ip, stream_url, stream_status, stream_interface, start_record_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := method.Exec(query, location, ip, owner, "null", 1, "null", "http://"+ip+":8081", "idle", "null", 0)
	if err != nil {
		logger.Error("Insert kvm error: " + err.Error())
		return err
	}
	return nil
}

func SetHighFPS(hostname string, state bool) error {
	_, err := method.Exec("UPDATE kvm SET high_frame_rate = ? WHERE hostname = ?", state, hostname)
	if err != nil {
		logger.Error("Update kvm high frame rate error: " + err.Error())
		return err
	}
	return nil
}

func UpdateStatus(hostname string, status string) {
	_, err := method.Exec("UPDATE kvm SET stream_status = ? WHERE hostname = ?", status, hostname)
	if err != nil {
		logger.Error("Update kvm status error: " + err.Error())
	}
}

func GetStatus(hostname string) (kvm structure.Kvm) {
	data, err := method.Query("SELECT hostname, stream_url, stream_status, stream_interface, high_frame_rate FROM kvm where hostname = " + "'" + hostname + "'")
	if err != nil {
		logger.Error("Query kvm " + hostname + " error: " + err.Error())
	}
	for data.Next() {
		err := data.Scan(&kvm.Hostname, &kvm.StreamUrl, &kvm.StreamStatus, &kvm.StreamInterface, &kvm.HighFrameRate)
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}
	return kvm
}

func GetRecordingKvms() (kvms []structure.Kvm) {
	recordingList, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'recording'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for recordingList.Next() {
		var tmp = structure.Kvm{}
		err := recordingList.Scan(&tmp.Hostname, &tmp.StreamUrl, &tmp.StreamStatus, &tmp.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func GetIdleKvms() (kvms []structure.Kvm) {
	idleList, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'idle'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for idleList.Next() {
		var tmp = structure.Kvm{}
		err := idleList.Scan(&tmp.Hostname, &tmp.StreamUrl, &tmp.StreamStatus, &tmp.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func GetAllKvms() (kvms []structure.Kvm) {
	allList, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for allList.Next() {
		var tmp = structure.Kvm{}
		err := allList.Scan(&tmp.Hostname, &tmp.StreamUrl, &tmp.StreamStatus, &tmp.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func GetAllFloorFromHostname() map[string]int {
	floors := make(map[string]int)
	projects, err := method.Query("SELECT hostname FROM kvm")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
		return floors
	}
	for projects.Next() {
		var Floor string
		err = projects.Scan(&Floor)
		if err != nil {
			logger.Error(err.Error())
			return floors
		}
		stringList := strings.Split(Floor, "_")
		_, ok := floors[stringList[0]]
		if ok {
			floors[stringList[0]]++
		} else {
			floors[stringList[0]] = 1
		}
	}
	return floors
}

func GetHostnamesByFloor(floor string) []string {
	var Hostnames []string
	hostnames, err := method.Query("SELECT hostname FROM kvm")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
		return Hostnames
	}
	for hostnames.Next() {
		var Floor string
		err = hostnames.Scan(&Floor)
		if err != nil {
			logger.Error(err.Error())
			return Hostnames
		}
		String_list := strings.Split(Floor, "_")
		if String_list[0] == floor {
			Hostnames = append(Hostnames, Floor)
		}
	}
	return Hostnames
}
func GetLinkStatus(hostnames []string) []bool {
	var Links []bool
	for _, v := range hostnames {
		exist := method.QueryRow("SELECT EXISTS(SELECT uuid FROM debug_unit WHERE hostname=?)", v)
		var EX bool
		err := exist.Scan(&EX)
		if err != nil {
			logger.Error("Reading debug_unit error: " + err.Error())
			return Links
		}
		Links = append(Links, EX)
	}
	return Links
}
func GetMessageCount(hostnames []string) []int {
	var Counts []int
	for _, v := range hostnames {
		exist := method.QueryRow("SELECT COUNT(message) FROM kvm_message where hostname=?;", v)
		var EX int
		err := exist.Scan(&EX)
		if err != nil {
			logger.Error("Reading debug_unit error: " + err.Error())
			return Counts
		}
		Counts = append(Counts, EX)
	}
	return Counts
}
func InsertMessage(hostname string, message string) {
	_, err := method.Exec("INSERT INTO kvm_message (hostname, message) VALUES (?,?)", hostname, message)
	if err != nil {
		logger.Error("Inserting message error: " + err.Error())
	}
}
func DeleteMessage(hostname string, message string) {
	_, err := method.Exec("delete from kvm_message where hostname=? and message=?", hostname, message)
	if err != nil {
		logger.Error("Deleting message error: " + err.Error())
	}
}
func GetMessage(hostname string) []string {
	var Messages []string
	rows, err := method.Query("SELECT message FROM kvm_message where hostname=?", hostname)
	if err != nil {
		logger.Error("Reading message error: " + err.Error())
	}
	for rows.Next() {
		var Message string
		err = rows.Scan(&Message)
		if err != nil {
			logger.Error(err.Error())
		}
		Messages = append(Messages, Message)
	}
	return Messages
}

func GetIP(hostname string) string {
	IP := method.QueryRow("SELECT ip FROM kvm where hostname = ?", hostname)
	var ip string
	err := IP.Scan(&ip)
	if err != nil {
		logger.Error("Reading kvm error: " + err.Error())
		return ""
	}
	return ip
}

func GetStreamStatus(hostname string) (string, error) {
	data := method.QueryRow("SELECT stream_status FROM kvm where hostname = ?", hostname)
	var status string
	err := data.Scan(&status)
	if err != nil {
		logger.Error("Reading kvm error: " + err.Error())
		return "", err
	}
	return status, nil
}

func DeleteKvm(hostname string) error {
	_, err := method.Exec("DELETE FROM debug_unit WHERE hostname = ?", hostname)
	if err != nil {
		logger.Error("Delete debug_unit error: " + err.Error())
		return err
	}
	_, err = method.Exec("DELETE FROM kvm WHERE hostname = ?", hostname)
	if err != nil {
		logger.Error("Delete kvm error: " + err.Error())
		return err
	}
	_, err = method.Exec("DELETE FROM kvm_message WHERE hostname = ?", hostname)
	if err != nil {
		logger.Error("Delete kvm message error: " + err.Error())
		return err
	}
	return nil
}
