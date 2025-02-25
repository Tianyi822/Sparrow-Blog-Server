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
	//   image_oss_path: images/ # 图片在 OSS 上的路径
	//   avatar_oss_path: images/avatar/ # 头像在 OSS 上的路径
	//   blog_oss_path: blogs/ # 博客在 OSS 上的路径
	//   webp: # webp 配置
	//     enable: true # 是否开启 webp 功能
	//     quality: 75 # webp 压缩质量
	//     size: 1 # 压缩后的大小，单位 MB
	configServerGroup.POST("/user", configUser)

	configServerGroup.GET("/shutdown", closeConfigServer)
}
