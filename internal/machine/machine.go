package machine

type Machine struct {
	Hostname         string
	Stream_url       string
	Stream_status    string
	Stream_pid       int
	Stream_interface string
}
type Target struct {
	Hostname string
	Status   string
	Ssim     int
	Wait     int
}

var Recording_machine map[string]Machine
var All_machine map[string]Machine
var Idle_machine map[string]Machine
