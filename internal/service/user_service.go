package service

import (
	"database/sql"
	"fmt"
	"math/rand"
	"middleproject/internal/model"
	"middleproject/internal/repository"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
		return
	}
	var data model.User
	if err_bind := c.ShouldBindJSON(&data); err_bind != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "注册绑定请求数据失败"})
		return
	}

	//校验最新验证码
	query := "SELECT code FROM verificationcodes WHERE email = ? AND expiration > NOW() ORDER BY expiration DESC LIMIT 1"
	row := db.QueryRow(query, data.Email)
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
	if err_re != nil || userid == "0" {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	}
	//默认头像地址
	avatar_0 := "postImage/image0.png"
	err_u, url := scripts.GetUrl(avatar_0)
	if err_u != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": url})
		return
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
		return

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

// 辅助函数：判断是否是邮箱格式
func isEmailFormat(input string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, input)
	return matched
}

// 用户登录函数实现
func Login(c *gin.Context) {
	var requestData model.LoginRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的请求数据"})
		return
	}
	fmt.Println(requestData)
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

	isEmail := isEmailFormat(requestData.Userid)
	fmt.Println("uid1", requestData.Userid)
	var query string
	if isEmail {
		query = "SELECT user_id, password, Uname, avatar FROM users WHERE email = ?"
	} else {
		query = "SELECT user_id, password, Uname, avatar FROM users WHERE user_id = ?"
	}
	fmt.Println(query)
	fmt.Println(requestData.Userid)
	row := db.QueryRow(query, requestData.Userid)
	info := row.Scan(&userID, &storedPassword, &userName, &Avatar)

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

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库事务开始失败"})
		return
	}
	var newsetting model.UpdatePersonalSettings
	if err := c.ShouldBindJSON(&newsetting); err != nil {
		tx.Rollback()
		//fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的 JSON 数据"})
		return
	}

	// 检查数据库中是否存在指定的 UID
	var userID string
	checkUserQuery := "SELECT user_id FROM users WHERE user_id = ?"
	err2 := db.QueryRow(checkUserQuery, newsetting.UserId).Scan(&userID)

	if err2 == sql.ErrNoRows {
		tx.Rollback()
		c.JSON(404, gin.H{"isok": false, "failreason": "用户不存在"})
		return
	} else if err != nil {
		tx.Rollback()
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
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"isok": false, "failreason": "无效的列名"})
		return
	}

	updateQuery := fmt.Sprintf("UPDATE users SET %s = ? WHERE user_id = ?", newsetting.Type)

	//SQL更新语句
	_, err1 := db.Exec(updateQuery, newsetting.Value, newsetting.UserId)
	if err1 != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"isok": false, "failreason": "更新失败"})
		return
	}
	err = tx.Commit()
	if err != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "事务提交失败"})
		return
	}
	c.JSON(200, gin.H{"isok": true})
}

// 找回密码
func ForgotPassword(c *gin.Context) {
	var requestData model.ResetPasswordReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "绑定请求数据失败", "uid": "", "uname": "", "uimage": ""})
		return
	}

	if !repository.VerifyCode(requestData.Mail, requestData.Code) {
		c.JSON(400, gin.H{"isok": false, "failreason": "验证码错误", "uid": "", "uname": "", "uimage": ""})
		return
	}

	user, err := updatePassword(requestData.Mail, requestData.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"isok": false, "failreason": "用户不存在", "uid": "", "uname": "", "uimage": ""})
			return
		}
		c.JSON(500, gin.H{"isok": false, "failreason": "更新失败", "uid": "", "uname": "", "uimage": ""})
		return
	}

	c.JSON(200, gin.H{
		"isok":       true,
		"failreason": "",
		"uid":        user.UserID,
		"uname":      user.Uname,
	})
}

