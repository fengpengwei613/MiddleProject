package model

//package main

import (
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
	db ,err_tx := db_link.Begin()
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
	err_commit :=db.Commit()
	if err_commit != nil {
		db.Rollback()
		return err_commit, "提交事务失败", "0"
	}
	return nil, "帖子创建成功", strconv.Itoa(int(postID))

}
