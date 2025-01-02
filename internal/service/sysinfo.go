package service

import (
	"middleproject/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
)

type Msgobj struct {
	Msgid   int    `json:"msgid"`
	Type    string `json:"type"`
	Time    string `json:"time"`
	Content string `json:"content"`
}

func Getsysinfo(c *gin.Context) {
	uidstr := c.DefaultQuery("uid", "-1")
	pagestr := c.DefaultQuery("page", "-1")
	uid, err_uid := strconv.Atoi(uidstr)
	page, err_page := strconv.Atoi(pagestr)
	var posts []Msgobj
	if err_uid != nil || uid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"msgobj": posts, "totalPages": -1})
		return
	}
	if err_page != nil || page == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"msgobj": posts, "totalPages": -1})
		return
	}
	page = page - 1

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": -1})
		return
	}
	defer db.Close()
	var total int
	err = db.QueryRow("SELECT COUNT(*) FROM sysinfo WHERE uid = ?", uid).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": -1})
		return
	}
	var totalPages int
	totalPages = total / 5
	if total%5 != 0 {
		totalPages++
	}
	rows, err := db.Query("SELECT * FROM sysinfo WHERE uid = ? ORDER BY time DESC limit ?,5", uid, page*5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": -1})
		return
	}
	for rows.Next() {
		var msg Msgobj
		err = rows.Scan(&msg.Msgid, &msg.Type, &msg.Content, &msg.Time, &uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": -1})
			return
		}
		msg.Time = msg.Time[0 : len(msg.Time)-3]
		posts = append(posts, msg)
	}

	c.JSON(http.StatusOK, gin.H{"msgobj": posts, "totalPages": totalPages})

}

func Del_sysinfo(c *gin.Context) {
	msgidstr := c.DefaultQuery("msgid", "-1")
	msgid, err_msgid := strconv.Atoi(msgidstr)
	if err_msgid != nil || msgid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "filereason": "msgid错误"})
		return
	}
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "filereason": "数据库连接失败"})
		return
	}
	defer db.Close()
	_, err = db.Exec("DELETE FROM sysinfo WHERE id = ?", msgid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "filereason": "sql删除消息失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true})
}
