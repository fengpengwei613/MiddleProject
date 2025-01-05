package scripts

import (
	"fmt"
	"middleproject/internal/repository"
	"net/smtp"
	"time"
)

func SendEmail(to string, subject string, body string, model string) string {
	db_link, err := repository.Connect()
	if err != nil {
		return "数据库连接失败"
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return "事务开启失败"
	}
	query_s := "SELECT email FROM Users WHERE email = ?"
	row := db.QueryRow(query_s, to)
	var email string
	err = row.Scan(&email)
	if model == "regist" {

		if err == nil {
			db.Rollback()
			return "邮箱已经注册"
		}
	} else if model == "forget" {
		if err != nil {
			db.Rollback()
			return "邮箱未注册"
		}
	}
	// 发送方的邮箱和密码
	from := "code_rode@163.com"
	password := "WKnRYHBepKeXafVH"
	// SMTP 服务器设置
	smtpHost := "smtp.163.com"
	smtpPort := "25"
	// 构建邮件内容
	message := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body))
	// 身份验证
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// 发送邮件
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		db.Rollback()
		return "邮件发送失败"
	}

	query := "INSERT INTO verificationcodes values(?,?,?)"
	//5分钟有效期
	location, _ := time.LoadLocation("Asia/Shanghai") // 设置时区为上海
	time.Local = location
	currentTime := time.Now()
	chinaTime := currentTime
	fmt.Println(chinaTime)
	newTime := chinaTime.Add(5 * time.Minute)
	fmt.Println(newTime)
	timestr := newTime.Format("2006-01-02 15:04:05")

	_, err = db.Exec(query, to, body, timestr)
	if err != nil {
		db.Rollback()
		return "验证码存储失败"
	}
	fmt.Println("邮件发送成功")
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return "事务提交失败"
	}
	return "成功"
}
