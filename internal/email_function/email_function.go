package emailfunction

import (
	"net/http"
	"recorder/config"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"strconv"
	"strings"
)

func String_to_Email(email_string string) []structure.Email_tamplate {
	Out_email_list := []structure.Email_tamplate{}
	var Out_email_template structure.Email_tamplate
	email_list := strings.Split(email_string, " ")
	for _, S := range email_list {
		t := strings.Split(S, ",")
		if len(t) == 3 {
			Out_email_template.Account = t[0]
			Out_email_template.Report, _ = strconv.ParseBool(t[1])
			Out_email_template.Alert, _ = strconv.ParseBool(t[2])
			Out_email_list = append(Out_email_list, Out_email_template)
		}
	}
	return Out_email_list
}
func Email_to_string(List []structure.Email_tamplate) string {
	var Email_list string
	for i, s := range List {
		if i == 0 {
			Email_list = s.Account + "," + strconv.FormatBool(s.Report) + "," + strconv.FormatBool(s.Alert)
		} else {
			Email_list = Email_list + " " + s.Account + "," + strconv.FormatBool(s.Report) + "," + strconv.FormatBool(s.Alert)
		}
	}
	return Email_list
}
func Send_alert_mail(machine_name string, code string, errortype int) {
	Ip := config.Viper.GetString("email_server_ip")
	Port := config.Viper.GetString("email_port")
	_, err := http.Get("http://" + Ip + ":" + Port + "/api/mail/send_alert?project=" + code + "&machine_name=" + machine_name + "&type=" + strconv.Itoa(errortype))
	if err != nil {
		logger.Error(err.Error())
	}

}
