package service

import (
	"middleproject/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 点赞帖子
func LikePost(c *gin.Context) {
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

	erro, msg := model.LikePost(uid, postid)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}

// 收藏帖子
func CollectPost(c *gin.Context) {
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

	erro, msg := model.CollectPost(uid, postid)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})

}

// 点赞评论
func LikeComment(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	if err_uid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}
	commentidstr := c.DefaultQuery("comid", "-1")
	commentid, err_cid := strconv.Atoi(commentidstr)
	if err_cid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的评论ID"})
		return
	}
	if uid == -1 || commentid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid或commentid"})
		return
	}

	erro, msg := model.LikeComment(uid, commentid)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}




