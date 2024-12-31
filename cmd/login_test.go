package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"middleproject/internal/repository"
	"middleproject/internal/service"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func server() {
	// 关闭gin的日志输出
	gin.SetMode(gin.ReleaseMode)
	// 数据库的日志输出在DBStruct/init.go中使用语句db.LogMode()配置，已改为false
	// 打开数据库
	repository.Connect()
	// 在这里配置一个路由器并运行，配置的路由器时不加token中间件，这样发送的请求不需要带token
	r := gin.Default()
	r.POST("api/login",service.Login)
	r.Run(":8080")
}

//测试登录
func TestLogin(t*testing.T) {
	go server()
    fmt.Println("服务器已启动")
	time.Sleep(time.Second * 2)
	url:="http://localhost:8080/api/login"
	data:=make(map[string]interface{})
	data["uid"]="20"
	data["password"]="123"
	byteData,err:=json.Marshal(data)
	if err!=nil{
	    t.Fatalf("转化为JSON格式错误,err:%v",err)
	}
	request,err:=http.NewRequest("POST",url,bytes.NewBuffer(byteData))
	if err!=nil{
	    t.Fatalf("创建请求错误,err:%v",err)
	}
	request.Header.Add("Content-Type","application/json")

	//发送请求体，获取响应
	client:=&http.Client{}
	response,err:=client.Do(request)
	if err!=nil{
	    t.Fatalf("发送请求错误,err:%v",err)
	}
	defer response.Body.Close()

    t.Log("Received response") // 添加调试日志
	body,err:=ioutil.ReadAll(response.Body)
	if err!=nil{
	    t.Fatalf("读取响应错误,err:%v",err)
	}

	var formatterBody bytes.Buffer
	err=json.Indent(&formatterBody,body,"","  ")
	if err!=nil{
	    t.Fatalf("格式化JSON错误,err:%v",err)
	}
	t.Log(formatterBody.String())
}