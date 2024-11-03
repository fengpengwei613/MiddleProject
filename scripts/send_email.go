package scripts
import (
	"fmt"
	"middleproject/internal/repository"
	"net/smtp"
	"time"
)

func SendEmail(to string, subject string, body string) string {
    db, err := repository.Connect()
	if err != nil {
		return "数据库连接失败"
	}
	defer db.Close()
    query_s := "SELECT email FROM Users WHERE email = ?"
	row := db.QueryRow(query_s, to)
	var email string
	err = row.Scan(&email)
	if err == nil {
		return "邮箱已经注册"
	}
	// 发送方的邮箱和密码
	from := "code_rode@163.com"    // 替换为你的邮箱
	password := "WKnRYHBepKeXafVH" // 替换为你的邮箱密码

	// SMTP 服务器设置
	smtpHost := "smtp.163.com" // 替换为你的 SMTP 服务器地址
	smtpPort := "25"           // SMTP 端口

	// 构建邮件内容
	message := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, body))

	// 身份验证
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// 发送邮件
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return "邮件发送失败"
	}
	
	query := "INSERT INTO verificationcodes values(?,?,?)"
	//5分钟有效期
	currentTime := time.Now()
    chinaTime := currentTime.Add(8 * time.Hour)
	newTime := chinaTime.Add(5 * time.Minute)

	_, err = db.Exec(query, to, body, newTime)
	if err != nil {
		return "验证码存储失败"
	}
	fmt.Println("邮件发送成功")
	return "成功"
}

