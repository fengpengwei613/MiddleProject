package service

import (
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
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
	uid, err := strconv.Atoi(c.Query("uid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}

	commentID, err := strconv.Atoi(c.Query("comid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的评论ID"})
		return
	}

	logID, err := strconv.Atoi(c.Query("logid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的帖子ID"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE comment_id = ? AND post_id = ?)", commentID, logID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "检查回复是否存在时发生错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "回复不存在或与帖子ID不对应"})
		return
	}

	var userPermission int
	err = db.QueryRow("SELECT peimission FROM users WHERE user_id = ?", uid).Scan(&userPermission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取用户权限失败"})
		return
	}

	var commenterID int
	err = db.QueryRow("SELECT commenter_id FROM comments WHERE comment_id = ?", commentID).Scan(&commenterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取评论者ID失败"})
		return
	}

	if commenterID != uid && userPermission != 1 {
		c.JSON(http.StatusForbidden, gin.H{"isok": false, "failreason": "无权限删除该评论"})
		return
	}

	_, err = db.Exec("DELETE FROM comments WHERE comment_id = ?", commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "删除评论失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": "删除评论成功"})
}

// 删除回复接口
func DeleteReply(c *gin.Context) {
	uid, err := strconv.Atoi(c.Query("uid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}

	commentID, err := strconv.Atoi(c.Query("comid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的评论ID"})
		return
	}

	replyID, err := strconv.Atoi(c.Query("replyid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的回复ID"})
		return
	}

	logID, err := strconv.Atoi(c.Query("logid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的帖子ID"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE comment_id = ? AND post_id = ?)", replyID, logID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "检查回复是否存在时发生错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "回复不存在或与帖子ID不对应"})
		return
	}

	var userPermission int
	err = db.QueryRow("SELECT peimission FROM users WHERE user_id = ?", uid).Scan(&userPermission)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取用户权限失败"})
		return
	}

	var replierID int
	err = db.QueryRow("SELECT commenter_id FROM comments WHERE comment_id = ?", replyID).Scan(&replierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取回复者ID失败"})
		return
	}

	if replierID != uid && userPermission != 1 {
		c.JSON(http.StatusForbidden, gin.H{"isok": false, "failreason": "无权限删除该回复"})
		return
	}

	_, err = db.Exec("DELETE FROM comments WHERE comment_id = ?", replyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "删除回复失败"})
		return
	}

	// _, err = db.Exec("UPDATE comments SET reply_count = reply_count - 1 WHERE comment_id = ?", commentID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "更新评论回复数量失败"})
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": "删除回复成功"})
}
