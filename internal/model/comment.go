package model

import (
	"database/sql"
	"fmt"

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
	//检查用户是否被禁言/封号
	var userPermission int
	newquery := "SELECT user_id from usermutes where user_id = ? and end_time > now()"
	err = db.QueryRow(newquery, c.CommenterID).Scan(&userPermission)
	if err == nil {
		db.Rollback()
		return sql.ErrNoRows, "您已被禁言/封号", "0"
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
	//检查用户是否被禁言/封号
	var userPermission int
	newquery := "SELECT user_id from usermutes where user_id = ? and end_time > now()"
	err := db.QueryRow(newquery, replyerID).Scan(&userPermission)
	if err == nil {
		db.Rollback()
		return sql.ErrNoRows, "您已被禁言/封号", "0"
	}

	var top_parentid int
	newquery = "SELECT top_parentid FROM comments WHERE comment_id = ? and parent_comment_id is not null"
	err = db.QueryRow(newquery, commentID).Scan(&top_parentid)
	if err != nil {
		top_parentid = 0
	}
	query := "INSERT INTO comments (commenter_id, post_id, parent_comment_id, content,top_parentid) VALUES (?, ?, ?, ?, ?)"
	var result sql.Result
	var err_in error
	if top_parentid != 0 {
		result, err_in = db.Exec(query, replyerID, PostID, commentID, content, top_parentid)
	} else {
		result, err_in = db.Exec(query, replyerID, PostID, commentID, content, commentID)
		top_parentid = commentID
	}

	if err_in != nil {
		db.Rollback()
		fmt.Println(err_in.Error())
		return err_in, "sql语句错误,评论创建失败", "0"
	}
	replyID, err_id := result.LastInsertId()
	if err_id != nil {
		db.Rollback()
		return err_id, "获取新评论ID失败", "0"
	}
	query = "UPDATE comments SET reply_count = reply_count + 1 WHERE comment_id = ?"
	_, err_update := db.Exec(query, top_parentid)
	if err_update != nil {
		db.Rollback()
		return err_update, "更新评论回复数量失败", "0"
	}

	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "事务提交失败", "0"
	}
	return nil, "评论创建成功", strconv.Itoa(int(replyID))
}

// 删除评论
// 假设用户表中判断是否是管理员的属性是permission，permission=1时是管理员
func DeleteCommentByUser(commentID int, uid int, postID int) (error, string) {
	db_link, err := repository.Connect()
	if err != nil {
		return err, "连接数据库失败"
	}
	defer db_link.Close()

	var userPermission int
	query := "SELECT permission FROM users WHERE user_id = ?"
	err = db_link.QueryRow(query, uid).Scan(&userPermission)
	if err != nil {
		return err, "获取用户权限失败"
	}

	var commenterID int
	query = "SELECT commenter_id FROM Comments WHERE comment_id = ? AND post_id = ?"
	err = db_link.QueryRow(query, commentID, postID).Scan(&commenterID)
	if err != nil {
		return err, "获取评论信息失败"
	}

	if commenterID != uid && userPermission != 1 {
		return fmt.Errorf("无权限删除该评论"), "无权限删除该评论"
	}

	tx, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败"
	}

	err = deleteAllChildComments(tx, commentID)
	if err != nil {
		tx.Rollback()
		return err, "删除评论的回复失败"
	}
	query = "DELETE FROM Comments WHERE comment_id = ?"
	_, err = tx.Exec(query, commentID)
	if err != nil {
		tx.Rollback()
		return err, "删除评论失败"
	}

	query = "UPDATE Posts SET comment_count = comment_count - 1 WHERE post_id = ?"
	_, err = tx.Exec(query, postID)
	if err != nil {
		tx.Rollback()
		return err, "更新帖子评论数量失败"
	}

	err_commit := tx.Commit()
	if err_commit != nil {
		tx.Rollback()
		return err_commit, "事务提交失败"
	}

	return nil, "删除评论成功"
}

// 删除所有子评论
func deleteAllChildComments(tx *sql.Tx, parentCommentID int) error {
	query := "SELECT comment_id FROM Comments WHERE parent_comment_id = ?"
	rows, err := tx.Query(query, parentCommentID)
	if err != nil {
		return fmt.Errorf("查询子评论失败: %v", err)
	}
	defer rows.Close()

	var childCommentIDs []int
	for rows.Next() {
		var childCommentID int
		if err := rows.Scan(&childCommentID); err != nil {
			return fmt.Errorf("读取子评论ID失败: %v", err)
		}
		childCommentIDs = append(childCommentIDs, childCommentID)
	}

	if len(childCommentIDs) == 0 {
		return nil
	}
	for _, childCommentID := range childCommentIDs {
		err := deleteAllChildComments(tx, childCommentID)
		if err != nil {
			return err
		}

		query = "DELETE FROM Comments WHERE comment_id = ?"
		_, err = tx.Exec(query, childCommentID)
		if err != nil {
			return fmt.Errorf("删除子评论失败: %v", err)
		}
	}

	return nil
}

// 删除回复
func DeleteReplyByUser(replyID int, uid int, postID int, commentID int) (error, string) {
	db_link, err := repository.Connect()
	if err != nil {
		return err, "连接数据库失败"
	}
	defer db_link.Close()

	var userPermission int
	query := "SELECT permission FROM users WHERE user_id = ?"
	err = db_link.QueryRow(query, uid).Scan(&userPermission)
	if err != nil {
		return err, "获取用户权限失败"

	}
	var replierID int
	query = "SELECT commenter_id FROM comments WHERE comment_id = ? AND parent_comment_id = ? AND post_id = ?"
	err = db_link.QueryRow(query, replyID, commentID, postID).Scan(&replierID)
	if err != nil {
		return err, "获取回复信息失败"
	}
	if replierID != uid && userPermission != 1 {
		return fmt.Errorf("无权限删除该回复"), "无权限删除该回复"
	}

	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败"
	}

	query = "DELETE FROM comments WHERE comment_id = ?"
	_, err = db.Exec(query, replyID)
	if err != nil {
		db.Rollback()
		return err, "删除回复失败"
	}

	query = "UPDATE comments SET reply_count = reply_count - 1 WHERE comment_id = ?"
	_, err = db.Exec(query, commentID)
	if err != nil {
		db.Rollback()
		return err, "更新评论回复数量失败"
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "事务提交失败"
	}

	return nil, "删除回复成功"
}
