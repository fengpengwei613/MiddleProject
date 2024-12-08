package service

import (
	"encoding/json"
	"fmt"
	"middleproject/internal/repository"
	"middleproject/scripts"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 查看个人发布的帖子
func GetPersonalPostLogs(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
	}
	defer db.Close()

	page := c.Query("page")
	uid := c.Query("uid")
	aimuid := c.Query("aimuid")

	if page == "" || uid == "" || aimuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求参数不能为空"})
		return
	}
	count, err := UserExists(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", uid)
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": failreason})
		return
	}
	count, err = UserExists(db, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", aimuid)
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": failreason})
		return
	}

	//查看页数是否无效
	pageint, err := strconv.Atoi(page)
	if err != nil || pageint < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的页数"})
		return
	}

	pagesize := 10
	startnumber := pageint * pagesize
	endnumber := (pageint + 1) * pagesize

	//查询帖子
	query := `SELECT p.post_id,p.title,p.user_id,u.uname,u.avatar,p.publish_time,LEFT(p.content,30) as somecontent,p.post_subject,p.friend_see 
	FROM Posts p
	JOIN Users u ON p.user_id=u.user_id
	WHERE p.user_id=?
	ORDER BY p.publish_time DESC`
	rows, err := db.Query(query, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子失败"})
		return
	}
	defer rows.Close()
	logs := []gin.H{}
	currentIndex := 0
	for rows.Next() {
		if (len(logs)) >= pagesize {
			break
		}
		var log struct {
			ID          int      `json:"id"`
			Title       string   `json:"title"`
			UID         string   `json:"uid"`
			Uname       string   `json:"uname"`
			Uimage      string   `json:"uimage"`
			Time        string   `json:"time"`
			SomeContent string   `json:"somecontent"`
			Subjects    []string `json:"subjects"`
			FriendSee   bool     `json:"friend_see"`
		}
		var subjectsJSON string
		if err := rows.Scan(&log.ID, &log.Title, &log.UID, &log.Uname, &log.Uimage, &log.Time, &log.SomeContent, &subjectsJSON, &log.FriendSee); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析失败"})
			return
		}
		var err_url error
		err_url, log.Uimage = scripts.GetUrl(log.Uimage)
		if err_url != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取头像Url失败"})
			return
		}

		//检查是否互关
		if log.FriendSee {
			friendCheckQuery := `SELECT COUNT(*)
			FROM userfollows f1
			JOIN userfollows f2 
			ON f1.follower_id=f2.followed_id
			AND f1.followed_id=f2.follower_id
			WHERE f1.follower_id=? AND f1.followed_id=?`
			var count int
			err := db.QueryRow(friendCheckQuery, uid, aimuid).Scan(&count)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询是否互关失败"})
				return
			}
			if count == 0 {
				continue
			}
		}

		if currentIndex >= startnumber && currentIndex < endnumber {
			var sujects []string
			if subjectsJSON != "" {
				if err := json.Unmarshal([]byte(subjectsJSON), &sujects); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析主题失败"})
					return
				}
			}
			log.Subjects = sujects
			logs = append(logs, gin.H{
				"id":          log.ID,
				"title":       log.Title,
				"uid":         log.UID,
				"uname":       log.Uname,
				"uimage":      log.Uimage,
				"time":        log.Time,
				"somecontent": log.SomeContent,
				"subjects":    log.Subjects,
			})
		}
		currentIndex++

	}
	countPostsQuery := "SELECT COUNT(*) FROM Posts WHERE user_id=?"
	var countPosts int
	if err := db.QueryRow(countPostsQuery, aimuid).Scan(&countPosts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子总数失败"})
	}
	tatalpages := (countPosts-1)/pagesize + 1
	c.JSON(http.StatusOK, gin.H{"isok": true, "logs": logs, "totalpages": tatalpages})

}

