package initrouter

import (
	"github.com/gin-gonic/gin"
)

func Routers(e *gin.Engine) {
	// 配置服务统一接口
	initGroup := e.Group("/init")

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
	initGroup.POST("/server", configServer)

	// 配置邮箱并发送验证码，并且将传入的用户配置参数先保存到全局中，必须在前端中保证只有验证码通过后，才能发起配置 User 的请求
	initGroup.POST("/config-email-send-code", configEmailAndSendCode)

	// 配置用户服务，其配置项如下，保存在 config.yaml 中：
	//
	// user:
	//  # 用户名
	//  username: chentyit
	//  # 用户邮箱
	//  user_email: chentyit@163.com
	//  # 邮箱 SMTP 账号
	//  smtp_account: chentyit@163.com
	//  # 邮箱 SMTP 服务器地址
	//  smtp_address: smtp.163.com
	//  # 邮箱 SMTP 端口
	//  smtp_port: 465
	//  # 邮箱 SMTP 密码
	//  smtp_auth_code: YThfU32Tcq3FdVvx
	initGroup.POST("/user", configUser)

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
	initGroup.POST("/mysql", configMysql)

	// 配置 OSS，其配置项如下，保存在 config.yaml 中：
	//
	// ossstore:
	//   # OSS endpoint
	//   endpoint: ossstore-xx-xxxxxx.xxxxxx.com
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
	initGroup.POST("/oss", configOss)

	// 配置缓存，其配置项如下，保存在 config.yaml 中：
	//
	// cache:
	//   # AOF 配置
	//   aof:
	//     # 是否开启AOF
	//     enable: true
	//     # AOF 文件路径
	//     path: ../temp/h2blog.aof
	//     # AOF 文件最大大小，单位-MB
	//     max_size: 1
	//     # 是否开启压缩
	//     compress: true
	initGroup.POST("/cache", configCache)

	// 配置日志，其配置项如下，保存在 config.yaml 中：
	//
	// logger:
	//  # 日志级别
	//  level: debug
	//  # 日志文件路径
	//  path: temp/h2blog.log
	//  # 日志文件最大大小，单位-MB
	//  max_size: 3
	//  # 日志文件最大备份数量
	//  max_backups: 30
	//  # 日志文件最大保存时间，单位-天
	//  max_age: 7
	//  # 是否压缩日志文件
	//  compress: true
	initGroup.POST("/logger", configLogger)

	// 完成配置并保存接口
	initGroup.GET("/complete-config", completeConfig)
}
