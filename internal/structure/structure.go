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
	Coord    string `json:"coord"`
}
