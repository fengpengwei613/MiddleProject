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

func GetMoreComment(c *gin.Context) {
	postidstr := c.DefaultQuery("logid", "-1")
	nowcommentstr := c.DefaultQuery("nowcomment", "-1")
	uidstr := c.DefaultQuery("uid", "-1")
	postid, err_pid := strconv.Atoi(postidstr)
	nowcomment, err_now := strconv.Atoi(nowcommentstr)
	uid, err_uid := strconv.Atoi(uidstr)
	if err_pid != nil || err_now != nil || err_uid != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	if postid == -1 || nowcomment == -1 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	err, posts := GetCommentInfo(nowcomment, postid, uid, -1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": posts})
}

func GetMoreReply(c *gin.Context) {
	commentidstr := c.DefaultQuery("comid", "-1")
	logidstr := c.DefaultQuery("logid", "-1")
	nowreplystr := c.DefaultQuery("nowrepnum", "-1")
	uidstr := c.DefaultQuery("uid", "-1")
	commentid, err_cid := strconv.Atoi(commentidstr)
	nowreply, err_now := strconv.Atoi(nowreplystr)
	uid, err_uid := strconv.Atoi(uidstr)
	postid, err_pid := strconv.Atoi(logidstr)
	if err_cid != nil || err_now != nil || err_uid != nil || err_pid != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	if commentid == -1 || nowreply == -1 || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	err, posts := GetCommentInfo(nowreply, postid, uid, commentid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	c.JSON(http.StatusOK, gin.H{"replies": posts})
}

// 删除评论接口
func DeleteComment(c *gin.Context) {
	var request struct {
		UID   string `json:"uid"`
		LogID string `json:"logid"`
		ComID string `json:"comid"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	uid, err := strconv.Atoi(request.UID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}

	commentID, err := strconv.Atoi(request.ComID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的评论ID"})
		return
	}
	postID, err := strconv.Atoi(request.LogID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的帖子ID"})
		return
	}

	erro, msg := model.DeleteCommentByUser(commentID, uid, postID)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}

// 删除回复接口
func DeleteReply(c *gin.Context) {
	var request struct {
		UID     string `json:"uid"`
		LogID   string `json:"logid"`
		ComID   string `json:"comid"`
		ReplyID string `json:"replyid"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}

	uid, err := strconv.Atoi(request.UID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}

	postID, err := strconv.Atoi(request.LogID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的帖子ID"})
		return
	}

	commentID, err := strconv.Atoi(request.ComID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的评论ID"})
		return
	}

	replyID, err := strconv.Atoi(request.ReplyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的回复ID"})
		return
	}

	erro, msg := model.DeleteReplyByUser(replyID, uid, postID, commentID)
	if erro != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": msg})
}
