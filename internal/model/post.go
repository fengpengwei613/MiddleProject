package model

//package main

import (
	"encoding/json"
	"fmt"
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
	db, err_conn := repository.Connect()
	if err_conn != nil {
		return err_conn, "发帖函数连接数据库失败", "0"
	}
	defer db.Close()

	query_str := "INSERT INTO posts (user_id, title, content, images, friend_see, post_subject) " +
		"VALUES(?, ?, ?, ?, ?, ?)"
	//将图片存到云盘
	var image_url []string
	for idx, image := range p.Images {
		// 上传图片到OSS
		filename := "image_" + strconv.Itoa(p.UserID) + "_" + strconv.Itoa(idx) + ".png"
		//objectKey在成功上传是文件路径，失败的话是错误信息
		err_up, objectKey := scripts.UploadImage(image, filename)
		if err_up != nil {
			return err_up, objectKey, "0"
		}
		image_url = append(image_url, objectKey)
	}
	p.Images = image_url
	//将切片序列化
	jsonImages, err_json := json.Marshal(p.Images)
	if err_json != nil {
		fmt.Println("JSON 序列化失败:", err_json)
		return err_json, "JSON 序列化失败", "0"
	}
	jsonSubject, err_json2 := json.Marshal(p.Subject)
	if err_json2 != nil {
		fmt.Println("JSON 序列化失败:", err_json)
		return err_json2, "JSON 序列化失败", "0"
	}
	result, err_sql := db.Exec(query_str, p.UserID, p.PostTitle, p.PostContent, jsonImages, p.Friend_See, jsonSubject)
	if err_sql != nil {
		return err_sql, "sql错误,帖子创建失败", "0"
	} else {
		postID, err := result.LastInsertId()
		if err != nil {
			return err, "获取新帖ID失败", "0"
		}
		p.PostID = int(postID)
		return nil, "帖子创建成功", strconv.Itoa(int(postID))
	}
}
