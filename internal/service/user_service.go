package service

import (
	"fmt"
	"math/rand"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"middleproject/scripts"

	_ "github.com/go-sql-driver/mysql"
)

// register 函数实现
func Register(c *gin.Context) {
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "连接数据库失败"})
	}
	var data model.User
	if err_bind := c.ShouldBindJSON(&data); err_bind != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "注册绑定请求数据失败"})
		return
	}
	//校验最新验证码
	query := "SELECT code FROM verificationcodes WHERE email = ? AND expiration > NOW() ORDER BY expiration DESC LIMIT 1"
	row := db.QueryRow(query, data.Email)
	fmt.Println(row)
	var code string
	err_check := row.Scan(&code)
	fmt.Println("code:", code)
	fmt.Println("data.VerifyCode:", data.VerifyCode)
	if err_check != nil || code != data.VerifyCode {
		c.JSON(400, gin.H{"isok": false, "failreason": "验证码错误"})
		return
	}
	//添加到数据库
	err_re, result, userid := data.CreateUser()
	if err_re != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	}
	//默认头像地址
	avatar_0 := "postImage/image0.png"
	var url = scripts.GetUrl(avatar_0)
	c.JSON(200, gin.H{"isok": true, "uid": userid, "uimage": url})
}

func SendMailInterface(c *gin.Context) {
	var requestData map[string]string
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "绑定请求数据失败"})
		return
	}
	mail, ok := requestData["mail"]
	type_server := c.DefaultQuery("type", "no")
	if !ok {
		c.JSON(400, gin.H{"isok": false, "failreason": "缺少邮箱"})
		return
	}
	//检查mail是否已经注册
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "连接数据库失败"})
	}
	query := "SELECT email FROM Users WHERE email = ?"
	row := db.QueryRow(query, mail)
	var email string
	err_check := row.Scan(&email)
	if err_check == nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "邮箱已经注册"})
		return
	}
	//生成随机数
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(999999-100000+1) + 100000
	strnum := strconv.Itoa(randomNum)
	//strnum := "123456"
	var result string
	if type_server == "regist" {
		result = scripts.SendEmail(mail, "注册验证码", strnum)
	} else if type_server == "find" {
		result = scripts.SendEmail(mail, "找回密码验证码", strnum)
	} else {
		c.JSON(400, gin.H{"isok": false, "failreason": "无效的type"})
		return
	}
	if result != "成功" {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	} else {
		c.JSON(200, gin.H{"isok": true})
	}

}
