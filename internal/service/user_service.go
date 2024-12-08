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

//忘记密码
func ForgotPassword(c *gin.Context) {
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "连接数据库失败"})
		return
	}

	var requestData model.ResetPasswordReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "绑定请求数据失败"})
		return
	}
	query := "SELECT code FROM verificationcodes WHERE email = ? AND expiration > NOW() ORDER BY expiration DESC LIMIT 1"
	row := db.QueryRow(query, requestData.Mail)
	var code string
	err_check := row.Scan(&code)
	if err_check != nil || code != requestData.Code {
		c.JSON(400, gin.H{"isok": false, "failreason": "验证码错误"})
		return
	}

	err_re, user, result := updatePassword(db, requestData.Mail, requestData.Password)
	if err_re != nil || user.UserID == 0 {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	}

	c.JSON(200, gin.H{
		"isok":       true,
		"failreason": "",
		"uid":        user.UserID,
		"uname":      user.Uname,
		"uimage":     user.Avatar,
	})
}

//更新密码
func updatePassword(db *sql.DB, mail string, newPassword string) (error, model.User, string) {

	tx, err := db.Begin()
	if err != nil {
		return err, model.User{}, "开启事务失败"
	}
	defer tx.Rollback() 

	stmt, err := tx.Prepare("UPDATE users SET password = ? WHERE email = ?")
	if err != nil {
		return err, model.User{}, "准备更新语句失败"
	}
	defer stmt.Close()

	var user model.User
	row := tx.QueryRow("SELECT user_id, uname, avatar FROM users WHERE email = ?", mail)
	if err := row.Scan(&user.UserID, &user.Uname, &user.Avatar); err != nil {
		return err, model.User{}, "查询用户信息失败"
	}

	_, err = stmt.Exec(newPassword, mail)
	if err != nil {
		return err, model.User{}, "更新密码失败"
	}

	err = tx.Commit()
	if err != nil {
		return err, model.User{}, "提交事务失败"
	}

	return nil, user, ""
}

// 获取个人信息函数
func GetPersonalInfo(db *sql.DB, uid int) (*model.PersonalInfo, error) {
	// 定义 SQL 查询语句
	query := `
        SELECT user_id, Uname, avatar, phone, email, address, birthday, registration_date, 
           sex, introduction, school, major, edutime, eduleval, companyname, 
           positionname, industry, interests, likenum, attionnum, fansnum
        FROM users WHERE user_id = ?`

	row := db.QueryRow(query, uid)
        info := &model.PersonalInfo{}

	err := row.Scan(
		&info.UserID, &info.UserName, &info.UImage, &info.Phone, &info.Mail,
		&info.Address, &info.Birthday, &info.RegTime, &info.Sex,
		&info.Introduction, &info.SchoolName, &info.Major, &info.EduTime,
		&info.EduLevel, &info.CompanyName, &info.PositionName, &info.Industry,
		&info.Interests, &info.LikeNum, &info.AttionNum, &info.IsAttion, &info.FansNum,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("数据库查询失败: %v", err)
	}

	if info.Interests == nil || len(info.Interests) == 0 {
		info.Interests = []string{}
	} else if len(info.Interests) == 1 && info.Interests[0] == "" {
		info.Interests = []string{}
	} else {
		info.Interests = strings.Split(info.Interests[0], ",")
	}

	return info, nil
}

// 处理获取个人信息的请求
func HandleGetPersonalInfo(c *gin.Context) {

	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库连接失败"})
		return
	}
	defer db.Close()

	uid := c.DefaultQuery("uid", "") // 获取 uid 参数
	id, err := strconv.Atoi(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	info, err := GetPersonalInfo(db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":       info.UserID,
		"userName":     info.UserName,
		"uimage":       info.UImage,
		"phone":        info.Phone,
		"mail":         info.Mail,
		"address":      info.Address,
		"birthday":     info.Birthday,
		"regtime":      info.RegTime,
		"sex":          info.Sex,
		"introduction": info.Introduction,
		"schoolname":   info.SchoolName,
		"major":        info.Major,
		"edutime":      info.EduTime,
		"edulevel":     info.EduLevel,
		"companyname":  info.CompanyName,
		"positionname": info.PositionName,
		"industry":     info.Industry,
		"interests":    info.Interests,
		"likenum":      info.LikeNum,
		"attionnum":    info.AttionNum,
		"isattion":     info.IsAttion,
		"fansnum":      info.FansNum,
	})
}


// 更新个人信息
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

	// 获取用户 ID
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