// 更新用户密码
func updatePassword(mail string, newPassword string) (*model.User, error) {

	db, err := repository.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	stmt, err := tx.Prepare("UPDATE users SET password = ? WHERE email = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newPassword, mail)
	if err != nil {
		return nil, err
	}

	var user model.User
	row := tx.QueryRow("SELECT user_id, uname, avatar FROM users WHERE email = ?", mail)
	if err := row.Scan(&user.UserID, &user.Uname); err != nil {
		return nil, err
	}

	return &user, nil
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
        SELECT user_id, uname, uimage, phone, mail, address, birthday, regtime, sex, 
               introduction, schoolname, major, edutime, edulevel, companyname, 
               positionname, industry, interests, likenum, attionnum, isattion, fansnum
        FROM users WHERE user_id = ?`, userID)

	var userInfo struct {
		UserID       int      `json:"userID"`
		UserName     string   `json:"userName"`
		UImage       string   `json:"uimage"`
		Phone        string   `json:"phone"`
		Mail         string   `json:"mail"`
		Address      string   `json:"address"`
		Birthday     string   `json:"birthday"`
		RegTime      string   `json:"regtime"`
		Sex          string   `json:"sex"`
		Introduction string   `json:"introduction"`
		SchoolName   string   `json:"schoolname"`
		Major        string   `json:"major"`
		EduTime      string   `json:"edutime"`
		EduLevel     string   `json:"edulevel"`
		CompanyName  string   `json:"companyname"`
		PositionName string   `json:"positionname"`
		Industry     string   `json:"industry"`
		Interests    []string `json:"interests"`
		LikeNum      string   `json:"likenum"`
		AttionNum    string   `json:"attionnum"`
		IsAttion     string   `json:"isattion"`
		FansNum      string   `json:"fansnum"`
	}
	err = row.Scan(
		&userInfo.UserID, &userInfo.UserName, &userInfo.UImage, &userInfo.Phone, &userInfo.Mail,
		&userInfo.Address, &userInfo.Birthday, &userInfo.RegTime, &userInfo.Sex,
		&userInfo.Introduction, &userInfo.SchoolName, &userInfo.Major, &userInfo.EduTime,
		&userInfo.EduLevel, &userInfo.CompanyName, &userInfo.PositionName, &userInfo.Industry,
		&userInfo.Interests, &userInfo.LikeNum, &userInfo.AttionNum, &userInfo.IsAttion, &userInfo.FansNum,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"isok": false, "failreason": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "数据库查询失败"})
		}
		return
	}

	if len(userInfo.Interests) == 0 {
		userInfo.Interests = []string{}
	}

	if userInfo.UImage != "" {
		userInfo.UImage = strings.Trim(userInfo.UImage, "[]")
		images := strings.Split(userInfo.UImage, ",")

		if len(images) > 0 {
			err, url := scripts.GetUrl(images[0])
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "获取头像失败"})
				return
			}
			userInfo.UImage = url
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"isok":         true,
		"userID":       userInfo.UserID,
		"userName":     userInfo.UserName,
		"uimage":       userInfo.UImage,
		"phone":        userInfo.Phone,
		"mail":         userInfo.Mail,
		"address":      userInfo.Address,
		"birthday":     userInfo.Birthday,
		"regtime":      userInfo.RegTime,
		"sex":          userInfo.Sex,
		"introduction": userInfo.Introduction,
		"schoolname":   userInfo.SchoolName,
		"major":        userInfo.Major,
		"edutime":      userInfo.EduTime,
		"edulevel":     userInfo.EduLevel,
		"companyname":  userInfo.CompanyName,
		"positionname": userInfo.PositionName,
		"industry":     userInfo.Industry,
		"interests":    userInfo.Interests,
		"likenum":      userInfo.LikeNum,
		"attionnum":    userInfo.AttionNum,
		"isattion":     userInfo.IsAttion,
		"fansnum":      userInfo.FansNum,
	})
}

func UpdatePersonalInfo(c *gin.Context) {

	db, err := repository.Connect()
	if err != nil {

		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "数据库连接失败",
		})
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "数据库事务开始失败",
		})
		return
	}

	var updateData struct {
		UserName  string   `json:"userName"`
		Phone     string   `json:"phone"`
		Email     string   `json:"email"`
		Address   string   `json:"address"`
		Avatar    string   `json:"avatar"`
		Signature string   `json:"signature"`
		Birthday  string   `json:"birthday"`
		Interests []string `json:"interests"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		tx.Rollback()
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "无效的请求数据",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		tx.Rollback()
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "未授权",
		})
		return
	}

	var dbUserID string
	checkUserQuery := "SELECT user_id FROM users WHERE user_id = ?"
	err2 := db.QueryRow(checkUserQuery, userID).Scan(&dbUserID)
	if err2 == sql.ErrNoRows {
		tx.Rollback()
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "用户不存在",
		})
		return
	} else if err2 != nil {
		tx.Rollback()
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "查询失败",
		})
		return
	}

	validColumns := map[string]bool{
		"userName":  true,
		"phone":     true,
		"email":     true,
		"address":   true,
		"avatar":    true,
		"signature": true,
		"birthday":  true,
		"interests": true,
	}

	if !validColumns["userName"] || !validColumns["phone"] || !validColumns["email"] {
		tx.Rollback()
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "无效的字段",
		})
		return
	}

	var interests interface{}
	if len(updateData.Interests) == 0 {
		interests = nil
	} else {
		interests = updateData.Interests
	}

	updateQuery := `
		UPDATE users
		SET uname = ?, phone = ?, email = ?, address = ?, avatar = ?, signature = ?, birthday = ?, interests = ?
		WHERE user_id = ?`

	_, err = tx.Exec(updateQuery,
		updateData.UserName,
		updateData.Phone,
		updateData.Email,
		updateData.Address,
		updateData.Avatar,
		updateData.Signature,
		updateData.Birthday,
		interests,
		userID)

	if err != nil {
		tx.Rollback()
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "更新失败",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(200, gin.H{
			"isok":       false,
			"failreason": "事务提交失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"isok":    true,
		"message": "个人信息更新成功",
		"userID":  userID,
		"updated": updateData,
	})
}
