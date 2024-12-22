package service

import (
	"database/sql"
	"fmt"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func SearchUser(c *gin.Context) {
	aimUidstr := c.DefaultQuery("aimuid", "-1")
	aimUname := c.DefaultQuery("aimuname", "")
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	uid, err_str2int := strconv.Atoi(aimUidstr)
	type Userinfo struct {
		Uid    string `json:"uid"`
		Uimage string `json: "uimage"`
		Uname  string `json: "uname"`
	}
	var users []Userinfo
	var totalPage int
	if err != nil || page == -1 || err_str2int != nil {
		c.JSON(http.StatusBadRequest, gin.H{"datas": users, "totalPages": 0})
	}
	if uid == -1 {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "SELECT user_id, Uname, avatar FROM users WHERE Uname like ? limit ?, 10"
		rows, err := db.Query(query, "%"+aimUname+"%", page*10)
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
		query := "SELECT count(*) FROM users WHERE Uname like ?"
		row := db.QueryRow(query, "%"+aimUname+"%")
		var temp int
		err = row.Scan(&temp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		totalPage = temp / 10
		if temp%10 != 0 {
			totalPage++
		}
	} else {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"datas": users, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "SELECT user_id, Uname, avatar FROM users WHERE user_id = ? limit ?, 10"
		rows, err := db.Query(query, uid, page*10)
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
		totalPage = 1
	}
	c.JSON(http.StatusOK, gin.H{"datas": users, "totalPages": totalPage})

}

func SearchPost(c *gin.Context) {
	aimPostidstr := c.DefaultQuery("aimlogid", "-1")
	aimPosttitle := c.DefaultQuery("aimtitle", "")
	pagestr := c.DefaultQuery("page", "-1")
	page, err := strconv.Atoi(pagestr)
	postid, err_str2int := strconv.Atoi(aimPostidstr)
	type Postinfo struct {
		Postid    string   `json:"id"`
		Posttitle string   `json:"title"`
		Uid       string   `json:"uid"`
		Uname     string   `json:"uname"`
		Uimage    string   `json:"uimage"`
		Time      string   `json:"time"`
		Subjects  []string `json:"subjects"`
	}
	var posts []Postinfo
	var totalPage int
	if err != nil || page == -1 || err_str2int != nil {
		c.JSON(http.StatusBadRequest, gin.H{"logs": posts, "totalPages": 0})
	}
	if postid == -1 {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "select post_id,posts.user_id,Uname,avatar,title,post_subject,publish_time from posts,users where posts.user_id = users.user_id and title like ? limit ?, 10"
		rows, err := db.Query(query, "%"+aimPosttitle+"%", page*10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		for rows.Next() {
			var post Postinfo
			var subjects sql.NullString
			err = rows.Scan(&post.Postid, &post.Uid, &post.Uname, &post.Uimage, &post.Posttitle, &subjects, &post.Time)
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
			post.Time = post.Time[:len(post.Time)-3]
			err, post.Uimage = scripts.GetUrl(post.Uimage)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
				return
			}
			posts = append(posts, post)
		}
		query := "SELECT count(*) FROM posts WHERE title like ?"
		row := db.QueryRow(query, "%"+aimPosttitle+"%")
		var temp int
		err = row.Scan(&temp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		totalPage = temp / 10
		if temp%10 != 0 {
			totalPage++
		}
	} else {
		db, err := repository.Connect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		defer db.Close()
		query := "select post_id,posts.user_id,Uname,avatar,title,post_subject,publish_time from posts,users where posts.user_id = users.user_id and post_id = ?"
		rows, err := db.Query(query, postid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
			return
		}
		for rows.Next() {
			var post Postinfo
			var subjects sql.NullString
			err = rows.Scan(&post.Postid, &post.Uid, &post.Uname, &post.Uimage, &post.Posttitle, &subjects, &post.Time)
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
			post.Time = post.Time[:len(post.Time)-3]
			err, post.Uimage = scripts.GetUrl(post.Uimage)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"logs": posts, "totalPages": 0})
				return
			}
			posts = append(posts, post)
		}
		totalPage = 1
	}
	c.JSON(http.StatusOK, gin.H{"logs": posts, "totalPages": totalPage})

}

