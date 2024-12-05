package model

import (
	"database/sql"
	"errors"
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
	fmt.Println("data12312312")
	if err_conn != nil {
		return err_conn, "创建新用户连接数据库失败", "0"
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败", "0"
	}
	fmt.Println("data12312312")
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
	fmt.Println("data12312312")
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
	Userid   int    `json:"user_id"`
	Password string `json:"password"`
}

type ResetPasswordReq struct {
	Password string `json:"password" binding:"required"`
	Mail     string `json:"mail" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
}

type UpdatePersonalInfoRequest struct {
	UserName  string `json:"userName" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Address   string `json:"address"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Birthday  string `json:"birthday"`
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

// 获取用户信息
func (u *User) GetUserInfo(userID string) (error, *User) {
	db, err := repository.Connect()
	if err != nil {
		return err, nil
	}
	defer db.Close()

	// 查询用户信息
	query := `SELECT user_id, uname, phone, email, address, avatar, signature, birthday, registration_date
              FROM Users WHERE user_id = ?`
	row := db.QueryRow(query, userID)

	var user User
	err = row.Scan(&user.UserID, &user.Uname, &user.Phone, &user.Email, &user.Address, &user.Avatar, &user.Signature, &user.Birthday, &user.RegistrationDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("用户不存在"), nil
		}
		return err, nil
	}

	return nil, &user
}

// 更新用户信息
func (u *User) UpdateUserInfo() (error, string) {
	db, err := repository.Connect()
	if err != nil {
		return err, "数据库连接失败"
	}
	defer db.Close()

	// 更新用户信息
	query := `UPDATE Users
              SET uname = ?, phone = ?, address = ?, avatar = ?, signature = ?, birthday = ?
              WHERE user_id = ?`
	_, err = db.Exec(query, u.Uname, u.Phone, u.Address, u.Avatar, u.Signature, u.Birthday, u.UserID)
	if err != nil {
		return err, "更新个人信息失败"
	}

	return nil, "个人信息更新成功"
}
