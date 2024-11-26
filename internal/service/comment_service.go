package service

import (
	"fmt"
	"middleproject/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 发评论接口
func PublishComment(c *gin.Context) {
	var data model.Comment
	if err_bind := c.ShouldBindJSON(&data); err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "发评论绑定请求数据失败"})
		return
	}
	uidstr := c.DefaultQuery("uid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	if err_uid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}
	postidstr := c.DefaultQuery("logid", "-1")
	postid, err_pid := strconv.Atoi(postidstr)
	if err_pid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的帖子ID"})
		return
	}
	if uid == -1 || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid或logid"})
		return
	}

	data.CommenterID = uid
	data.PostID = postid
	erro, msg, idstr := data.AddComment()
	//fmt.Println(msg)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "id": idstr})
	fmt.Println("返回的消息：", idstr)
}