// 生成封禁禁言警告系统消息(封禁、禁言、警告)，参数：类型，被举报类型（帖子、评论）、类型id(帖子id、评论id)，用户id，天数
func MakeSysinfo(Htype string, rtype string, id int, day int) (bool, string) {
	db, err := repository.Connect()
	var content string
	if rtype == "log" {
		select_query := "SELECT title FROM posts WHERE post_id = ?"
		row := db.QueryRow(select_query, id)
		err := row.Scan(&content)
		if err != nil {
			return false, "查询帖子失败"
		}
		content = "您的帖子《" + content + "》违反社区规则，已被管理员删除。"

	} else if rtype == "comment" || rtype == "reply" {
		select_query := "SELECT content FROM comments WHERE comment_id = ?"
		row := db.QueryRow(select_query, id)
		err := row.Scan(&content)
		if err != nil {
			return false, "查询评论失败"
		}
		content = "您的评论《" + content + "》违反社区规则，已被管理员删除。"
	}
	var info string
	currentTime := time.Now()
	chinaTime := currentTime.Add(8 * time.Hour)
	if Htype == "封禁" {
		start := chinaTime
		end := start.Add(time.Hour * 24 * day)
		var startstr string
		var endstr string
		startstr = start.Format("2006-01-02 15:04:05")
		endstr = end.Format("2006-01-02 15:04:05")
		info = "我们遗憾地通知您，由于您在本网站的行为违反了我们的社区规范，您的账户已被暂时封禁(" + startstr + "-" + endstr + ")。具体原因如下：\n  "
	} else if Htype == "禁言" {
		start := chinaTime
		end := start.Add(time.Hour * 24 * day)
		var startstr string
		var endstr string
		startstr = start.Format("2006-01-02 15:04:05")
		endstr = end.Format("2006-01-02 15:04:05")
		info = "我们遗憾地通知您，由于您在本网站的行为违反了我们的社区规范，您的账户已被暂时禁言(" + startstr + "-" + endstr + ")。具体原因如下：\n  "
	} else if Htype == "警告" {
		info = "我们遗憾地通知您，您在本网站的行为违反了我们的社区规范，我们已对您的行为进行了警告。具体原因如下：\n  "
	}
	info = info + content + "\n"
	info = info + "我们重视每一位用户的体验，并致力于维护一个健康、积极的社区环境。请您在未来遵守以下社区行为准则：\n"
	info = info + "1、尊重他人，保持友善的交流。\n2、禁止发布任何违反法律法规的内容。\n3、禁止发布任何侮辱、攻击、歧视性的言论。\n"
	return true, info
}

// 用户反馈(返回要存储在数据库的信息)，被处理人类型，天数和id
func UserFeedback(Htype string, day int, uid int) (bool, string) {
	db, err := repository.Connect()
	if err != nil {
		return false, "数据库连接失败"
	}
	defer db.Close()
	var uname string
	select_query := "SELECT Uname FROM users WHERE user_id = ?"
	row := db.QueryRow(select_query, uid)
	err = row.Scan(&uname)
	if err != nil {
		return false, "查询用户失败"
	}
	var infor string
	var content string
	infor = "尊敬的用户，您好！您向我们提出的反馈我们已经处理，处理结果如下：\n  "
	if Htype == "封禁" {
		content = uname + "发布的内容因为违反社区规则，已被封禁" + day + "天。"
	} else if Htype == "禁言" {
		content = uname + "发布的内容因为违反社区规则，已被禁言" + day + "天。"
	} else if Htype == "警告" {
		content = uname + "发布的内容因为违反社区规则，已被警告。"
	}
	infor = infor + content + "\n"
	infor = infor + "  感谢您对净化社区环境的贡献，我们将继续努力，为您提供更好的服务！"
	return true, infor
}
