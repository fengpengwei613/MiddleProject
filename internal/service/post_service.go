package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"middleproject/scripts"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Advpost struct {
	PostID  int      `json:"id"`
	Title   string   `json:"title"`
	Uname   string   `json:"uname"`
	Uid     string   `json :"uid"`
	Uimge   string   `json:"uimage"`
	Time    string   `json:"time"`
	Subject []string `json:"subject"`
}

// 推荐逻辑设计
func AdvisePost(uid int, page int, isattention string) ([]Advpost, error, int) {
	db, err_conn := repository.Connect()
	if err_conn != nil {
		return nil, err_conn, 0
	}
	defer db.Close()
	var posts []Advpost
	if isattention == "true" {
		//获取关注的人的帖子，按喜欢数量排序
		query := "SELECT posts.post_id,posts.title,users.Uname,users.user_id,users.avatar,posts.publish_time,posts.post_subject from posts,users,userfollows where posts.user_id=users.user_id AND posts.user_id=userfollows.followed_id AND userfollows.follower_id=1 order by posts.view_count"
		rows, err_query := db.Query(query)
		if err_query != nil {
			fmt.Println(err_query.Error())
			return posts, err_query, 0
		}
		for rows.Next() {
			var post Advpost

			var subject sql.NullString
			var uidint int
			err_scan := rows.Scan(&post.PostID, &post.Title, &post.Uname, &uidint, &post.Uimge, &post.Time, &subject)
			if err_scan != nil {
				fmt.Println(err_scan.Error())
				return posts, err_scan, 0
			}
			post.Uid = strconv.Itoa(uidint)
			//post.Time = time.Format("2006-01-02 15:04:05")
			var err_url error
			err_url, post.Uimge = scripts.GetUrl(post.Uimge)
			if err_url != nil {
				return posts, err_url, 0
			}
			if subject.Valid {
				str := subject.String
				post.Subject = strings.Split(str[1:len(str)-1], ",")
				fmt.Println(post.Subject)

			} else {
				fmt.Println("111")
				fmt.Println(subject.String)

				post.Subject = []string{"123", "233"}
			}
			posts = append(posts, post)

		}

	} else {

	}
	return posts, nil, len(posts)

}

// 发帖子接口
func PublishPost(c *gin.Context) {
	var data model.Post
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "发帖绑定请求数据失败"})
		return
	}
	uidstr := c.DefaultQuery("uid", "-1")
	uid, err := strconv.Atoi(uidstr)
	if err != nil || uid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的用户ID"})
		return
	}
	data.UserID = uid
	erro, msg, idstr := data.AddPost()
	if erro != nil {
		fmt.Println(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": msg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "logid": idstr})
	//fmt.Println("返回的消息：", idstr)

}

// 获得推荐帖子
func GetRecommendPost(c *gin.Context) {
	var pagestr string = c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pagestr)

	var isattention string = c.DefaultQuery("isattion", "false")
	var uidstr = c.DefaultQuery("uid", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	var posts []Advpost
	if err_uid != nil {
		c.JSON(http.StatusBadRequest, gin.H{"logs": posts, "totalPages": -1})
		return
	}
	posts, err_adv, num := AdvisePost(uid, page, isattention)

	fmt.Println(posts)
	if err_adv != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": posts, "totalPages": num})
	return

}
