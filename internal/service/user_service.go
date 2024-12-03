package service

import (
	"database/sql"
	"fmt"
	"math/rand"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"middleproject/scripts"

	_ "github.com/go-sql-driver/mysql"
)

// register 函数实现
func Register(c *gin.Context) {
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "连接数据库失败"})
	}
	var data model.User
	if err_bind := c.ShouldBindJSON(&data); err_bind != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "注册绑定请求数据失败"})
		return
	}
	//校验最新验证码
	query := "SELECT code FROM verificationcodes WHERE email = ? AND expiration > NOW() ORDER BY expiration DESC LIMIT 1"
	row := db.QueryRow(query, data.Email)
	fmt.Println(row)
	var code string
	err_check := row.Scan(&code)
	fmt.Println("code:", code)
	fmt.Println("data.VerifyCode:", data.VerifyCode)
	if err_check != nil || code != data.VerifyCode {
		c.JSON(400, gin.H{"isok": false, "failreason": "验证码错误"})
		return
	}
	//添加到数据库
	err_re, result, userid := data.CreateUser()
	if err_re != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	}
	//默认头像地址
	avatar_0 := "postImage/image0.png"
	err_u, url := scripts.GetUrl(avatar_0)
	if err_u != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": url})
	}
	c.JSON(200, gin.H{"isok": true, "uid": userid, "uimage": url})
}

func SendMailInterface(c *gin.Context) {
	var requestData map[string]string
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "绑定请求数据失败"})
		return
	}
	mail, ok := requestData["mail"]
	type_server := c.DefaultQuery("type", "no")
	if !ok {
		c.JSON(400, gin.H{"isok": false, "failreason": "缺少邮箱"})
		return
	}
	//检查mail是否已经注册
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "连接数据库失败"})
	}
	query := "SELECT email FROM Users WHERE email = ?"
	row := db.QueryRow(query, mail)
	var email string
	err_check := row.Scan(&email)
	if err_check == nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "邮箱已经注册"})
		return
	}
	//生成随机数
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(999999-100000+1) + 100000
	strnum := strconv.Itoa(randomNum)
	//strnum := "123456"
	var result string
	if type_server == "regist" {
		result = scripts.SendEmail(mail, "注册验证码", strnum)
	} else if type_server == "find" {
		result = scripts.SendEmail(mail, "找回密码验证码", strnum)
	} else {
		c.JSON(400, gin.H{"isok": false, "failreason": "无效的type"})
		return
	}
	if result != "成功" {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	} else {
		c.JSON(200, gin.H{"isok": true})
	}

}

// 用户登录函数实现
func Login(c *gin.Context) {
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
	// 检查用户名和密码

	row := db.QueryRow("SELECT user_id,password,uname,avatar FROM users WHERE user_id = ?", requestData.Userid)
	var storedPassword string
	var userID int
	var userName string
	var Avatar string

	info := row.Scan(&userID, &storedPassword, &userName, &Avatar)
	if info != nil {
		if info == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"isok": false, "failreason": "用户不存在"})
		}
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库查询失败"})
	}
	if storedPassword != requestData.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "密码错误"})
		return
	}
	err, Avatar = scripts.GetUrl(Avatar)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": Avatar})
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "uid": userID, "uname": userName, "uimage": Avatar})
}

// 获取个人设置函数
func GetPersonalSettings(db *sql.DB, uid int) (*model.PersonalSettings, error) {

	query := "SELECT showlike,showcollect,showphone,showmail FROM users WHERE user_id = ?"
	row := db.QueryRow(query, uid)
	settings := &model.PersonalSettings{}
	err := row.Scan(&settings.ShowLike, &settings.ShowCollect, &settings.ShowPhone, &settings.ShowMail)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("数据库查询失败")

	}
	return settings, nil
}

// 处理获取个人设置的请求
func HandleGetPersonalSettings(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库连接失败"})
		return
	}
	defer db.Close()

	uid := c.Query("uid")
	id, err := strconv.Atoi(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	settings, err := GetPersonalSettings(db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"showlike":    settings.ShowLike,
		"showcollect": settings.ShowCollect,
		"showphone":   settings.ShowPhone,
		"showmail":    settings.ShowMail,
	})
}

func UpdatePersonalSettings(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库连接失败"})
		return
	}
	defer db.Close()
	var newsetting model.UpdatePersonalSettings
	if err := c.ShouldBindJSON(&newsetting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的 JSON 数据"})
		return
	}

	// 检查数据库中是否存在指定的 UID
	var userID string
	checkUserQuery := "SELECT user_id FROM users WHERE user_id = ?"
	err2 := db.QueryRow(checkUserQuery, newsetting.UserId).Scan(&userID)

	if err2 == sql.ErrNoRows {
		c.JSON(404, gin.H{"isok": false, "failreason": "用户不存在"})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "查询失败"})
		return
	}

	// 确保请求中的列名合法
	validColumns := map[string]bool{
		"showlike":    true, // 允许更新的列名
		"showcollect": true,
		"showphone":   true,
		"showmail":    true,
	}
	if !validColumns[newsetting.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的列名"})
		return
	}

	updateQuery := fmt.Sprintf("UPDATE users SET %s = ? WHERE user_id = ?", newsetting.Type)

	//SQL更新语句
	_, err1 := db.Exec(updateQuery, newsetting.Value, newsetting.UserId)
	if err1 != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "更新失败"})
		return
	}
	c.JSON(200, gin.H{"isok": true})
}

// 找回密码
func ForgotPassword(c *gin.Context) {
	var request model.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT user_id FROM users WHERE email = ?", request.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"isok": false, "failreason": "邮箱未注册"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库查询失败"})
		}
		return
	}

	requestData := map[string]string{
		"mail": request.Email,
	}
	c.Set("type", "find")
	c.Set("requestData", requestData)

	SendMailInterface(c)

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": "验证码已发送到您的邮箱"})
}

// 获取个人信息
func GetPersonalInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "未授权"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	row := db.QueryRow(`
        SELECT user_id, uname, phone, email, address, avatar, signature, birthday
        FROM users WHERE user_id = ?`, userID)

	var userName, phone, email, address, avatar, signature string
	var birthday string
	var uid int

	err = row.Scan(&uid, &userName, &phone, &email, &address, &avatar, &signature, &birthday)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"isok": false, "failreason": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库查询失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isok":      true,
		"userID":    uid,
		"userName":  userName,
		"phone":     phone,
		"email":     email,
		"address":   address,
		"avatar":    avatar,
		"signature": signature,
		"birthday":  birthday,
	})
}

// 更新个人信息
func UpdatePersonalInfo(c *gin.Context) {
	var updateData model.UpdatePersonalInfoRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "未授权"})
		return
	}

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(`
        UPDATE users
        SET uname = ?, phone = ?, email = ?, address = ?, avatar = ?, signature = ?, birthday = ?
        WHERE user_id = ?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "准备更新失败"})
		return
	}

	_, err = stmt.Exec(
		updateData.UserName,
		updateData.Phone,
		updateData.Email,
		updateData.Address,
		updateData.Avatar,
		updateData.Signature,
		updateData.Birthday,
		userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"isok": true, "message": "个人信息更新成功"})
}
