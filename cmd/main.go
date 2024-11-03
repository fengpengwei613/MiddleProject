package main

import (
	"fmt"
	"math/rand"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"middleproject/scripts"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	// POST 请求：发验证码
	r.POST("/api/regist/mail", func(c *gin.Context) {
		fmt.Println("收到发验证码请求")
		var requestData map[string]string
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
			return
		}
		// 获取请求参数
		mail, ok := requestData["mail"]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求数据缺少 mail 字段"})
			return
		}
		rand.Seed(time.Now().UnixNano())
		randomNum := rand.Intn(999999-100000+1) + 100000
		strnum := strconv.Itoa(randomNum)
		//strnum := "123456"
		result := scripts.SendEmail(mail, "注册验证码", strnum)

		if result != "成功" {
			c.JSON(500, gin.H{"isok": false, "failreason": result})
			return
		}
		c.JSON(200, gin.H{"isok": true})
	})
	r.POST("/api/regist", func(c *gin.Context) {
		fmt.Println("收到注册请求")
		// 获取请求参数
		var requestData map[string]string
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
			return
		}
		mail, ok := requestData["mail"]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求数据缺少 mail 字段"})
			return
		}
		code, ok_c := requestData["code"]
		if !ok_c {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求数据缺少 code 字段"})
			return
		}
		// 验证码校验
		if !repository.VerifyCode(mail, code) {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "验证码错误"})
			return
		}
		// 注册
		user := model.User{}
		user.Email = mail
		user.Uname = requestData["uname"]
		user.Password = requestData["password"]
		err, result, userid := user.CreateUser()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": result})
			return
		}
		c.JSON(http.StatusOK, gin.H{"isok": true, "uid": userid})

	})
	// 启动 HTTP 服务器
	if err := r.Run(":8080"); err != nil {
		fmt.Println("启动服务器失败:", err)
	}
}
