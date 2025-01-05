package service

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"middleproject/internal/middleware"
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
	token, err := middleware.GenerateToken(userid, "user")
	if err != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "生成 token 失败"})
		return
	}
	c.JSON(200, gin.H{"isok": true, "uid": userid, "uimage": url, "token": token})
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
	location, _ := time.LoadLocation("Asia/Shanghai") // 设置时区为上海
	time.Local = location
	//生成随机数
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(999999-100000+1) + 100000
	strnum := strconv.Itoa(randomNum)
	//strnum := "123456"
	var result string
	if type_server == "regist" {
		result = scripts.SendEmail(mail, "注册验证码", strnum, "regist")
	} else if type_server == "find" {
		result = scripts.SendEmail(mail, "找回密码验证码", strnum, "forget")
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
	//查询用户是否被封禁
	var userPermission int
	newquery := "SELECT user_id from usermutes where user_id = ? and end_time > now() and type = 0"
	myerr := db.QueryRow(newquery, requestData.Userid).Scan(&userPermission)
	if myerr == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"isok": false, "failreason": "您已被封禁"})
		return
	}

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
	jwtToken, err := middleware.GenerateToken(userID, "user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"isok": false, "failreason": "生成 token 失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"isok": true, "uid": userID, "uname": userName, "uimage": Avatar, "token": jwtToken})

	//c.JSON(http.StatusOK, gin.H{"isok": true, "uid": userID, "uname": userName, "uimage": Avatar})
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

// 忘记密码
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
	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", requestData.Mail).Scan(&userExists)
	if err != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "检查用户是否存在时发生错误"})
		return
	}
	if !userExists {
		c.JSON(400, gin.H{"isok": false, "failreason": "用户不存在"})
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
	err, avatarURL := scripts.GetUrl(user.Avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "获取头像签名失败"})
		return
	}
	c.JSON(200, gin.H{
		"isok":       true,
		"failreason": "",
		"uid":        user.UserID,
		"uname":      user.Uname,
		"uimage":     avatarURL,
	})
}

