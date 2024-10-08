package project_query

import (
	emailfunction "recorder/internal/email_function"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	dut_query "recorder/pkg/mariadb/dut"
	kvm_query "recorder/pkg/mariadb/kvm"
	"recorder/pkg/mariadb/method"
	"strings"
	"time"
)

func CreateProject(name, code string) error {
	query := `
		INSERT INTO project (project_name, short_name, owner, email_list, status, freeze_detection)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := method.Exec(query, name, code, "", "", 0, "open")
	if err != nil {
		logger.Error("Insert project error: " + err.Error())
		return err
	}
	return nil
}

func ProjectCodeExists(code string) (bool, error) {
	// see if the project code exists
	row := method.QueryRow("SELECT project_name FROM project WHERE short_name = ?", code)
	var project string
	err := row.Scan(&project)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		logger.Error("Check project code error: " + err.Error())
		return false, err
	}
	return true, nil
}

func Update_project_setting(setting structure.ProjectTemplate, Email_string string) {
	_, err := method.Exec("UPDATE project SET short_name = ?, owner=?, email_list=? WHERE project_name = ?", setting.ShortName, setting.Owner, Email_string, setting.ProjectName)
	if err != nil {
		logger.Error("Update project status error: " + err.Error())
	}
}
func Get_all_projects_setting() []structure.ProjectTemplate {
	Project, err := method.Query("SELECT project_name,short_name,owner,email_list FROM project")
	if err != nil {
		logger.Error("Query all project error: " + err.Error())
	}
	var project_list []structure.ProjectTemplate
	for Project.Next() {
		var project_setting structure.ProjectTemplate
		var email_string string
		err := Project.Scan(&project_setting.ProjectName, &project_setting.ShortName, &project_setting.Owner, &email_string)
		if err != nil {
			logger.Error(err.Error())
			return project_list
		}
		project_setting.EmailList = emailfunction.String_to_Email(email_string)
		project_list = append(project_list, project_setting)
	}
	return project_list
}
func Get_all_Floor() map[string]int {
	Floors := make(map[string]int)
	Projects, err := method.Query("SELECT project_name FROM project")
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
func GetDuts(project string) []structure.DUT {
	Dut, err := method.Query("SELECT machine_name FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var DUTS []structure.DUT
	for Dut.Next() {
		var Tmp string
		err = Dut.Scan(&Tmp)
		if err != nil {
			logger.Error(err.Error())
			return DUTS
		}
		d := dut_query.GetDutStatus(Tmp)
		DUTS = append(DUTS, d)
	}
	return DUTS
}
func Get_kvms(project string) []structure.Kvm {
	Kvm, err := method.Query("SELECT hostname FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var KVMS []structure.Kvm
	for Kvm.Next() {
		var Tmp string
		err = Kvm.Scan(&Tmp)
		if err != nil {
			logger.Error(err.Error())
			return KVMS
		}
		d := kvm_query.GetKvmStatus(Tmp)
		KVMS = append(KVMS, d)
	}
	return KVMS
}
func Get_dbgs(project string) []string {
	Dut, err := method.Query("SELECT ip FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var IPs []string
	for Dut.Next() {
		var Tmp string
		err = Dut.Scan(&Tmp)
		IPs = append(IPs, Tmp)
	}
	return IPs
}
func Get_Units(project string) []structure.UnitDetail {
	Unit, err := method.Query("SELECT machine_name, hostname, ip FROM debug_unit where project=?", project)
	if err != nil {
		logger.Error("Query dut from project error: " + err.Error())
	}
	var UNITS []structure.UnitDetail
	for Unit.Next() {
		var unit structure.UnitDetail
		var Tmp, Tmp2, Tmp3 string
		err = Unit.Scan(&Tmp, &Tmp2, &Tmp3)
		if err != nil {
			logger.Error(err.Error())
			return UNITS
		}
		Detail := method.QueryRow("SELECT test_item, sku, image, bios, config FROM machine_status where machine_name=?", Tmp)
		err = Detail.Scan(&unit.TestItem, &unit.Sku, &unit.Image, &unit.Bios, &unit.Config)
		if err != nil {
			logger.Error(err.Error())
		}
		unit.MachineName = dut_query.GetDutStatus(Tmp)
		unit.Ip = Tmp3
		unit.Hostname = kvm_query.GetKvmStatus(Tmp2)
		unit.Project = project
		UNITS = append(UNITS, unit)
	}
	return UNITS
}
func Get_ssim_and_threshold(project string) structure.SsimAndThreshold {
	var Resp structure.SsimAndThreshold
	row := method.QueryRow("SELECT A.ssim, A.threshold from machine A, debug_unit B WHERE B.project = ? && A.machine_name = B.machine_name limit 1;", project)
	err := row.Scan(&Resp.Ssim, &Resp.Thresh)
	Resp.Project = project
	if err != nil {
		logger.Error("Reading ssim status error: " + err.Error())
	}
	return Resp
}
func Update_ssim_and_threshold(Resp structure.SsimAndThreshold) {
	_, err := method.Exec("UPDATE machine A, debug_unit B SET A.ssim = ?, A.threshold = ? WHERE B.project = ? && A.machine_name = B.machine_name;", Resp.Ssim, Resp.Thresh, Resp.Project)
	if err != nil {
		logger.Error("Update ssim status error: " + err.Error())
	}
}
func GetUpperBound(project string) int {
	var upper_bound int
	row := method.QueryRow("SELECT upper_bound from project where project_name = ?", project)
	err := row.Scan(&upper_bound)
	if err != nil {
		logger.Error("Reading upper bound error: " + err.Error())
	}
	return upper_bound
}
func Delete_project(project string) {
	_, err := method.Exec("DELETE FROM project WHERE project_name = ?", project)
	if err != nil {
		logger.Error("Delete project error: " + err.Error())
	}
	_, err = method.Exec("DELETE FROM debug_unit WHERE project = ?", project)
	if err != nil {
		logger.Error("Delete project error: " + err.Error())
	}
}
func Add_start_time(project string) {
	_, err := method.Exec("UPDATE project SET start_time = ? WHERE project_name = ?", time.Now().Format("2006-01-02 15:04:05"), project)
	if err != nil {
		logger.Error("INSERT project error: " + err.Error())
	}
}
func GetStartTime(project string) string {
	var start_time string
	row := method.QueryRow("SELECT start_time from project where project_name = ?", project)
	err := row.Scan(&start_time)
	if err != nil {
		logger.Error("Reading start time error: " + err.Error())
	}
	return start_time
}
