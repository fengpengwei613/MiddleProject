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
	db, err := repository.Connect()
	if err != nil {
		return err, "发评论连接数据库失败", "0"
	}
	defer db.Close()

	query := "INSERT INTO Comments (commenter_id,post_id, content) VALUES (?, ?, ?)"
	result, err := db.Exec(query, c.CommenterID, c.PostID, c.Content)
	if err != nil {
		return err, "sql语句错误,评论创建失败", "0"
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return err, "获取新评论ID失败", "0"
	}
	c.CommentID = int(commentID)
	return nil, "评论创建成功", strconv.Itoa(int(commentID))
}
