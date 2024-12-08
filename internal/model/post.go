package model

//package main

import (
	"database/sql"
	"encoding/json"
	"middleproject/internal/repository"
	"middleproject/scripts"
	"strconv"
	"time"
)

type Post struct {
	PostID        int       `json:"post_id"`
	UserID        int       `json:"uid"`
	PostTitle     string    `json:"title"`
	PostContent   string    `json:"content"`
	Images        []string  `json:"images"`
	PublishTime   time.Time `json:"publish_time"`
	CommentCount  int       `json:"comment_count"`
	ViewCount     int       `json:"view_count"`
	LikeCount     int       `json:"like_count"`
	FavoriteCount int       `json:"favorite_count"`
	Friend_See    bool      `json:"needfriend"`
	Subject       []string  `json:"subject"`
}

// 发帖功能
func (p *Post) AddPost() (error, string, string) {
	db_link, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "发帖函数连接数据库失败", "0"
	}
	defer db_link.Close()
	db, err_tx := db_link.Begin()
	if err_tx != nil {
		return err_tx, "事务开启失败", "0"
	}
	query_str := "INSERT INTO posts (user_id, title, content, images, friend_see, post_subject) " +
		"VALUES(?, ?, ?, ?, ?, ?)"
	var image_url = p.Images
	p.Images = []string{}
	//序列化
	
	jsonImages, err_json := json.Marshal(p.Images)
	if err_json != nil {
		db.Rollback()
		return err_json, "JSON 序列化失败", "0"
	}
	jsonSubject, err_json2 := json.Marshal(p.Subject)
	if err_json2 != nil {
		db.Rollback()
		return err_json2, "JSON 序列化失败", "0"
	}
	result, err_sql := db.Exec(query_str, p.UserID, p.PostTitle, p.PostContent, jsonImages, p.Friend_See, jsonSubject)
	if err_sql != nil {
		db.Rollback()
		return err_sql, "sql错误,帖子创建失败", "0"
	}
	postID, err := result.LastInsertId()
	if err != nil {
		db.Rollback()
		return err, "获取新帖ID失败", "0"
	}
	p.PostID = int(postID)
	postIDstr := strconv.Itoa(int(postID))
	realUrl := []string{}
	for idx, image := range image_url {
		// 上传图片到OSS
		filename := "image_" + postIDstr + "_" + strconv.Itoa(idx) + ".png"
		//objectKey在成功上传是文件路径，失败的话是错误信息
		err_up, objectKey := scripts.UploadImage(image, filename)
		if err_up != nil {
			db.Rollback()
			return err_up, objectKey, "0"
		}
		realUrl = append(realUrl, objectKey)
	}
	realUrlJson, err_json3 := json.Marshal(realUrl)
	if err_json3 != nil {
		db.Rollback()
		return err_json3, "JSON 序列化失败", "0"
	}
	//更新数据库
	update_str := "UPDATE posts SET images = ? WHERE post_id = ?"
	_, err_sql2 := db.Exec(update_str, realUrlJson, postID)
	if err_sql2 != nil {
		db.Rollback()
		return err_sql2, "sql错误,更新Url失败", "0"
	}
	err_commit := db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "提交事务失败", "0"
	}
	return nil, "帖子创建成功", strconv.Itoa(int(postID))

}

// 点赞帖子
func LikePost(userID, postID int) (error, string) {
	db_link, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "连接数据库失败"
	}
	defer db_link.Close()

	var count int
	query_check_like := "SELECT COUNT(*) FROM postlikes WHERE post_id = ? AND  liker_id = ?"
	err := db_link.QueryRow(query_check_like, postID, userID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err, "查询点赞记录失败"
	}

	// 取消点赞
	if count > 0 {
		_, err_del := db_link.Exec("DELETE FROM postlikes WHERE post_id = ? AND liker_id = ?", postID, userID)
		if err_del != nil {
			return err_del, "取消点赞失败"
		}
		_, err_update := db_link.Exec("UPDATE posts SET like_count = like_count - 1 WHERE post_id = ?", postID)
		if err_update != nil {
			return err_update, "更新点赞数失败"
		}
		return nil, "取消点赞成功"
	} else {
		_, err_add := db_link.Exec("INSERT INTO postlikes (post_id, liker_id) VALUES (?, ?)", postID, userID)
		if err_add != nil {
			return err_add, "点赞失败"
		}
		_, err_update := db_link.Exec("UPDATE posts SET like_count = like_count + 1 WHERE post_id = ?", postID)
		if err_update != nil {
			return err_update, "更新点赞数失败"
		}
		return nil, "点赞成功"
	}
}

// 收藏帖子
func CollectPost(userID, postID int) (error, string) {
	db_link, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "连接数据库失败"
	}
	defer db_link.Close()

	var count int
	query_check_favorite := "SELECT COUNT(*) FROM postfavorites WHERE post_id = ? AND user_id = ?"
	err := db_link.QueryRow(query_check_favorite, postID, userID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err, "查询收藏记录失败"
	}

	// 取消收藏
	if count > 0 {
		_, err_del := db_link.Exec("DELETE FROM postfavorites WHERE post_id = ? AND user_id = ?", postID, userID)
		if err_del != nil {
			return err_del, "取消收藏失败"
		}
		_, err_update := db_link.Exec("UPDATE posts SET favorite_count = favorite_count - 1 WHERE post_id = ?", postID)
		if err_update != nil {
			return err_update, "更新收藏数失败"
		}
		return nil, "取消收藏成功"
	} else {
		_, err_add := db_link.Exec("INSERT INTO postfavorites (post_id, user_id) VALUES (?, ?)", postID, userID)
		if err_add != nil {
			return err_add, "收藏失败"
		}
		_, err_update := db_link.Exec("UPDATE posts SET favorite_count = favorite_count + 1 WHERE post_id = ?", postID)
		if err_update != nil {
			return err_update, "更新收藏数失败"
		}
		return nil, "收藏成功"
	}
}

// 点赞评论
func LikeComment(userID, commentID int) (error, string) {
	db_link, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "连接数据库失败"
	}
	defer db_link.Close()

	var count int
	query_check_like := "SELECT COUNT(*) FROM commentlikes WHERE comment_id = ? AND liker_id = ?"
	err := db_link.QueryRow(query_check_like, commentID, userID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err, "查询点赞记录失败"
	}

	// 取消点赞
	if count > 0 {
		_, err_del := db_link.Exec("DELETE FROM commentlikes WHERE comment_id = ? AND liker_id = ?", commentID, userID)
		if err_del != nil {
			return err_del, "取消点赞失败"
		}
		_, err_update := db_link.Exec("UPDATE comments SET like_count = like_count - 1 WHERE comment_id = ?", commentID)
		if err_update != nil {
			return err_update, "更新评论点赞数失败"
		}
		return nil, "取消点赞成功"
	} else {
		_, err_add := db_link.Exec("INSERT INTO commentlikes (comment_id, liker_id) VALUES (?, ?)", commentID, userID)
		if err_add != nil {
			return err_add, "点赞失败"
		}
		_, err_update := db_link.Exec("UPDATE comments SET like_count = like_count + 1 WHERE comment_id = ?", commentID)
		if err_update != nil {
			return err_update, "更新评论点赞数失败"
		}
		return nil, "点赞成功"
	}
}
