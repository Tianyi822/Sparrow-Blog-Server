package configRouters

import "github.com/gin-gonic/gin"

func Routers(e *gin.Engine) {
	// 配置服务统一接口
	configServerGroup := e.Group("/config-server")

	// 配置基础服务，其配置项如下，保存在 config.yaml 中：
	//
	//	server: # 服务配置
	//	  port: 6666 # 服务端口
	//	  token_key: A9H0YhKZw=8GO&hJgVmNS # 生成 Token 的密钥
	//	  token_expire_duration: 30 # Token 过期时间，单位-天
	//	  cors: # 跨域配置
	//	    origins: # 跨域源
	//	      - http://localhost:3000
	//	    headers: # 跨域请求头
	//	      - Content-Type
	//	      - AccessToken
	//	      - X-CSRF-Token
	//	      - Authorization
	//	      - Token
	//	    methods: # 跨域请求方法
	//	      - POST
	//	      - PUT
	//	      - DELETE
	//	      - GET
	//	      - OPTIONS
	configServerGroup.POST("/base", configBase)

	configServerGroup.GET("/shutdown", closeConfigServer)
}
