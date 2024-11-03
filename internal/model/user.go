package model

import (
	"middleproject/internal/repository"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	UserID           int       `json:"uid"`
	Uname            string    `json:"uname"`
	Phone            string    `json:"phone"`
	Email            string    `json:"mail"`
	Address          string    `json:"address"`
	Password         string    `json:"password"`
	Avatar           string    `json:"avatar"`
	Signature        string    `json:"signature"`
	Birthday         time.Time `json:"birthday"`
	RegistrationDate time.Time `json:"registration_date"`
}

func (u *User) CreateUser() (error, string, string) {
	db, err := repository.Connect()
	if err != nil {
		return err, "数据库连接失败", "0"
	}
	defer db.Close()
	//检查邮箱是否已经注册
	query := "SELECT email FROM Users WHERE email = u.Email"
	row := db.QueryRow(query)
	var email string
	err = row.Scan(&email)
	if err == nil {
		return err, "邮箱已经注册", "0"
	}
	if u.Birthday.IsZero() {
		u.Birthday = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	query = `INSERT INTO Users (Uname, phone, email, address, password, avatar, signature, birthday, registration_date)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, DEFAULT)`

	result, err := db.Exec(query, u.Uname, u.Phone, u.Email, u.Address, u.Password, u.Avatar, u.Signature, u.Birthday)
	if err != nil {
		return err, "用户创建失败", "0"
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return err, "获取用户ID失败", "0"
	}
	u.UserID = int(userID)

	return nil, "注册成功", strconv.Itoa(int(userID))
}
