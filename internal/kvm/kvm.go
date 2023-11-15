package kvm

import (
	kvm_query "recorder/pkg/mariadb/kvm"
)

// var Recording_kvm map[string]*Kvm
// var All_kvm map[string]*Kvm
// var Idle_kvm map[string]*Kvm
// var Error_kvm map[string]*Kvm

// func Add(kvm Kvm) {
// 	var Kvm = new(Kvm)
// 	Kvm.Hostname = kvm.Hostname
// 	Kvm.Stream_url = kvm.Stream_url
// 	Kvm.Stream_status = kvm.Stream_status
// 	Kvm.Stream_interface = kvm.Stream_interface
// 	Kvm.Start_record_time = kvm.Start_record_time
// 	All_kvm[kvm.Hostname] = Kvm
// 	if kvm.Stream_status == "idle" {
// 		redis.Redis_append("idle_kvm_list", kvm.Hostname+",")
// 		Idle_kvm[kvm.Hostname] = Kvm
// 	} else {
// 		redis.Redis_append("recording_kvm_list", kvm.Hostname+",")
// 		Recording_kvm[kvm.Hostname] = Kvm
// 	}
// }

// func Remove(hostname string) {
// 	delete(All_kvm, hostname)
// 	// To Do
// }

func RecordtoIdle(hostname string) {
	// Idle_kvm[hostname] = Recording_kvm[hostname]
	// delete(Recording_kvm, hostname)
	kvm_query.Update_kvm_status(hostname, "idle")
}
func ErrortoIdle(hostname string) {
	// Idle_kvm[hostname] = Error_kvm[hostname]
	// delete(Error_kvm, hostname)
	kvm_query.Update_kvm_status(hostname, "idle")
}
func ErrortoRecord(hostname string) {
	// Recording_kvm[hostname] = Error_kvm[hostname]
	// delete(Error_kvm, hostname)
	kvm_query.Update_kvm_status(hostname, "recording")
}
func IdletoError(hostname string) {
	// Error_kvm[hostname] = Idle_kvm[hostname]
	// delete(Idle_kvm, hostname)
	kvm_query.Update_kvm_status(hostname, "error")
}
func RecordtoError(hostname string) {
	// Error_kvm[hostname] = Recording_kvm[hostname]
	// delete(Recording_kvm, hostname)
	kvm_query.Update_kvm_status(hostname, "error")
}
func IdletoRecord(hostname string) {
	// Recording_kvm[hostname] = Idle_kvm[hostname]
	// delete(Idle_kvm, hostname)
	kvm_query.Update_kvm_status(hostname, "recording")
}
