package service

import (
	"database/sql"
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
	fmt.Println("查看个人发布的贴子")
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
	if err != nil || pageint < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的页数"})
		return
	}
	pageint -= 1

	pagesize := 10
	startnumber := pageint * pagesize

	//查询帖子
	query := `
	(
		-- 非互关情况，直接查询
		SELECT p.post_id, p.title, p.user_id, u.uname, u.avatar, p.publish_time,
			LEFT(p.content, 30) AS somecontent, p.post_subject, p.friend_see
		FROM Posts p
		JOIN Users u ON p.user_id = u.user_id
		WHERE p.user_id = ? AND p.friend_see = FALSE
	)
	UNION
	(
		-- 互关情况，通过联合查询确保只有互关的情况下才显示帖子
		SELECT p.post_id, p.title, p.user_id, u.uname, u.avatar, p.publish_time,
			LEFT(p.content, 30) AS somecontent, p.post_subject, p.friend_see
		FROM Posts p
		JOIN Users u ON p.user_id = u.user_id
		JOIN userfollows f1 ON f1.follower_id = ? AND f1.followed_id = p.user_id
		JOIN userfollows f2 ON f2.follower_id = p.user_id AND f2.followed_id = ?
		WHERE p.user_id = ? AND p.friend_see = TRUE
	)
	ORDER BY publish_time DESC
	LIMIT ?, ?`
	rows, err := db.Query(query, aimuid, uid, uid, aimuid, startnumber, pagesize)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"isvalid": true, "logs": []gin.H{}, "totalPages": 0})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子失败"})
			return
		}
	}
	defer rows.Close()
	logs := []gin.H{}
	for rows.Next() {
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

		// 获取头像URL
		err, log.Uimage = scripts.GetUrl(log.Uimage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取头像Url失败"})
			return
		}

		var subjects []string
		if subjectsJSON != "" {
			if err := json.Unmarshal([]byte(subjectsJSON), &subjects); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析主题失败"})
				return
			}
		}
		log.Subjects = subjects
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

	// 获取帖子总数
	countPostsQuery := `SELECT COUNT(*)
	FROM Posts p 
	JOIN Users u ON p.user_id = u.user_id
	LEFT JOIN userfollows f1 ON f1.follower_id = ? AND f1.followed_id = p.user_id
	LEFT JOIN userfollows f2 ON f2.follower_id = p.user_id AND f2.followed_id = ?
	WHERE p.user_id = ?
	AND (
		(p.friend_see = FALSE) OR
		(p.friend_see = TRUE AND f1.follower_id IS NOT NULL AND f2.follower_id IS NOT NULL)
	)`
	var countPosts int
	if err := db.QueryRow(countPostsQuery, uid, uid, aimuid).Scan(&countPosts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子总数失败"})
		return
	}

	totalPages := (countPosts-1)/pagesize + 1
	c.JSON(http.StatusOK, gin.H{"isok": true, "logs": logs, "totalPages": totalPages})

}

// 查看个人喜欢的帖子
func GetPersonalLikePosts(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "数据库连接失败"})
	}
	defer db.Close()

	page := c.Query("page")
	uid := c.Query("uid")
	aimuid := c.Query("aimuid")

	if page == "" || uid == "" || aimuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "请求参数不能为空"})
		return
	}
	count, err := UserExists(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", uid)
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": failreason})
		return
	}
	count, err = UserExists(db, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", aimuid)
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": failreason})
		return
	}

	//查看页数是否无效
	pageint, err := strconv.Atoi(page)
	if err != nil || pageint < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "无效的页数"})
		return
	}
	pageint -= 1

	pagesize := 10
	startnumber := pageint * pagesize

	//查看是否showlike
	query := "SELECT showlike FROM Users WHERE user_id=?"
	var showlike bool
	if err := db.QueryRow(query, aimuid).Scan(&showlike); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询用户是否显示喜欢失败"})
		return
	}
	if !showlike && uid != aimuid {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "用户不显示喜欢"})
		return
	}

	//查询帖子
	query = `
		(
			-- 非互关的帖子：friend_see 为 false
			SELECT p.post_id, p.title, p.user_id, u.uname, u.avatar, p.publish_time,
				LEFT(p.content, 30) AS somecontent, p.post_subject, p.friend_see
			FROM Postlikes pl
			JOIN Posts p ON pl.post_id = p.post_id
			JOIN Users u ON p.user_id = u.user_id
			WHERE pl.liker_id = ? AND p.friend_see = FALSE
		)
		UNION
		(
			-- 互关的帖子：friend_see 为 true 且用户之间互相关注
			SELECT p.post_id, p.title, p.user_id, u.uname, u.avatar, p.publish_time,
				LEFT(p.content, 30) AS somecontent, p.post_subject, p.friend_see
			FROM Postlikes pl
			JOIN Posts p ON pl.post_id = p.post_id
			JOIN Users u ON p.user_id = u.user_id
			JOIN userfollows f1 ON f1.follower_id = ? AND f1.followed_id = p.user_id
			JOIN userfollows f2 ON f2.follower_id = p.user_id AND f2.followed_id = ?
			WHERE pl.liker_id = ? AND p.friend_see = TRUE
		)
		ORDER BY publish_time DESC
		LIMIT ?, ?`
	rows, err := db.Query(query, aimuid, uid, uid, aimuid, startnumber, pagesize)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"isvalid": true, "logs": []gin.H{}, "totalPages": 0})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询帖子失败"})
			return
		}
	}
	defer rows.Close()
	logs := []gin.H{}
	for rows.Next() {
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
			c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "解析失败"})
			return
		}
		var err_url error
		err_url, log.Uimage = scripts.GetUrl(log.Uimage)
		if err_url != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "获取头像Url失败"})
			return
		}

		var sujects []string
		if subjectsJSON != "" {
			if err := json.Unmarshal([]byte(subjectsJSON), &sujects); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "解析主题失败"})
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

	countPostsQuery := `SELECT COUNT(*)
	FROM Postlikes pl
	JOIN Posts p ON pl.post_id = p.post_id
	JOIN Users u ON p.user_id = u.user_id
	LEFT JOIN userfollows f1 ON f1.follower_id = ? AND f1.followed_id = p.user_id
	LEFT JOIN userfollows f2 ON f2.follower_id = p.user_id AND f2.followed_id = ?
	WHERE pl.liker_id = ?
	AND (
		(p.friend_see = FALSE) OR
		(p.friend_see = TRUE AND f1.follower_id IS NOT NULL AND f2.follower_id IS NOT NULL)
	)`
	var countPosts int
	if err := db.QueryRow(countPostsQuery, uid, uid, aimuid).Scan(&countPosts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询帖子总数失败"})
	}
	tatalpages := (countPosts-1)/pagesize + 1
	c.JSON(http.StatusOK, gin.H{"isvalid": true, "logs": logs, "totalPages": tatalpages})
}