// 查看个人喜欢的帖子
func GetPersonalLikePosts(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
	}
	defer db.Close()

	page := c.Query("page")
	uid := c.Query("uid")
	aimuid := c.Query("aimuid")

	if page == "" || uid == "" || aimuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求参数不能为空"})
		return
	}
	count, err := UserExists(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", uid)
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": failreason})
		return
	}
	count, err = UserExists(db, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", aimuid)
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": failreason})
		return
	}

	//查看页数是否无效
	pageint, err := strconv.Atoi(page)
	if err != nil || pageint < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的页数"})
		return
	}

	pagesize := 10
	startnumber := pageint * pagesize
	endnumber := (pageint + 1) * pagesize

	//查看是否showlike
	query := "SELECT showlike FROM Users WHERE user_id=?"
	var showlike bool
	if err := db.QueryRow(query, aimuid).Scan(&showlike); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否显示喜欢失败"})
		return
	}
	if !showlike {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "用户不显示喜欢"})
		return
	}

	//查询帖子
	checkLikeQuery := `SELECT p.post_id,p.title,p.user_id,u.uname,u.avatar,p.publish_time,LEFT(p.content,30) as somecontent,p.post_subject,p.friend_see 
	FROM Postlikes pl
	JOIN Posts p ON pl.post_id=p.post_id
	JOIN Users u ON p.user_id=u.user_id
	WHERE pl.liker_id=?
	ORDER BY p.publish_time DESC`
	rows, err := db.Query(checkLikeQuery, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子失败"})
		return
	}
	defer rows.Close()
	logs := []gin.H{}
	currentIndex := 0
	for rows.Next() {
		if (len(logs)) >= pagesize {
			break
		}
		var log struct {
			ID          int      `json:"id"`
			Title       string   `json:"title"`
			UID         string   `json:"uid"`
			Uname       string   `json:"uname"`
			Uimage      string   `json:"uimage"`
			Time        string   `json:"time"`
			SomeContent string   `json:"somecontent"`
			Subjects    []string `json:"subjects"`
			FriendSee   bool     `json:"friend_see"`
		}
		var subjectsJSON string
		if err := rows.Scan(&log.ID, &log.Title, &log.UID, &log.Uname, &log.Uimage, &log.Time, &log.SomeContent, &subjectsJSON, &log.FriendSee); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析失败"})
			return
		}
		var err_url error
		err_url, log.Uimage = scripts.GetUrl(log.Uimage)
		if err_url != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取头像Url失败"})
			return
		}

		//检查是否互关
		if log.FriendSee {
			friendCheckQuery := `SELECT COUNT(*)
			FROM userfollows f1
			JOIN userfollows f2 
			ON f1.follower_id=f2.followed_id
			AND f1.followed_id=f2.follower_id
			WHERE f1.follower_id=? AND f1.followed_id=?`
			var count int
			err := db.QueryRow(friendCheckQuery, uid, log.UID).Scan(&count)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询是否互关失败"})
				return
			}
			if count == 0 {
				continue
			}
		}

		if currentIndex >= startnumber && currentIndex < endnumber {
			var sujects []string
			if subjectsJSON != "" {
				if err := json.Unmarshal([]byte(subjectsJSON), &sujects); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析主题失败"})
					return
				}
			}
			log.Subjects = sujects
			logs = append(logs, gin.H{
				"id":          log.ID,
				"title":       log.Title,
				"uid":         log.UID,
				"uname":       log.Uname,
				"uimage":      log.Uimage,
				"time":        log.Time,
				"somecontent": log.SomeContent,
				"subjects":    log.Subjects,
			})
		}
		currentIndex++

	}
	countPostsQuery := "SELECT COUNT(*) FROM Postlikes WHERE liker_id=?"
	var countPosts int
	if err := db.QueryRow(countPostsQuery, aimuid).Scan(&countPosts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子总数失败"})
	}
	tatalpages := countPosts/pagesize + 1
	c.JSON(http.StatusOK, gin.H{"isvaild": true, "logs": logs, "totalpages": tatalpages})
}

