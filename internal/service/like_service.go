package service

import (
	"middleproject/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	typestr := c.DefaultQuery("type", "1")
	likeType, err_type := strconv.Atoi(typestr)
	if err_type != nil || (likeType != 0 && likeType != 1) {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的类型参数"})
		return
	}

	if uid == -1 || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid或logid"})
		return
	}

	var erro error
	var msg string
	if likeType == 1 {
		erro, msg = model.LikePost(uid, postid)
	} else {
		// 取消点赞操作
		erro, msg = model.DislikePost(uid, postid)
	}

	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}

// 收藏或取消收藏帖子
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

	typestr := c.DefaultQuery("type", "1") // 默认为收藏
	favType, err_type := strconv.Atoi(typestr)
	if err_type != nil || (favType != 0 && favType != 1) {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的type参数"})
		return
	}

	if uid == -1 || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid或logid"})
		return
	}

	var erro error
	var msg string
	if favType == 1 {
		erro, msg = model.CollectPost(uid, postid)
	} else {
		erro, msg = model.UncollectPost(uid, postid)
	}

	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}

// 点赞或取消点赞评论
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

	typestr := c.DefaultQuery("type", "1") // 默认为点赞
	likeType, err_type := strconv.Atoi(typestr)
	if err_type != nil || (likeType != 0 && likeType != 1) {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的type参数"})
		return
	}

	if uid == -1 || commentid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid或commentid"})
		return
	}

	var erro error
	var msg string
	if likeType == 1 {
		erro, msg = model.LikeComment(uid, commentid)
	} else {
		erro, msg = model.UnLikeComment(uid, commentid)
	}

	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}
