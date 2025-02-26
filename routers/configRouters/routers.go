package configRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
)

func Routers(e *gin.Engine) {
	// 配置服务统一接口
	configGroup := e.Group("/config")

	// 基础配置接口，其配置项如下，保存在 config.yaml 中：
	//
	// server: # 服务配置
	//   port: 6666 # 服务端口
	//   token_key: A9H0YhKZw=8GO&hJgVmNS # 生成 Token 的密钥
	//   token_expire_duration: 30 # Token 过期时间，单位-天
	//   cors: # 跨域配置
	//     origins: # 跨域源
	//       - http://localhost:3000
	//     headers: # 跨域请求头
	//       - Content-Type
	//       - AccessToken
	//       - X-CSRF-Token
	//       - Authorization
	//       - Token
	//     methods: # 跨域请求方法
	//       - POST
	//       - PUT
	//       - DELETE
	//       - GET
	//       - OPTIONS
	configGroup.POST("/base", configBase)

	// 配置用户服务，其配置项如下，保存在 config.yaml 中：
	//
	// user:
	//   username: name # 用户名
	//   email: mail@xx.com # 邮箱
	configGroup.POST("/user", configUser)

	// 配置 MySQL 数据库
	//
	// # 定义MySQL数据库的连接配置
	// mysql:
	//   # 数据库用户名
	//   user: xxxxxx
	//   # 数据库密码
	//   password: xxxxxx
	//   # 数据库主机地址
	//   host: 127.0.0.1
	//   # 数据库端口号
	//   port: 3306
	//   # 数据库名称
	//   database: xxxxxx
	//   # 最大打开的数据库连接数
	//   max_open: 10
	//   # 最大空闲的数据库连接数
	//   max_idle: 5
	configGroup.POST("/mysql", configMysql)

	// 配置 OSS，其配置项如下，保存在 config.yaml 中：
	//
	// oss:
	//   # OSS endpoint
	//   endpoint: oss-xx-xxxxxx.xxxxxx.com
	//   # OSS 地域
	//   region: cn-xxxxxx
	//   # OSS AccessKey ID
	//   access_key_id: xxxxxx
	//   # OSS AccessKey Secret
	//   access_key_secret: xxxxxx
	//   # OSS bucket 名称
	//   bucket: xxx
	//   # 图片在 OSS 上的路径
	//   image_oss_path: images/
	//   # 头像在 OSS 上的路径
	//   avatar_oss_path: images/avatar/
	//   # 博客在 OSS 上的路径
	//   blog_oss_path: blogs/
	//   # webp 配置
	//   webp:
	//     # 是否开启 webp 功能
	//     enable: true
	//     # webp 压缩质量
	//     quality: 75
	//     # 压缩后的大小，单位 MB
	//     size: 1
	configGroup.POST("/oss", configOss)

	// 只有在 CONFIG_SERVER_ENV 环境下才允许关闭配置服务
	if env.CurrentEnv == env.ConfigServerEnv {
		configGroup.GET("/shutdown", closeConfigServer)
	}
}
