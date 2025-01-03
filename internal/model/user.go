package model

import (
	"fmt"
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
	VerifyCode       string    `json:"code"`
}

func (u *User) CreateUser() (error, string, string) {
	db_link, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "创建新用户连接数据库失败", "0"
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败", "0"
	}
	//检查邮箱是否已经注册
	query := "SELECT email FROM Users WHERE email = ?"
	row := db.QueryRow(query, u.Email)
	var email string
	err_check := row.Scan(&email)
	if err_check == nil {
		db.Rollback()
		return err_check, "邮箱已经注册", "0"
	}
	fmt.Println("data12312312")
	query = `INSERT INTO Users (Uname, email, password, avatar)
              VALUES (?, ?, ?, ?)`

	result, err_insert := db.Exec(query, u.Uname, u.Email, u.Password, "postImage/image0.png")
	if err_insert != nil {
		db.Rollback()
		return err_insert, "sql语句用户创建失败", "0"
	}
	userID, err_id := result.LastInsertId()
	if err_id != nil {
		db.Rollback()
		return err_id, "获取新用户ID失败", "0"
	}
	u.UserID = int(userID)
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "事务提交失败", "0"
	}
	fmt.Println("用户注册成功")
	return nil, "注册成功", strconv.Itoa(int(userID))
}

// 个人设置结构体
type PersonalSettings struct {
	ShowLike    bool
	ShowCollect bool
	ShowPhone   bool
	ShowMail    bool
}

type UpdatePersonalSettings struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	UserId string `json:"uid"`
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Userid   string `json:"uid"`
	Password string `json:"password"`
}

type ResetPasswordReq struct {
	Password string `json:"password" binding:"required"`
	Mail     string `json:"mail" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
}

// 更新密码
func (u *User) UpdatePassword(email, newPassword string) (error, string) {
	db, err := repository.Connect()
	if err != nil {
		return err, "数据库连接失败"
	}
	defer db.Close()

	query := "UPDATE Users SET password = ? WHERE email = ?"
	_, err = db.Exec(query, newPassword, email)
	if err != nil {
		return err, "更新密码失败"
	}

	return nil, "密码更新成功"
}

type PersonalInfo struct {
	Signature    string   `json:"persign"`
	UserID       string   `json:"userID"`
	UserName     string   `json:"uname"`
	UImage       string   `json:"uimage"`
	Phone        string   `json:"phone"`
	Mail         string   `json:"mail"`
	Address      string   `json:"address"`
	Birthday     string   `json:"birthday"`
	RegTime      string   `json:"regtime"`
	Sex          string   `json:"sex"`
	Introduction string   `json:"introduction"`
	SchoolName   string   `json:"schoolname"`
	Major        string   `json:"major"`
	EduTime      string   `json:"edutime"`
	EduLevel     string   `json:"edulevel"`
	CompanyName  string   `json:"companyname"`
	PositionName string   `json:"positionname"`
	Industry     string   `json:"industry"`
	Interests    []string `json:"interests"`
	LikeNum      string   `json:"likenum"`
	AttionNum    string   `json:"attionnum"`
	IsAttion     bool     `json:"isattion"`
	FansNum      string   `json:"fansnum"`
}


type ChangeEmail struct {
	Uid string `json:"uid"`
	Code string `json:"code"`
	NewMail string `json:"newemail"`

}