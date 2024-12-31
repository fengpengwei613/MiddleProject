package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"middleproject/scripts"
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
	if erro == sql.ErrNoRows {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": msg})
		return
	}
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
	nowcommentstr := c.DefaultQuery("nowcomnum", "-1")
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

// 帖子内部评论结构体
type CommentReply struct {
	CID     string `json:"id"`
	UID     string `json:"uid"`
	Content string `json:"content"`
	UName   string `json:"uname"`
	UImage  string `json:"uimage"`
	Time    string `json:"time"`
	IsLike  bool   `json:"islike"`
	Likenum int    `json:"likenum"`
	Touid   string `json:"touid"`
	Touname string `json:"touname"`
}

func GetCommentReply(page_num int, requester_uid int, comid int) ([]CommentReply, error) {
	var replys []CommentReply
	db, err := repository.Connect()
	if err != nil {
		return replys, err
	}
	defer db.Close()

	//查询评论的回复
	querystr := "SELECT com1.comment_id,com1.commenter_id,com1.content,com1.comment_time,com1.like_count,u1.Uname AS commenter_name_1,u1.avatar AS commenter_avatar_1, u2.Uname AS commenter_name_2,u2.user_id AS commenter_id_2 FROM comments com1 JOIN comments com2 ON com2.comment_id = com1.parent_comment_id JOIN users u1 ON com1.commenter_id = u1.user_id JOIN users u2 ON com2.commenter_id = u2.user_id WHERE com1.top_parentid = ?"
	querystr += " ORDER BY com1.comment_time DESC LIMIT ?, 5"
	rows, err := db.Query(querystr, comid, page_num)
	if err != nil {
		fmt.Println(err.Error())
		return replys, err
	}
	for rows.Next() {
		var reply CommentReply
		err = rows.Scan(&reply.CID, &reply.UID, &reply.Content, &reply.Time, &reply.Likenum, &reply.UName, &reply.UImage, &reply.Touid, &reply.Touname)
		if err != nil {
			fmt.Println(err.Error())
			return replys, err
		}
		//查询是否点赞
		querystr = "SELECT EXISTS(SELECT 1 FROM commentlikes WHERE comment_id = ? AND liker_id = ?)"
		var exists bool
		err = db.QueryRow(querystr, reply.CID, requester_uid).Scan(&exists)
		if err != nil {
			fmt.Println(err.Error())
			return replys, err
		}
		reply.IsLike = exists
		//Geturl
		var err_url error
		err_url, reply.UImage = scripts.GetUrl(reply.UImage)
		if err_url != nil {
			return replys, err_url

		}

		replys = append(replys, reply)
	}
	return replys, nil

}

func GetMoreReply(c *gin.Context) {
	commentidstr := c.DefaultQuery("comid", "-1")
	logidstr := c.DefaultQuery("logid", "-1")
	nowreplystr := c.DefaultQuery("nowrepnum", "-1")
	uidstr := c.DefaultQuery("uid", "-1")
	commentid, err_cid := strconv.Atoi(commentidstr)
	nowreply, err_now := strconv.Atoi(nowreplystr)
	uid, err_uid := strconv.Atoi(uidstr)
	if err_uid != nil {
		uid = -1
	}
	postid, err_pid := strconv.Atoi(logidstr)
	if err_cid != nil || err_now != nil || err_pid != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	if commentid == -1 || nowreply == -1 || postid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	replys, err := GetCommentReply(nowreply, uid, commentid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"replies": replys})
		return
	}
	c.JSON(http.StatusOK, gin.H{"replies": replys})
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
	//查询评论的帖子id
	querystr := "SELECT post_id FROM comments WHERE comment_id = ?"
	var post_id int
	err = db.QueryRow(querystr, commentID).Scan(&post_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取评论的帖子id失败"})
		return
	}

	_, err = db.Exec("DELETE FROM comments WHERE comment_id = ?", commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "删除评论失败"})
		return
	}

	_, err = db.Exec("UPDATE posts SET comment_count = comment_count - 1 WHERE post_id = ?", post_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "更新帖子评论数量失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true})
}

// 删除回复接口
func DeleteReply(c *gin.Context) {
	uid, err := strconv.Atoi(c.Query("uid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}

	_, err = strconv.Atoi(c.Query("comid"))
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
	//查询评论的top_parentid
	querystr := "SELECT top_parentid FROM comments WHERE comment_id = ?"
	var top_parent_id int
	err = db.QueryRow(querystr, replyID).Scan(&top_parent_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取回复的top_parent_id失败"})
		return
	}

	_, err = db.Exec("DELETE FROM comments WHERE comment_id = ?", replyID)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "删除回复失败"})
		return
	}

	_, err = db.Exec("UPDATE comments SET reply_count = reply_count - 1 WHERE comment_id = ?", top_parent_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "更新评论回复数量失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": "删除回复成功"})
}
