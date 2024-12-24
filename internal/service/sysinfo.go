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
	uid, err_uid := strconv.Atoi(uidstr)
	var posts []Msgobj
	if err_uid != nil || uid == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"msgobj": posts, "totalPages": 0})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": 0})
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM sysinfo WHERE uid = ? ORDER BY time DESC", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": 0})
		return
	}
	for rows.Next() {
		var msg Msgobj
		err = rows.Scan(&msg.Msgid, &msg.Type, &msg.Time, &msg.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msgobj": posts, "totalPages": 0})
			return
		}
		msg.Time = msg.Time[0 : len(msg.Time)-3]
		posts = append(posts, msg)
	}
	c.JSON(http.StatusOK, gin.H{"msgobj": posts, "totalPages": 0})

}
