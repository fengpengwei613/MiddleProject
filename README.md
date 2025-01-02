# MiddleProject
后端配置：
1. 阿里云
图片存储在阿里云里，后端数据库只保留图片地址，处理图片数据时调用scripts/saveimage.go文件里的GetUrl函数，即可返回其对应的url值。阿里云相关配置如下：首先注册阿里云账号，然后开通OSS服务，创建有OSS管理权限的RAM用户AccessKey，在cmd中运行以下命令：
setx OSS_ACCESS_KEY_ID "YOUR_ACCESS_KEY_ID"
setx OSS_ACCESS_KEY_SECRET "YOUR_ACCESS_KEY_SECRET"
即可完成阿里云相关配置。
2. 数据库
本项目使用mysql数据库，具体代码详见internal/repository/connectdb.go文件，如需使用，只需更改dsn配置，将root更改为你的mysql用户名，将passward改为你实际的mysql密码，将blog_db改为你设置的数据库名称，127.0.0.1为数据库服务器的 IP 地址，3306为mysql数据库的默认端口号，可按需要更改。
3. 运行
运行后端代码需配置go的相关环境，在项目文件夹cmd目录下打开终端，使用“go run .”命令即可运行后端。
4. 测试