// 更新密码
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
func GetPersonalInfo(db *sql.DB, uid string, requestid string, tokenstring string) (*model.PersonalInfo, error) {
	query := `
	SELECT signature, user_id, Uname, avatar, phone, email, address, birthday, registration_date, 
		   sex, introduction, school, major, edutime, eduleval, companyname, positionname, 
		   industry, interests, likenum, attionnum, fansnum,showphone,showmail
	FROM users WHERE user_id = ?`

	info := &model.PersonalInfo{}
	var (
		signatureNULL    sql.NullString
		avatarNull       sql.NullString
		phoneNull        sql.NullString
		emailNull        sql.NullString
		addressNull      sql.NullString
		birthdayNull     sql.NullString
		registrationDate sql.NullString
		sexNull          sql.NullInt64
		introductionNull sql.NullString
		schoolNull       sql.NullString
		majorNull        sql.NullString
		edutimeNull      sql.NullString
		edulevelNull     sql.NullString
		companyNull      sql.NullString
		positionNull     sql.NullString
		industryNull     sql.NullString
		interestsNull    sql.NullString
		likenumNull      sql.NullInt64
		attionumNull     sql.NullInt64
		fansnumNull      sql.NullInt64
		showphoneNull    sql.NullInt64
		showmailNull     sql.NullInt64
	)

	err := db.QueryRow(query, uid).Scan(&signatureNULL,
		&info.UserID, &info.UserName, &avatarNull, &phoneNull, &emailNull, &addressNull, &birthdayNull, &registrationDate,
		&sexNull, &introductionNull, &schoolNull, &majorNull, &edutimeNull, &edulevelNull, &companyNull, &positionNull,
		&industryNull, &interestsNull, &likenumNull, &attionumNull, &fansnumNull, &showphoneNull, &showmailNull)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户不存在")
		}
		return nil, fmt.Errorf("数据库查询失败: %v", err)
	}

	var isAttention bool
	followQuery := "SELECT COUNT(*) FROM userfollows WHERE follower_id = ? AND followed_id = ?"
	err = db.QueryRow(followQuery, requestid, uid).Scan(&isAttention)

	if err != nil {
		return nil, fmt.Errorf("检查关注状态失败: %v", err)
	}
	info.IsAttion = isAttention
	info.Signature = signatureNULL.String
	if avatarNull.Valid {
		err, signedURL := scripts.GetUrl(avatarNull.String)
		if err != nil {
			return nil, fmt.Errorf("获取头像URL失败: %v", err)
		}
		info.UImage = signedURL
	} else {
		info.UImage = ""
	}
	var showphone bool
	var showmail bool
	if showphoneNull.Valid {
		if showphoneNull.Int64 == 1 {
			showphone = true
		} else {
			showphone = false
		}
	}
	if showmailNull.Valid {
		if showmailNull.Int64 == 1 {
			showmail = true
		} else {
			showmail = false
		}
	}
	_, err_token := middleware.ParseToken(tokenstring)

	if showmail == false {
		if err_token != nil {
			info.Phone = "已隐藏"
		} else if requestid != uid {
			info.Phone = "已隐藏"
		} else {
			info.Phone = phoneNull.String
		}
	} else {
		info.Phone = phoneNull.String
	}
	if showphone == false {
		if err_token != nil {
			info.Mail = "已隐藏"
		} else if requestid != uid {
			info.Mail = "已隐藏"
		} else {
			info.Mail = emailNull.String
		}
	} else {
		info.Mail = emailNull.String
	}

	info.Address = addressNull.String
	info.Birthday = birthdayNull.String
	info.RegTime = registrationDate.String
	if sexNull.Valid {
		if sexNull.Int64 == 1 {
			info.Sex = "男"
		} else if sexNull.Int64 == 2 {
			info.Sex = "女"
		} else {
			info.Sex = "未知"
		}
	} else {
		info.Sex = "未知"
	}
	info.Introduction = introductionNull.String
	info.SchoolName = schoolNull.String
	info.Major = majorNull.String
	info.EduTime = edutimeNull.String
	info.EduLevel = edulevelNull.String
	info.CompanyName = companyNull.String
	info.PositionName = positionNull.String
	info.Industry = industryNull.String
	if interestsNull.Valid {
		var interestsArray []string
		err := json.Unmarshal([]byte(interestsNull.String), &interestsArray)
		if err != nil {
			return nil, fmt.Errorf("解析兴趣字段失败: %v", err)
		}
		info.Interests = interestsArray
	} else {
		info.Interests = []string{}
	}
	info.LikeNum = strconv.FormatInt(likenumNull.Int64, 10)
	info.AttionNum = strconv.FormatInt(likenumNull.Int64, 10)
	info.FansNum = strconv.FormatInt(fansnumNull.Int64, 10)

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
	// 从请求头中获取 Authorization
	tokenString := c.GetHeader("Authorization")

	// 去掉 "Bearer " 前缀
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	uid := c.Query("uid")
	requestid := c.Query("requesteruid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

	personalInfo, err := GetPersonalInfo(db, uid, requestid, tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	}

	interests := personalInfo.Interests
	interestsFormatted := make([]string, len(interests))
	for i, interest := range interests {
		interestsFormatted[i] = strings.Trim(interest, "\"\n") // 移除字符串中的引号和换行符
	}

	c.JSON(http.StatusOK, personalInfo)
}

// 验证电话号码是否为11位数字
func isValidPhoneNumber(phone string) bool {
	re := regexp.MustCompile(`^\d{11}$`)
	return re.MatchString(phone)
}
func decodeBase64Image(base64Str string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("解码 Base64 字符串失败: %v", err)
	}
	return data, nil
}