// 查看个人收藏帖子
func GetPersonalCollectPosts(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
	}
	defer db.Close()

	page := c.Query("page")
	uid := c.Query("uid")
	aimuid := c.Query("aimuid")

	if page == "" || uid == "" || aimuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求参数不能为空"})
		return
	}
	count, err := UserExists(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", uid)
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": failreason})
		return
	}
	count, err = UserExists(db, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", aimuid)
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": failreason})
		return
	}

	//查看页数是否无效
	pageint, err := strconv.Atoi(page)
	if err != nil || pageint < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的页数"})
		return
	}

	pagesize := 10
	startnumber := pageint * pagesize
	endnumber := (pageint + 1) * pagesize

	//查看是否showcollect
	query := "SELECT showcollect FROM Users WHERE user_id=?"
	var showcollect bool
	if err := db.QueryRow(query, aimuid).Scan(&showcollect); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询用户是否显示收藏失败"})
		return
	}
	if !showcollect {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "用户不显示收藏"})
		return
	}

	//查询帖子
	checkCollectQuery := `SELECT p.post_id,p.title,p.user_id,u.uname,u.avatar,p.publish_time,LEFT(p.content,30) as somecontent,p.post_subject,p.friend_see 
	FROM Postfavorites pf
	JOIN Posts p ON pf.post_id=p.post_id
	JOIN Users u ON p.user_id=u.user_id
	WHERE pf.user_id=?
	ORDER BY p.publish_time DESC`
	rows, err := db.Query(checkCollectQuery, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子失败"})
		return
	}
	defer rows.Close()
	logs := []gin.H{}
	currentIndex := 0
	for rows.Next() {
		if (len(logs)) >= pagesize {
			break
		}
		var log struct {
			ID          int      `json:"id"`
			Title       string   `json:"title"`
			UID         string   `json:"uid"`
			Uname       string   `json:"uname"`
			Uimage      string   `json:"uimage"`
			Time        string   `json:"time"`
			SomeContent string   `json:"somecontent"`
			Subjects    []string `json:"subjects"`
			FriendSee   bool     `json:"friend_see"`
		}
		var subjectsJSON string
		if err := rows.Scan(&log.ID, &log.Title, &log.UID, &log.Uname, &log.Uimage, &log.Time, &log.SomeContent, &subjectsJSON, &log.FriendSee); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析失败"})
			return
		}
		var err_url error
		err_url, log.Uimage = scripts.GetUrl(log.Uimage)
		if err_url != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取头像Url失败"})
			return
		}

		//检查是否互关
		if log.FriendSee {
			friendCheckQuery := `SELECT COUNT(*)
			FROM userfollows f1
			JOIN userfollows f2 
			ON f1.follower_id=f2.followed_id
			AND f1.followed_id=f2.follower_id
			WHERE f1.follower_id=? AND f1.followed_id=?`
			var count int
			err := db.QueryRow(friendCheckQuery, uid, log.UID).Scan(&count)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询是否互关失败"})
				return
			}
			if count == 0 {
				continue
			}
		}

		if currentIndex >= startnumber && currentIndex < endnumber {
			var sujects []string
			if subjectsJSON != "" {
				if err := json.Unmarshal([]byte(subjectsJSON), &sujects); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析主题失败"})
					return
				}
			}
			log.Subjects = sujects
			logs = append(logs, gin.H{
				"id":          log.ID,
				"title":       log.Title,
				"uid":         log.UID,
				"uname":       log.Uname,
				"uimage":      log.Uimage,
				"time":        log.Time,
				"somecontent": log.SomeContent,
				"subjects":    log.Subjects,
			})
		}
		currentIndex++

	}
	countPostsQuery := "SELECT COUNT(*) FROM Postfavorites WHERE user_id=?"
	var countPosts int
	if err := db.QueryRow(countPostsQuery, aimuid).Scan(&countPosts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子总数失败"})
	}
	tatalpages := countPosts/pagesize + 1
	c.JSON(http.StatusOK, gin.H{"isvaild": true, "logs": logs, "totalpages": tatalpages})

}
