package configRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	// 配置服务统一接口
	configServerGroup := e.Group("/config-server")

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
	configServerGroup.POST("/base", configBase)

	// 配置用户服务，其配置项如下，保存在 config.yaml 中：
	//
	// user:
	//   username: name # 用户名
	//   email: mail@xx.com # 邮箱
	configServerGroup.POST("/user", configUser)

	// 配置 OSS，其配置项如下，保存在 config.yaml 中：
	//
	// oss:
	//   # OSS endpoint
	//   endpoint: oss-xx-xxxxxx.xxxxxx.com
	//   # OSS 地域
	//   region: cn-xxxxxx
	//   # OSS AccessKey ID
	//   access_key_id: xxxxxxxxxxxxxxxxxxxxxx
	//   # OSS AccessKey Secret
	//   access_key_secret: xxxxxxxxxxxxxxxxxxxxxx
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
	configServerGroup.POST("/oss", configOss)

	configServerGroup.GET("/shutdown", closeConfigServer)
}
