package service

import (
	"fmt"
	"middleproject/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)



// 获取举报目标的接口
func GetReports(c *gin.Context) {
    db, err := repository.Connect()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
        return
    }
    defer db.Close()

    // 获取请求参数
    page := c.DefaultQuery("page", "1")
    uid := c.Query("uid")
    if uid == "" {
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "uid不能为空"})
        return
    }

    // 将 page 转换为整数
    pageInt, err := strconv.Atoi(page)
    if err != nil || pageInt < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的页数"})
        return
    }

    pageSize := 10// 每页 10 条数据
    startNumber := pageInt * pageSize

    // 查询用户是否存在
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM Users WHERE user_id = ?", uid).Scan(&count)
    if err != nil || count == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": fmt.Sprintf("用户 %s 不存在", uid)})
        return
    }

    // 获取该用户举报的帖子、评论、回复等数据
    query := `(
    SELECT
        r.reporter_id, 
        'log' AS type,  -- 帖子举报
        r.post_id, 
        -1 AS comment_id, 
        -1 AS reply_id, 
        u.user_id AS uid, 
        u.Uname AS uname, 
        r.reason AS reason,
        r.report_time
    FROM PostReports r
    JOIN Posts p ON p.post_id = r.post_id
    JOIN Users u ON u.user_id = p.user_id
    WHERE r.reporter_id = ?
)
UNION
(
    SELECT 
        r.reporter_id, 
        'comment' AS type,  -- 评论举报
        c.post_id AS post_id, 
        r.comment_id, 
        -1 AS reply_id, 
        u.user_id AS uid, 
        u.Uname AS uname, 
        r.reason AS reason,
        r.report_time
    FROM CommentReports r
    JOIN Comments c ON c.comment_id = r.comment_id
    JOIN Users u ON u.user_id = c.commenter_id

    WHERE r.reporter_id = ?
    AND c.parent_comment_id IS NULL
)
UNION
(
    SELECT 
        r.reporter_id, 
        'reply' AS type,  -- 回复举报
        c.post_id AS post_id, 
        c.parent_comment_id AS comment_id, 
        c.comment_id AS reply_id,  
        u.user_id AS uid, 
        u.Uname AS uname, 
        r.reason AS reason,
        r.report_time
    FROM CommentReports r
    JOIN Comments c ON c.comment_id = r.comment_id
    JOIN Users u ON u.user_id = c.commenter_id
    WHERE r.reporter_id = ?
    AND c.parent_comment_id IS NOT NULL
)
ORDER BY report_time DESC
LIMIT ?,?
`

    rows, err := db.Query(query, uid,uid,uid, startNumber, pageSize)
    if err != nil {
        // 输出详细的错误信息
        fmt.Println("SQL 错误：", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "isok": false,
            "failreason": fmt.Sprintf("查询举报数据失败，错误信息: %v", err),
        })
        return
    }
    
    defer rows.Close()

    var reports []gin.H
    for rows.Next() {
        var report struct {
            ReporterID  int    `json:"reporter_id"`
            Type        string `json:"type"`
            PostID      int   `json:"logid"`
            CommentID   int    `json:"commentid"`
            ReplyID     int    `json:"replyid"`
            UID         int    `json:"uid"`
            UName       string `json:"uname"`
            Reason      string `json:"reason"`
            ReportTime  string `json:"report_time"`
        }
        err := rows.Scan(&report.ReporterID, &report.Type, &report.PostID, &report.CommentID, &report.ReplyID, &report.UID, &report.UName, &report.Reason, &report.ReportTime)
        if err != nil {
            fmt.Println("Error scanning row:", err)  
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "解析举报数据失败"})
            return
        }
        
        if report.Type == "log" {
            reports = append(reports, gin.H{
                "rid":        report.ReporterID,
                "type":       report.Type,
                "logid":      report.PostID,
                "uid":        report.UID,
                "uname":      report.UName,
                "reason":     report.Reason,
            })
        }else if report.Type == "comment" {
            reports = append(reports, gin.H{
                "rid":         report.ReporterID,
                "type":       report.Type,
                "logid":     report.PostID,
                "commentid":  report.CommentID,
                "uid":        report.UID,
                "uname":      report.UName,
                "reason":     report.Reason,
            })
        }else if report.Type == "reply" {
            reports = append(reports, gin.H{
                "rid":         report.ReporterID,
                "type":       report.Type,
                "logid":     report.PostID,
                "replyid":    report.ReplyID,
                "uid":        report.UID,
                "uname":      report.UName,
                "reason":     report.Reason,
            })
        }

    }

    if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "遍历举报数据失败"})
        return
    }

    // 计算总页数
    var totalCount int
    countQuery := `
    SELECT COUNT(*) 
    FROM (
        SELECT r.reporter_id FROM PostReports r WHERE r.reporter_id = ?
        UNION ALL
        SELECT r.reporter_id FROM CommentReports r WHERE r.reporter_id = ?
    ) AS reports`
    err = db.QueryRow(countQuery, uid, uid).Scan(&totalCount)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询总数失败"})
        return
    }
    fmt.Println(totalCount)
    totalPages := (totalCount-1)/pageSize + 1

    c.JSON(http.StatusOK, gin.H{
        "isok":      true,
        "rptarget":  reports,
        "totalpages": totalPages,
    })
}


