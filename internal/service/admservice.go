package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"middleproject/scripts"

	_ "github.com/go-sql-driver/mysql"
)

func AdmLogin(c *gin.Context) {
	var requestData model.LoginRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	var storedPassword string
	var userID string
	var userName string
	var Avatar string
	var peimission int
	isEmail := isEmailFormat(requestData.Userid)
	var query string
	if isEmail {
		query = "SELECT user_id, password, Uname, avatar, peimission FROM users WHERE email = ?"
	} else {
		query = "SELECT user_id, password, Uname, avatar, peimission FROM users WHERE user_id = ?"
	}
	row := db.QueryRow(query, requestData.Userid)
	info := row.Scan(&userID, &storedPassword, &userName, &Avatar, &peimission)

	if info != nil {
		if info == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"isok": false, "failreason": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库查询失败"})
		return
	}
	if storedPassword != requestData.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "密码错误"})
		return
	}
	if peimission != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "您不是管理员，请使用客户端登录"})
		return
	}
	err, Avatar = scripts.GetUrl(Avatar)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": Avatar})
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "uid": userID, "uname": userName, "uimage": Avatar})
}

type Userinfo struct {
	Uid    string `json:"uid"`
	Uimage string `json: "uimage"`
	Uname  string `json: "uname"`
}

func GetallUser(c *gin.Context) {
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	var users []Userinfo
	if err != nil || page == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"datas": users, "totalPages": 0})
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
		return
	}
	defer db.Close()
	query := "SELECT user_id, Uname, avatar FROM users limit ?, 10"
	rows, err := db.Query(query, page*10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
		return
	}
	for rows.Next() {
		var user Userinfo
		err = rows.Scan(&user.Uid, &user.Uname, &user.Uimage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		err, user.Uimage = scripts.GetUrl(user.Uimage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		users = append(users, user)
	}
	query = "SELECT count(*) FROM users"
	row := db.QueryRow(query)
	var total int
	err = row.Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
		return
	}
	totalPages := total / 10
	if total%10 != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, gin.H{"datas": users, "totalPages": totalPages})
}

type Postinfo struct {
	Postid      string   `json:"id"`
	Posttitle   string   `json:"title"`
	Uid         string   `json:"uid"`
	Uname       string   `json:"uname"`
	Uimage      string   `json:"uimage"`
	Time        string   `json:"time"`
	Somecontent string   `json:"content"`
	Subjects    []string `json:"subjects"`
}

func GetallPost(c *gin.Context) {
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	var posts []Postinfo
	if err != nil || page == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"logs": posts, "totalPages": 0})
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	defer db.Close()
	query := "select post_id,posts.user_id,Uname,avatar,title,content,post_subject,publish_time from posts,users where posts.user_id = users.user_id limit ?, 10"
	rows, err := db.Query(query, page*10)
	fmt.Println("1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	for rows.Next() {
		var post Postinfo
		var subjects sql.NullString
		err = rows.Scan(&post.Postid, &post.Uid, &post.Uname, &post.Uimage, &post.Posttitle, &post.Somecontent, &subjects, &post.Time)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		if subjects.Valid {
			str := subjects.String
			post.Subjects = strings.Split(str[1:len(str)-1], ",")
			//去除双引号
			for i := 0; i < len(post.Subjects); i++ {
				if i == 0 {
					post.Subjects[i] = "#" + post.Subjects[i][1:len(post.Subjects[i])-1]

				} else {
					post.Subjects[i] = "#" + post.Subjects[i][2:len(post.Subjects[i])-1]
				}
			}

		}
		if len(post.Somecontent) > 300 {
			post.Somecontent = post.Somecontent[:300] + "..."
		}
		post.Time = post.Time[:len(post.Time)-3]
		err, post.Uimage = scripts.GetUrl(post.Uimage)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		posts = append(posts, post)
	}
	query = "SELECT count(*) FROM posts"
	row := db.QueryRow(query)
	var total int
	err = row.Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	totalPages := total / 10
	if total%10 != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, gin.H{"logs": posts, "totalPages": totalPages})

}
