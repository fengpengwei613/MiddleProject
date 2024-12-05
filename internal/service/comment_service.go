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
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "id": idstr})
	fmt.Println("返回的消息：", idstr)
}

func PublishReply(c *gin.Context) {
	data := map[string]string{}
	if err_bind := c.ShouldBindJSON(&data); err_bind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "发回复绑定请求数据失败"})
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
	commentidstr := c.DefaultQuery("comid", "-1")
	commentid, err_cid := strconv.Atoi(commentidstr)
	if err_cid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的评论ID"})
		return
	}
	replyIDstr := c.DefaultQuery("replyid", "-1")
	replyID, err_re := strconv.Atoi(replyIDstr)
	if err_re != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的回复ID"})
		return
	}
	if uid == -1 || postid == -1 || commentid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid或logid或comid"})
		return
	}
	if replyID == -1 {
		erro, msg, idstr := model.AddReply(uid, postid, commentid, data["content"])
		if erro != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
			return
		}
		c.JSON(http.StatusOK, gin.H{"isok": true, "id": idstr})
		return
	}

	erro, msg, idstr := model.AddReply(uid, postid, replyID, data["content"])
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "id": idstr})

}