// 更新个人信息
func UpdatePersonal(db *sql.DB, uid, fieldType, value string) error {
	var query string
	var err error

	if fieldType == "sex" {
		if value == "男" {
			value = "1"
			query = "UPDATE users SET sex = ? WHERE user_id = ?"
		} else if value == "女" {
			value = "2"
			query = "UPDATE users SET sex = ? WHERE user_id = ?"
		} else {
			return fmt.Errorf("无效的性别值: %s", value)
		}
		_, err = db.Exec(query, uid)
	} else if fieldType == "uimage" {
		filename := "avatar_" + uid + ".png"
		err_up, imageURL := scripts.UploadImage(value, filename)
		if err_up != nil {
			return fmt.Errorf("上传图片失败: %v", imageURL)
		}

		query = "UPDATE users SET avatar = ? WHERE user_id = ?"
		value = imageURL
	} else {
		switch fieldType {
		case "uname":
			query = "UPDATE users SET Uname = ? WHERE user_id = ?"
		case "phone":
			if !isValidPhoneNumber(value) {
				return fmt.Errorf("无效的电话号码")
			}
			query = "UPDATE users SET phone = ? WHERE user_id = ?"
		case "mail":
			query = "UPDATE users SET email = ? WHERE user_id = ?"
		case "address":
			query = "UPDATE users SET address = ? WHERE user_id = ?"
		case "birthday":
			query = "UPDATE users SET birthday = ? WHERE user_id = ?"
		case "introduction":
			query = "UPDATE users SET introduction = ? WHERE user_id = ?"
		case "schoolname":
			query = "UPDATE users SET school = ? WHERE user_id = ?"
		case "major":
			query = "UPDATE users SET major = ? WHERE user_id = ?"
		case "edutime":
			query = "UPDATE users SET edutime = ? WHERE user_id = ?"
		case "edulevel":
			query = "UPDATE users SET eduleval = ? WHERE user_id = ?"
		case "companyname":
			query = "UPDATE users SET companyname = ? WHERE user_id = ?"
		case "positionname":
			query = "UPDATE users SET positionname = ? WHERE user_id = ?"
		case "industry":
			query = "UPDATE users SET industry = ? WHERE user_id = ?"
		case "interests":
			interests := strings.Split(value, ",")
			for i, interest := range interests {
				interests[i] = strings.TrimSpace(interest)
			}

			interestsJSON, err := json.Marshal(interests)
			if err != nil {
				return fmt.Errorf("转换 interests 为 JSON 字符串失败: %v", err)
			}

			query = "UPDATE users SET interests = ? WHERE user_id = ?"
			value = string(interestsJSON)
		case "persign":
			query = "UPDATE users SET signature = ? WHERE user_id = ?"
		default:
			return fmt.Errorf("无效的更新类型: %s", fieldType)
		}
	}

	_, err = db.Exec(query, value, uid)
	if err != nil {
		return fmt.Errorf("更新失败: %v", err)
	}
	return nil
}

