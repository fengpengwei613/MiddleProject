package service

import (
	"fmt"
	"middleproject/internal/repository"
	"net/http"
	"strconv"
    "database/sql"
    "time"
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


    page := c.Query("page")


    // 将 page 转换为整数
    pageInt, err := strconv.Atoi(page)
    if err != nil || pageInt < 0 {
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的页数"})
        return
    }

    pageSize := 10// 每页 10 条数据
    startNumber := pageInt * pageSize


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
    WHERE c.parent_comment_id IS NULL
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
    WHERE c.parent_comment_id IS NOT NULL
)
ORDER BY report_time DESC
LIMIT ?,?
`

    rows, err := db.Query(query,startNumber, pageSize)
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
        SELECT r.reporter_id FROM PostReports r
        UNION ALL
        SELECT r.reporter_id FROM CommentReports r
    ) AS reports`
    err = db.QueryRow(countQuery).Scan(&totalCount)
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

    type1:=c.Query("type")
    logid:=c.Query("logid")
    commentid:=c.Query("commentid")
    replyid:=c.Query("replyid")

    if type1 == "log" {
        if logid == "" {
            c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少logid参数"})
        }
        query:="SELECT LEFT(p.content,30) AS content,p.title,p.user_id,u.uname FROM Posts p JOIN Users u ON p.user_id = u.user_id WHERE p.post_id = ?"
        var loginfo struct {
            Content string `json:"content"`
            Title string `json:"title"`
            //Images string `json:"images"`
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
        query1:="SELECT LEFT(p.content,30) AS content,p.title,p.user_id,u.uname FROM Posts p JOIN Users u ON p.user_id = u.user_id JOIN Comments c ON p.post_id = c.post_id WHERE c.comment_id = ?"
        var loginfo struct {
            Content string `json:"content"`
            Title string `json:"title"`
            //Images string `json:"images"`
            User_id string `json:"user_id"`
            Uname string `json:"uname"`
        }
        err = db.QueryRow(query1, commentid).Scan(&loginfo.Content,&loginfo.Title,&loginfo.User_id,&loginfo.Uname)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询评论对应的帖子失败"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"isok": true, "loginfo": loginfo,"commentinfo": commentinfo})
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

        query1:="SELECT LEFT(p.content,30) AS content,p.title,p.user_id,u.uname FROM Posts p JOIN Users u ON p.user_id = u.user_id JOIN Comments c ON p.post_id = c.post_id WHERE c.comment_id = ?"
        var loginfo struct {
            Content string `json:"content"`
            Title string `json:"title"`
            //Images string `json:"images"`
            User_id string `json:"user_id"`
            Uname string `json:"uname"`
        }
        err = db.QueryRow(query1, replyid).Scan(&loginfo.Content,&loginfo.Title,&loginfo.User_id,&loginfo.Uname)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询回复对应的帖子失败"})
            return
        }

        query2:="SELECT LEFT(c.content,30) AS content,c.commenter_id,u.uname FROM Comments c JOIN Users u ON c.commenter_id  = u.user_id JOIN Comments r ON c.comment_id = r.parent_comment_id WHERE r.comment_id = ?"
        var commentinfo struct {
            Content string `json:"content"`
            Commenter_id string `json:"commenter_id"`
            Uname string `json:"uname"`
        }
        err = db.QueryRow(query2, replyid).Scan(&commentinfo.Content,&commentinfo.Commenter_id,&commentinfo.Uname)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "查询回复对应的评论失败"})
            return
        }


        c.JSON(http.StatusOK, gin.H{"isok": true,"loginfo": loginfo,"commentinfo": commentinfo,"replyinfo": replyinfo})
    }else{
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "type参数错误"})
        return
    }
}






type UserMuteStatus struct {
	Status   string `json:"status"`      
	Lifttime string `json:"lifttime"`    
	Days     int    `json:"days"`        
}

//获取用户状态
func GetUserStatus(c *gin.Context) {
    db, err := repository.Connect()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
        return
    }
    defer db.Close()
    
    uid:= c.Query("uid")
    if uid == "" {
        c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少uid参数"})
        return
    }
	query := `
		SELECT type, start_time,end_time
		FROM usermutes 
		WHERE user_id = ? AND NOW() BETWEEN start_time AND end_time 
		ORDER BY start_time DESC LIMIT 1`
    
    var muteType int
	var startTimeBytes, endTimeBytes []byte 
    err = db.QueryRow(query, uid).Scan(&muteType,&startTimeBytes, &endTimeBytes)
	if err == sql.ErrNoRows {
		// 用户没有封禁或禁言记录
		c.JSON(http.StatusOK, gin.H{
			"status": "normal",
		})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询禁言/封禁记录失败"})
		return
	}

    startTime, err := time.Parse("2006-01-02 15:04:05", string(startTimeBytes))
    if err != nil {
        fmt.Printf("Error parsing start_time: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid start_time"})
        return
    }

    endTime, err := time.Parse("2006-01-02 15:04:05", string(endTimeBytes))
    if err != nil {
        fmt.Printf("Error parsing end_time: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid end_time"})
        return
    }
    days:=int(endTime.Sub(startTime).Hours() / 24)
    fmt.Println(days)
	var status string
	var lifttime string

	if muteType == 0 {
		// 封禁状态
		status = "baned"
		lifttime = endTime.Format("2006-01-02 15:04:05")
	} else if muteType == 1 {
		// 禁言状态
		status = "restricted"
		lifttime = endTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK,UserMuteStatus{
        Status:   status,
		Lifttime: lifttime,
		Days:     days,
	})

}



// 解除禁言封禁接口
func HandleUnmute(c *gin.Context) {
	var req struct {
		Uid string `json:"uid"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求参数格式错误"})
		return
	}

	if req.Uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少必要参数"})
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

	err = UnmuteUser(tx, req.Uid)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "failreason": ""})
}

// 解除禁言封禁用户
func UnmuteUser(tx *sql.Tx, uid string) error {
	_, err := tx.Exec("DELETE FROM usermutes WHERE user_id = ? AND (type = 0 OR type = 1)", uid)
	if err != nil {
		return fmt.Errorf("解除禁言封禁失败：%s", err.Error())
	}

	return nil
}

// 增加或减少禁言封禁天数
func HandleUpdateMuteTime(c *gin.Context) {
	var req struct {
		Uid  string `json:"uid"`
		Days int    `json:"days"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "请求参数格式错误"})
		return
	}

	if req.Uid == "" || req.Days == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "缺少必要参数"})
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

	err = UpdateMuteTime(tx, req.Uid, req.Days)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": err.Error()})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "failreason": ""})
}

// 增加/减少禁言封禁天数
func UpdateMuteTime(tx *sql.Tx, uid string, days int) error {
	query := `UPDATE usermutes SET end_time = DATE_ADD(end_time, INTERVAL ? DAY) WHERE user_id = ?`
	_, err := tx.Exec(query, days, uid)
	return err
}