// 查看个人收藏帖子
func GetPersonalCollectPosts(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "数据库连接失败"})
	}
	defer db.Close()

	page := c.Query("page")
	uid := c.Query("uid")
	aimuid := c.Query("aimuid")

	if page == "" || uid == "" || aimuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "请求参数不能为空"})
		return
	}
	count, err := UserExists(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", uid)
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": failreason})
		return
	}
	count, err = UserExists(db, aimuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询用户是否存在失败"})
		return
	}
	if !count {
		failreason := fmt.Sprintf("用户%s不存在", aimuid)
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": failreason})
		return
	}

	//查看页数是否无效
	pageint, err := strconv.Atoi(page)
	if err != nil || pageint < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "无效的页数"})
		return
	}
	pageint -= 1

	pagesize := 10
	startnumber := pageint * pagesize

	//查看是否showcollect
	query := "SELECT showcollect FROM Users WHERE user_id=?"
	var showcollect bool
	if err := db.QueryRow(query, aimuid).Scan(&showcollect); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询用户是否显示收藏失败"})
		return
	}
	if !showcollect&& uid!=aimuid {
		c.JSON(http.StatusBadRequest, gin.H{"isvalid": false, "failreason": "用户不显示收藏"})
		return
	}

	//查询帖子
	checkCollectQuery := `
	(SELECT p.post_id,p.title,p.user_id,u.uname,u.avatar,p.publish_time,LEFT(p.content,30) as somecontent,p.post_subject,p.friend_see 
	FROM Postfavorites pf
	JOIN Posts p ON pf.post_id=p.post_id
	JOIN Users u ON p.user_id=u.user_id
	WHERE pf.user_id=? and p.friend_see=false)
   UNION(
	SELECT p.post_id,p.title,p.user_id,u.uname,u.avatar,p.publish_time,LEFT(p.content,30) as somecontent,p.post_subject,p.friend_see
	FROM Postfavorites pf
	JOIN Posts p ON pf.post_id=p.post_id
	JOIN Users u ON p.user_id=u.user_id
	JOIN userfollows f1 ON f1.follower_id=? AND f1.followed_id=p.user_id
	JOIN userfollows f2 ON f2.follower_id=p.user_id AND f2.followed_id=?
	WHERE pf.user_id=? and p.friend_see=true
   )
	ORDER BY publish_time DESC
	LIMIT ?, ?`
	rows, err := db.Query(checkCollectQuery, aimuid, uid, uid, aimuid, startnumber, pagesize)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"isvalid": true, "logs": []gin.H{}, "totalPages": 0})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询帖子失败"})
			return
		}
	}
	defer rows.Close()
	logs := []gin.H{}
	for rows.Next() {
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
			c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "解析失败"})
			return
		}
		var err_url error
		err_url, log.Uimage = scripts.GetUrl(log.Uimage)
		if err_url != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "获取头像Url失败"})
			return
		}

		var sujects []string
		if subjectsJSON != "" {
			if err := json.Unmarshal([]byte(subjectsJSON), &sujects); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "解析主题失败"})
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

	countPostsQuery := `SELECT COUNT(*)
	FROM Postfavorites pf
	JOIN Posts p ON pf.post_id = p.post_id
	JOIN Users u ON p.user_id = u.user_id
	LEFT JOIN userfollows f1 ON f1.follower_id = ? AND f1.followed_id = p.user_id
	LEFT JOIN userfollows f2 ON f2.follower_id = p.user_id AND f2.followed_id = ?
	WHERE pf.user_id = ?
	AND (
		(p.friend_see = FALSE) OR
		(p.friend_see = TRUE AND f1.follower_id IS NOT NULL AND f2.follower_id IS NOT NULL)
	)`
	var countPosts int
	if err := db.QueryRow(countPostsQuery, uid, uid, aimuid).Scan(&countPosts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isvalid": false, "failreason": "查询帖子总数失败"})
	}
	tatalpages := (countPosts-1)/pagesize + 1
	c.JSON(http.StatusOK, gin.H{"isvalid": true, "logs": logs, "totalPages": tatalpages})

}