//获取举报目标详情
func GetReportInfo(c*gin.Context) {
    db, err := repository.Connect()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
        return
    }
    defer db.Close()

    uid:=c.Query("uid")
    type1:=c.Query("type")
    logid:=c.Query("logid")
    commentid:=c.Query("commentid")
    replyid:=c.Query("replyid")

    if uid==""{
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid参数"})
    }

    // 查询用户是否存在
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM Users WHERE user_id = ?", uid).Scan(&count)
    if err != nil || count == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": fmt.Sprintf("用户 %s 不存在", uid)})
        return
    }

    if type1 == "log" {
        if logid == "" {
            c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少logid参数"})
        }
        query:="SELECT LEFT(p.content,30) AS content,p.title,p.user_id,u.uname FROM Posts p JOIN Users u ON p.user_id = u.user_id WHERE p.post_id = ?"
        var loginfo struct {
            Content string `json:"content"`
            Title string `json:"title"`
            User_id string `json:"user_id"`
            Uname string `json:"uname"`
        }
        err = db.QueryRow(query, logid).Scan(&loginfo.Content,&loginfo.Title,&loginfo.User_id,&loginfo.Uname)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询帖子失败"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"isok": true, "postinfo": loginfo})
    }else if type1 == "comment" {
        if commentid == "" {
            c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少commentid参数"})
            return
        }
        query:="SELECT LEFT(c.content,30) AS content,c.commenter_id,u.uname FROM Comments c JOIN Users u ON c.commenter_id  = u.user_id WHERE c.comment_id = ?"
        var commentinfo struct {
            Content string `json:"content"`
            Commenter_id string `json:"commenter_id"`
            Uname string `json:"uname"`
        }
        err = db.QueryRow(query, commentid).Scan(&commentinfo.Content,&commentinfo.Commenter_id,&commentinfo.Uname)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询评论失败"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"isok": true, "commentinfo": commentinfo})
    }else if type1 == "reply" {
        if replyid == "" {
            c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少replyid参数"})
            return
        }
        query:="SELECT LEFT(r.content,30) AS content,r.commenter_id,u.uname FROM comments r JOIN Users u ON r.commenter_id  = u.user_id WHERE r.comment_id = ? "
        var replyinfo struct {
            Content string `json:"content"`
            Commenter_id string `json:"commenter_id"`
            Cname string `json:"uname"`
        }
        err = db.QueryRow(query, replyid).Scan(&replyinfo.Content,&replyinfo.Commenter_id,&replyinfo.Cname)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询回复失败"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"isok": true, "replyinfo": replyinfo})
    }else{
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "type参数错误"})
        return
    }
}



// 数据库中users表添加一行status，类型string，表示禁言还是正常的状态，例如alter table `users` add column `status` varchar(15) not null default "normal";
// 禁言接口
func HandleMute(c *gin.Context) {
	var req struct {
		Uid   string `json:"uid"`
		Day   string `json:"day"`
		Type  string `json:"type"`
		Rtype string `json:"rtype"`
		ID    string `json:"id"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求体格式错误"})
		return
	}

	if req.Uid == "" || req.Day == "" || req.Type == "" || req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少必要参数"})
		return
	}

	if req.Day != "-1" {
		_, err := strconv.Atoi(req.Day)
		if err != nil || req.Day == "" {
			c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的天数"})
			return
		}
	}

	if req.Type != "gag" && req.Type != "ban" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的type值"})
		return
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务开启失败"})
		return
	}

	if req.Type == "gag" {
		err = MuteUser(tx, req.Uid, req.Day, req.Rtype, req.ID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else if req.Type == "ban" {
		err = BanUser(tx, req.Uid, req.Day, req.Rtype, req.ID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的操作类型"})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "failreason": ""})
}

// 禁言用户
func MuteUser(tx *sql.Tx, uid string, day string, rtype string, id string) error {
	var endTime string
	if day == "-1" {
		endTime = "NULL"
	} else {
		endTime = fmt.Sprintf("DATE_ADD(NOW(), INTERVAL %s DAY)", day)
	}

	if rtype == "log" {
		_, err := tx.Exec("INSERT INTO usermutes (user_id, start_time, end_time) VALUES (?, NOW(), ?)", uid, endTime)
		if err != nil {
			return fmt.Errorf("禁言失败：%s", err.Error())
		}
	} else if rtype == "comment" {
		_, err := tx.Exec("INSERT INTO usermutes (user_id, start_time, end_time) VALUES (?, NOW(), ?)", uid, endTime)
		if err != nil {
			return fmt.Errorf("禁言失败：%s", err.Error())
		}
	} else {
		return fmt.Errorf("无效的违规内容类型")
	}

	return nil
}

// 用户封号
func BanUser(tx *sql.Tx, uid string, day string, rtype string, id string) error {
	if rtype == "log" {
		_, err := tx.Exec("UPDATE users SET status = 'banned' WHERE user_id = ?", uid)
		if err != nil {
			return fmt.Errorf("封号失败：%s", err.Error())
		}
	} else if rtype == "comment" {
		_, err := tx.Exec("UPDATE users SET status = 'banned' WHERE user_id = ?", uid)
		if err != nil {
			return fmt.Errorf("封号失败：%s", err.Error())
		}
	} else {
		return fmt.Errorf("无效的违规内容类型")
	}

	return nil
}
