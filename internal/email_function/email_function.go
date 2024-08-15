package emailfunction

import (
	"recorder/internal/structure"
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
