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
	//发回复
	r.POST("/api/newreply", func(c *gin.Context) {
		fmt.Println("收到回复请求")
		service.PublishReply(c)
	})
	r.GET("/api/logs", func(c *gin.Context) {
		fmt.Println("收到获取帖子请求")
		service.GetRecommendPost(c)
	})
	//获取帖子详情
	r.GET("/api/logs/alog", func(c *gin.Context) {
		fmt.Println("收到获取帖子详情请求")
		service.GetPostInfo(c)
	})
	//获取帖子图片
	r.GET("/api/logs/image", func(c *gin.Context) {
		fmt.Println("收到获取帖子图片请求")
		service.GetPostImage(c)
	})
	//获取更多评论
	r.GET("/api/morecom", func(c *gin.Context) {
		fmt.Println("收到获取更多评论请求")
		service.GetMoreComment(c)
	})
	//获取更多回复
	r.GET("/api/morereply", func(c *gin.Context) {
		fmt.Println("收到获取更多回复请求")
		service.GetMoreReply(c)
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
	r.GET("/api/perinfo", service.HandleGetPersonalInfo)
	//修改个人信息
	r.POST("/api/perinfo/edit", service.UpdatePersonalInfo)
	//点赞帖子
	r.POST("/api/likelog", service.LikePost)
	//收藏帖子
	r.POST("/api/collectlog", service.CollectPost)
	//喜欢评论
	r.POST("/api/likecomment", service.LikeComment)
	//关注用户
	r.POST("/api/attionusr", service.HandleFollowAction)
	//喜欢回复
	r.POST("/api/likereply", service.LikeReply)
	//举报接口
	r.POST("/api/report", service.HandleReport)
	//获取个人发帖
	r.GET("/api/perlogs", service.GetPersonalPostLogs)

	//获取个人喜欢帖子
	r.GET("/api/perlikelogs", service.GetPersonalLikePosts)

	//获取个人收藏帖子
	r.GET("/api/percollectlogs", service.GetPersonalCollectPosts)

	//删帖
	r.POST("/api/deletelog", service.DeletePost)
	//删除评论
	r.POST("/api/deletereply", service.DeleteReply)
	//删除回复
	r.POST("/api/deletecomment", service.DeleteComment)

	// 启动 HTTP 服务器
	if err := r.Run(":8080"); err != nil {
		fmt.Println("启动服务器失败:", err)
	}
}
