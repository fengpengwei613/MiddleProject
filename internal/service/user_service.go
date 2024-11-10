package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"middleproject/internal/model"
	"middleproject/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

// register 函数实现
func register(info json.RawMessage) (bool, string, int) {
	// 数据库连接字符串
	db, err = repository.connect()
	
	if err != nil {
		return false, "数据库连接失败", 0
	}
	defer db.Close()
	
	// 解析 JSON 数据
	var user model.User
	err = json.Unmarshal(info, &user)
	if err != nil {
		return false, "无效的 JSON 数据", 0
	}
	// 检查必需字段
	if user.Uname == "" || user.Password == "" || user.Email == "" {
		return false, "缺少必需字段", 0
	}
	// 创建用户
	err,_,_ = user.CreateUser()
	if err != nil {
		log.Println("用户创建失败:", err)
		return false, "注册失败", 0
	}

	return true, "注册成功", user.UserID
}