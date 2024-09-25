package user_query

import (
	"recorder/config"
	"recorder/internal/structure"
	"recorder/pkg/logger"
	"recorder/pkg/mariadb/method"
)

func InsertInitUser() {
	initUsername := config.Viper.GetString("INIT_USER")
	initEmail := config.Viper.GetString("INIT_EMAIL")
	exists, err := CheckUser(initUsername)
	if err != nil {
		logger.Error("Check init user failed: " + err.Error())
		return
	}
	switch exists {
	case -1:
		AddUser(initUsername, initEmail, 2) // default permission is 2 (admin)
		logger.Info("Init user " + initUsername + " added")
	case 0, 1:
		EditUserPermission(initUsername, 2) // default permission is 2 (admin)
		logger.Info("Init user " + initUsername + " permission updated")
	case 2:
		logger.Info("Init user " + initUsername + " already exists")
	}
}

func CheckUser(username string) (int, error) {
	query := "SELECT permission FROM user_info WHERE username = ?"
	row := method.QueryRow(query, username)
	var permission int
	err := row.Scan(&permission)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return -1, nil
		}
		logger.Error("Check user permission failed: " + err.Error())
		return -1, err
	}
	return permission, nil
}

func AddUser(username string, email string, permission int) error {
	query := "INSERT INTO user_info (username, email, permission) VALUES (?, ?, ?)"
	_, err := method.Exec(query, username, email, permission)
	if err != nil {
		logger.Error("Add user failed: " + err.Error())
		return err
	}
	return nil
}

func ListUser() ([]structure.User, error) {
	query := "SELECT * FROM user_info ORDER BY permission DESC"
	rows, err := method.Query(query)
	if err != nil {
		logger.Error("List users failed: " + err.Error())
		return nil, err
	}
	defer rows.Close()
	var users []structure.User
	for rows.Next() {
		var user structure.User
		err := rows.Scan(&user.Username, &user.Email, &user.Permission)
		if err != nil {
			logger.Error("List users failed: " + err.Error())
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func EditUserPermission(username string, permission int) error {
	query := "UPDATE user_info SET permission = ? WHERE username = ?"
	_, err := method.Exec(query, permission, username)
	if err != nil {
		logger.Error("Edit user permission failed: " + err.Error())
		return err
	}
	return nil
}

func DeleteUser(username string) error {
	query := "DELETE FROM user_info WHERE username = ?"
	_, err := method.Exec(query, username)
	if err != nil {
		logger.Error("Delete user failed: " + err.Error())
		return err
	}
	return nil
}
