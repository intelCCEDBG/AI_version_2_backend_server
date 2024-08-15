package email_query

import (
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func Get_email_threestrike_status(code string) int {
	status := method.QueryRow("SELECT three_mail_constraint FROM project where short_name=?", code)
	var three_mail_constraint int
	err := status.Scan(&three_mail_constraint)
	if err != nil {
		logger.Error("Query three strike status error: " + err.Error())
	}
	return three_mail_constraint
}
