package model

import (
	"middleproject/internal/repository"
	"strconv"
	"time"
)

type Comment struct {
	CommentID       int       `json:"comment_id"`
	CommenterID     int       `json:"commenter_id"`
	PostID          int       `json:"post_id"`
	ParentCommentID int       `json:"parent_comment_id"`
	CommentTime     time.Time `json:"comment_time"`
	Content         string    `json:"content"`
	LikeCount       int       `json:"like_count"`
	ReplyCount      int       `json:"reply_count"`
}

func (c *Comment) AddComment() (error, string, string) {
	db_link, err := repository.Connect()
	if err != nil {
		return err, "发评论连接数据库失败", "0"
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败", "0"
	}
	query := "INSERT INTO Comments (commenter_id,post_id, content) VALUES (?, ?, ?)"
	result, err := db.Exec(query, c.CommenterID, c.PostID, c.Content)
	if err != nil {
		db.Rollback()
		return err, "sql语句错误,评论创建失败", "0"
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		db.Rollback()
		return err, "获取新评论ID失败", "0"
	}

	query = "UPDATE Posts SET comment_count = comment_count + 1 WHERE post_id = ?"
	_, err = db.Exec(query, c.PostID)
	if err != nil {
		db.Rollback()
		return err, "更新帖子评论数量失败", "0"
	}
	c.CommentID = int(commentID)
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "事务提交失败", "0"
	}
	return nil, "评论创建成功", strconv.Itoa(int(commentID))
}

func AddReply(replyerID int, PostID int, commentID int, content string) (error, string, string) {
	db_link, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "发评论连接数据库失败", "0"
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败", "0"
	}
	query := "INSERT INTO comments (commenter_id, post_id, parent_comment_id, content) VALUES (?, ?, ?, ?)"
	result, err_in := db.Exec(query, replyerID, PostID, commentID, content)
	if err_in != nil {
		db.Rollback()
		return err_in, "sql语句错误,评论创建失败", "0"
	}
	replyID, err_id := result.LastInsertId()
	if err_id != nil {
		db.Rollback()
		return err_id, "获取新评论ID失败", "0"
	}
	query = "UPDATE comments SET reply_count = reply_count + 1 WHERE comment_id = ?"
	_, err_update := db.Exec(query, commentID)
	if err_update != nil {
		db.Rollback()
		return err_update, "更新评论回复数量失败", "0"
	}
	query = "UPDATE posts SET comment_count = comment_count + 1 WHERE post_id = ?"
	_, err_update = db.Exec(query, PostID)
	if err_update != nil {
		db.Rollback()
		return err_update, "更新帖子评论数量失败", "0"
	}

	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "事务提交失败", "0"
	}
	return nil, "评论创建成功", strconv.Itoa(int(replyID))
}
