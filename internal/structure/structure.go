package structure

type Kvm struct {
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
type Result_message struct {
	Hostname string `json:"hostname"`
	Label    int64  `json:"label"`
	Coords   string `json:"coords"`
	Created  int64  `json:"created"`
	Expire   int64  `json:"expire_time"`
}
type DUT struct {
	Machine_name string
	Ssim         int
	Cycle_cnt    int
	Status       string
	Threshhold   int
}
