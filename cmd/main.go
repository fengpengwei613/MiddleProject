package main

import (
	"fmt"
	"middleproject/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	// POST 请求：发验证码
	r.POST("/api/regist/mail", func(c *gin.Context) {
		fmt.Println("收到发验证码请求")
		service.SendMailInterface(c)
	})
	r.POST("/api/regist", func(c *gin.Context) {
		fmt.Println("收到注册请求")
		service.Register(c)
	})
	//发帖
	r.POST("/api/newlog", func(c *gin.Context) {
		fmt.Println("收到登录请求")
		service.PublishPost(c)
	})
	//发评论
	r.POST("/api/newcomment", func(c *gin.Context) {
		fmt.Println("收到评论请求")
		service.PublishComment(c)
	})
	//登录
	r.POST("/api/login", func(c *gin.Context) {
		fmt.Println("收到登录请求")
		service.Login(c)
	})
	//获取个人设置接口
	r.GET("/api/persetting", service.HandleGetPersonalSettings)

	//更新个人设置接口
	r.POST("/api/persetting/edit", service.UpdatePersonalSettings)

	//忘记密码
	r.POST("/api/forget", service.ForgotPassword)

	//获取个人信息
	r.GET("/api/perinfo", service.GetPersonalInfo)
	//修改个人信息
	r.POST("/api/perinfo/edit", service.UpdatePersonalInfo)

	// 启动 HTTP 服务器
	if err := r.Run(":8080"); err != nil {
		fmt.Println("启动服务器失败:", err)
	}
}
