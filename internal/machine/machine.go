package machine

import "recorder/pkg/redis"

type Machine struct {
	Hostname          string
	Stream_url        string
	Stream_status     string
	Stream_interface  string
	Start_record_time int64
}
type Target struct {
	Hostname string
	Status   string
	Ssim     int
	Wait     int
}

var Recording_machine map[string]*Machine
var All_machine map[string]*Machine
var Idle_machine map[string]*Machine

func Add(machine Machine) {
	var Machine = new(Machine)
	Machine.Hostname = machine.Hostname
	Machine.Stream_url = machine.Stream_url
	Machine.Stream_status = machine.Stream_status
	Machine.Stream_interface = machine.Stream_interface
	Machine.Start_record_time = machine.Start_record_time
	All_machine[machine.Hostname] = Machine
	if machine.Stream_status == "idle" {
		redis.Redis_append("idle_machine_list", machine.Hostname+",")
		Idle_machine[machine.Hostname] = Machine
	} else {
		redis.Redis_append("recording_machine_list", machine.Hostname+",")
		Recording_machine[machine.Hostname] = Machine
	}
}

func Remove(hostname string) {
	delete(All_machine, hostname)
	// To Do
}

func RecordtoIdle(hostname string) {
	Idle_machine[hostname] = Recording_machine[hostname]
	delete(Recording_machine, hostname)
}

func IdletoRecord(hostname string) {
	Recording_machine[hostname] = Idle_machine[hostname]
	delete(Idle_machine, hostname)
}

func Get(hostname string) Machine {
	return *All_machine[hostname]
}