// 处理更新用户某个字段的请求
func UpdatePersonalInfo(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "数据库连接失败"})
		return
	}
	defer db.Close()
	var request struct {
		Type  string `json:"type"`
		Value string `json:"value"`
		Uid   string `json:"uid"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "请求参数错误"})
		return
	}

	if request.Type == "" || request.Value == "" || request.Uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "缺少必需的参数"})
		return
	}

	err = UpdatePersonal(db, request.Uid, request.Type, request.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isok":       true,
		"failreason": "",
	})
}

// 定义粉丝结构体类型
type Follower struct {
	FollowerID int    `json:"follower_id"`
	Uid        string `json:"uid"`
	Avatar     string `json:"uimage"`
	Uname      string `json:"uname"`
	IsAttion   bool   `json:"isattion"`
}

// 定义关注结构体类型
type Following struct {
	FollowedID int    `json:"followed_id"`
	Uid        string `json:"uid"`
	Avatar     string `json:"uimage"`
	Uname      string `json:"uname"`
}

// 查粉丝
func GetFollowers(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	uid := c.Query("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "缺少用户ID"})
		return
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ?)", uid).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "检查用户是否存在时发生错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "用户不存在"})
		return
	}

	page := c.DefaultQuery("page", "1")
	pageSize := 5
	currentPage, _ := strconv.Atoi(page)
	offset := (currentPage - 1) * pageSize

	var followers []Follower

	query := `
   SELECT u.user_id, u.avatar, u.uname
   FROM userfollows uf
   JOIN users u ON uf.follower_id = u.user_id
   WHERE uf.followed_id = ?
   LIMIT ? OFFSET ?
`
	rows, err := db.Query(query, uid, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "查询粉丝失败"})
		return
	}
	defer rows.Close()

	// 获取当前用户的关注状态
	for rows.Next() {
		var follower Follower
		if err := rows.Scan(&follower.FollowerID, &follower.Avatar, &follower.Uname); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"failreason": "读取粉丝数据失败"})
			return
		}
		follower.Uid = strconv.Itoa(follower.FollowerID)

		// 判断当前用户是否关注该粉丝
		var isFollowing bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM userfollows WHERE follower_id = ? AND followed_id = ?)", uid, follower.FollowerID).Scan(&isFollowing)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"failreason": "检查关注关系时发生错误"})
			return
		}
		follower.IsAttion = isFollowing // 动态设置是否关注

		if follower.Avatar == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"failreason": "头像地址为空"})
			return
		}
		err, signedAvatarUrl := scripts.GetUrl(follower.Avatar)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"failreason": "获取头像签名失败"})
			return
		}
		follower.Avatar = signedAvatarUrl
		followers = append(followers, follower)
	}

	var totalFollowers int
	queryTotal := `SELECT COUNT(*) FROM userfollows WHERE followed_id = ?`
	err = db.QueryRow(queryTotal, uid).Scan(&totalFollowers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "获取粉丝总数失败"})
		return
	}

	totalPages := int(math.Ceil(float64(totalFollowers) / float64(pageSize)))

	c.JSON(http.StatusOK, gin.H{
		"datas":      followers,
		"totalPages": totalPages,
	})
}

// 查关注列表
func GetFollowing(c *gin.Context) {
	db, err := repository.Connect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "数据库连接失败"})
		return
	}
	defer db.Close()

	uid := c.Query("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "缺少用户ID"})
		return
	}
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ?)", uid).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "检查用户是否存在时发生错误"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "用户不存在"})
		return
	}
	page := c.DefaultQuery("page", "1") // 当前页码，默认为1

	if uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"failreason": "缺少用户ID"})
		return
	}

	pageSize := 5
	currentPage, _ := strconv.Atoi(page)
	offset := (currentPage - 1) * pageSize

	var followings []Following
	query := `
        SELECT u.user_id, u.avatar, u.uname
        FROM userfollows uf
        JOIN users u ON uf.followed_id = u.user_id
        WHERE uf.follower_id = ? 
        LIMIT ? OFFSET ?
    `
	rows, err := db.Query(query, uid, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "查询关注失败"})
		return
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
		var following Following
		if err := rows.Scan(&following.FollowedID, &following.Avatar, &following.Uname); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"failreason": "读取关注数据失败"})
			return
		}

		following.Uid = strconv.Itoa(following.FollowedID)
		err, signedAvatarUrl := scripts.GetUrl(following.Avatar)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"failreason": "获取头像签名失败"})
			return
		}
		following.Avatar = signedAvatarUrl
		followings = append(followings, following)
	}

	if count == 0 {
		c.JSON(http.StatusOK, gin.H{"failreason": "没有找到关注的数据"})
		return
	}

	var totalFollowings int
	queryTotal := `SELECT COUNT(*) FROM userfollows WHERE follower_id = ?`
	err = db.QueryRow(queryTotal, uid).Scan(&totalFollowings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"failreason": "获取关注总数失败"})
		return
	}

	totalPages := int(math.Ceil(float64(totalFollowings) / float64(pageSize)))

	c.JSON(http.StatusOK, gin.H{
		"datas":      followings,
		"totalPages": totalPages,
	})
}

// 修改邮箱
func ChangeEmail(c *gin.Context) {
	db, err_conn := repository.Connect()
	if err_conn != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "连接数据库失败"})
		return
	}
	var requestData model.ChangeEmail
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{"isok": false, "failreason": "绑定请求数据失败"})
		return
	}
	fmt.Println(requestData.Code)
	fmt.Println(requestData.NewMail)
	fmt.Println(requestData.Uid)
	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE user_id = ? or email = ?)", requestData.Uid, requestData.Uid).Scan(&userExists)
	if err != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": "检查用户是否存在时发生错误"})
		return
	}
	if !userExists {
		c.JSON(400, gin.H{"isok": false, "failreason": "用户不存在"})
		return
	}

	query := "SELECT code FROM verificationcodes WHERE email = ? AND expiration > NOW() ORDER BY expiration DESC LIMIT 1"
	row := db.QueryRow(query, requestData.NewMail)
	var code string
	err_check := row.Scan(&code)
	if err_check != nil || code != requestData.Code {
		c.JSON(400, gin.H{"isok": false, "failreason": "验证码错误"})
		return
	}
	isemail := isEmailFormat(requestData.NewMail)
	if !isemail {
		c.JSON(400, gin.H{"isok": false, "failreason": "邮箱格式错误"})
		return
	}

	err_re, result := updateEmail(db, requestData.Uid, requestData.NewMail)
	if err_re != nil {
		c.JSON(500, gin.H{"isok": false, "failreason": result})
		return
	}

	c.JSON(200, gin.H{
		"isok": true,
	})
}

// 更新邮箱
func updateEmail(db *sql.DB, uid string, newemail string) (error, string) {

	tx, err := db.Begin()
	if err != nil {
		return err, "开启事务失败"
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE users SET email = ? WHERE user_id = ?")
	if err != nil {
		return err, "准备更新语句失败"
	}
	defer stmt.Close()

	_, err = stmt.Exec(newemail, uid)
	if err != nil {
		return err, "更新密码失败"
	}

	err = tx.Commit()
	if err != nil {
		return err, "提交事务失败"
	}
	return nil, ""
}
