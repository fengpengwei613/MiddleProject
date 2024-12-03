package scripts

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// 上传图片，这个函数想要使用需要设置环境变量
func UploadImage(base64Str string, filename string) (error, string) {
	// 从环境变量中获取AK和SK
	provider, err_ecv := oss.NewEnvironmentVariableCredentialsProvider()
	if err_ecv != nil {
		return err_ecv, "Oss获取AK和SK失败"
	}
	// 创建OSSClient实例
	clientOptions := []oss.ClientOption{oss.SetCredentialsProvider(&provider)}
	clientOptions = append(clientOptions, oss.Region("cn-shenzhen"))
	clientOptions = append(clientOptions, oss.AuthVersion(oss.AuthV4))
	client, err_n := oss.New("https://oss-cn-shenzhen.aliyuncs.com", "", "", clientOptions...)
	if err_n != nil {
		return err_n, "Oss创建用户实例失败"
	}
	//存储桶
	bucketName := "middleproject"
	bucket, err_b := client.Bucket(bucketName)
	if err_b != nil {
		return err_b, "Oss获取存储桶失败"
	}
	//定义文件路径
	objectKey := "postImage/" + filename
	//将前端传送的base64字符串解码成bytes数组
	fmt.Println(base64Str)
	for i := 0; i < len(base64Str); i++ {
		if base64Str[i] == ',' {
			base64Str = base64Str[i+1:]
			break
		}
	}
	data, err_base64 := base64.StdEncoding.DecodeString(base64Str)
	reader := bytes.NewReader(data)
	if err_base64 != nil {
		return err_base64, "oss解码base64失败"
	}
	//上传文件
	err_upload := bucket.PutObject(objectKey, reader)
	if err_upload != nil {
		return err_upload, "oss上传失败"
	}
	return nil, objectKey
}

func GetUrl(filename string) (error, string) {
	provider, err_ecv := oss.NewEnvironmentVariableCredentialsProvider()
	if err_ecv != nil {
		return err_ecv, "Oss获取AK和SK失败"
	}
	// 创建OSSClient实例
	clientOptions := []oss.ClientOption{oss.SetCredentialsProvider(&provider)}
	clientOptions = append(clientOptions, oss.Region("cn-shenzhen"))
	clientOptions = append(clientOptions, oss.AuthVersion(oss.AuthV4))
	client, err_n := oss.New("https://oss-cn-shenzhen.aliyuncs.com", "", "", clientOptions...)
	if err_n != nil {
		return err_n, "Oss创建用户实例失败"
	}
	//存储桶
	bucketName := "middleproject"
	bucket, err_b := client.Bucket(bucketName)
	if err_b != nil {
		return err_b, "Oss获取存储桶失败"
	}
	signedURL, err_url := bucket.SignURL(filename, oss.HTTPGet, 1800)
	if err_url != nil {
		return err_url, "Oss获取Url失败"
	}

	return nil, signedURL
}
