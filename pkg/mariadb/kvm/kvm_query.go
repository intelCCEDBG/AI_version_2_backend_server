package kvm_query

import (
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
	"strings"
)

func Update_kvm_status(hostname string, status string) {
	_, err := method.Exec("UPDATE kvm SET stream_status = ? WHERE hostname = ?", status, hostname)
	if err != nil {
		logger.Error("Update kvm status error: " + err.Error())
	}
}

func GetKvmStatus(hostname string) (kvmTemplate structure.Kvm) {
	KVM, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where hostname = " + "'" + hostname + "'")
	if err != nil {
		logger.Error("Query kvm " + hostname + " error: " + err.Error())
	}
	for KVM.Next() {
		err := KVM.Scan(&kvmTemplate.Hostname, &kvmTemplate.StreamUrl, &kvmTemplate.StreamStatus, &kvmTemplate.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}
	return kvmTemplate
}

func Get_recording_kvms() (kvms []structure.Kvm) {
	Recording_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'recording'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for Recording_list.Next() {
		var tmp = structure.Kvm{}
		err := Recording_list.Scan(&tmp.Hostname, &tmp.StreamUrl, &tmp.StreamStatus, &tmp.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func Get_idle_kvms() (kvms []structure.Kvm) {
	Idle_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm where stream_status = 'idle'")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for Idle_list.Next() {
		var tmp = structure.Kvm{}
		err := Idle_list.Scan(&tmp.Hostname, &tmp.StreamUrl, &tmp.StreamStatus, &tmp.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}

func Get_all_kvms() (kvms []structure.Kvm) {
	All_list, err := method.Query("SELECT hostname,stream_url,stream_status,stream_interface FROM kvm")
	if err != nil {
		logger.Error("Query idle kvm error: " + err.Error())
	}
	for All_list.Next() {
		var tmp = structure.Kvm{}
		err := All_list.Scan(&tmp.Hostname, &tmp.StreamUrl, &tmp.StreamStatus, &tmp.StreamInterface)
		if err != nil {
			logger.Error(err.Error())
			return
		}
		kvms = append(kvms, tmp)
	}
	return kvms
}
func Get_all_Floor_from_hostname() map[string]int {
	Floors := make(map[string]int)
	Projects, err := method.Query("SELECT hostname FROM kvm")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
		return Floors
	}
	for Projects.Next() {
		var Floor string
		err = Projects.Scan(&Floor)
		if err != nil {
			logger.Error(err.Error())
			return Floors
		}
		String_list := strings.Split(Floor, "_")
		_, ok := Floors[String_list[0]]
		if ok {
			Floors[String_list[0]]++
		} else {
			Floors[String_list[0]] = 1
		}
	}
	return Floors
}
func Get_hostnames_by_floor(floor string) []string {
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
func Get_link_status(hostnames []string) []bool {
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
func Get_messagecount(hostnames []string) []int {
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
func Insert_message(hostname string, message string) {
	_, err := method.Exec("INSERT INTO kvm_message (hostname, message) VALUES (?,?)", hostname, message)
	if err != nil {
		logger.Error("Inserting message error: " + err.Error())
	}
}
func Delete_message(hostname string, message string) {
	_, err := method.Exec("delete from kvm_message where hostname=? and message=?", hostname, message)
	if err != nil {
		logger.Error("Deleting message error: " + err.Error())
	}
}
func Get_kvm_message(hostname string) []string {
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
	Status := method.QueryRow("SELECT stream_status FROM kvm where hostname = ?", hostname)
	var status string
	err := Status.Scan(&status)
	if err != nil {
		logger.Error("Reading kvm error: " + err.Error())
		return "", err
	}
	return status, nil
}
