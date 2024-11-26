package scripts

import (
	"bytes"
	"encoding/base64"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// 上传图片，这个函数想要使用需要设置环境变量
func UploadImage(base64Str string, filename string) (error, string) {
	// 从环境变量中获取AK和SK
	provider, err_ecv := oss.NewEnvironmentVariableCredentialsProvider()
	if err_ecv != nil {
		return err_ecv, "get AK and SK error"
	}
	// 创建OSSClient实例
	clientOptions := []oss.ClientOption{oss.SetCredentialsProvider(&provider)}
	clientOptions = append(clientOptions, oss.Region("cn-shenzhen"))
	clientOptions = append(clientOptions, oss.AuthVersion(oss.AuthV4))
	client, err_n := oss.New("https://oss-cn-shenzhen.aliyuncs.com", "", "", clientOptions...)
	if err_n != nil {
		return err_n, "Oss create client error"
	}
	//存储桶
	bucketName := "middleproject"
	bucket, err_b := client.Bucket(bucketName)
	if err_b != nil {
		return err_b, "Oss get bucket error"
	}
	//定义文件路径
	objectKey := "postImage/" + filename
	//将前端传送的base64字符串解码成bytes数组
	data, err := base64.StdEncoding.DecodeString(base64Str)
	reader := bytes.NewReader(data)
	if err != nil {
		return err, "oss base64 decode error"
	}
	//上传文件
	err = bucket.PutObject(objectKey, reader)
	return nil, objectKey
}

func GetUrl(filename string) string {
	return "test"
}
