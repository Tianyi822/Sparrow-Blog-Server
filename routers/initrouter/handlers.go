package initrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/email"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"
	"path/filepath"
	"strings"
)

// initServer 用于初始化服务器配置
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func initServer(ctx *gin.Context) {
	serverConfig := config.ServerConfigData{}

	// 配置跨域相关的固定值
	serverConfig.Cors.Headers = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	serverConfig.Cors.Methods = []string{"POST", "PUT", "DELETE", "GET"}

	// 解析请求数据
	mapData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	// 解析端口配置
	port, err := tools.AnalyzePort(mapData["server.port"].(string))
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err.Error())
		return
	}
	serverConfig.Port = port

	// 解析Token密钥
	tokenKey := strings.TrimSpace(mapData["server.token_key"].(string))
	if err = tools.AnalyzeTokenKey(tokenKey); err != nil {
		resp.BadRequest(ctx, "Token密钥配置错误", err.Error())
		return
	}
	serverConfig.TokenKey = tokenKey

	// 解析Token过期时间
	tokenExpireDuration := strings.TrimSpace(mapData["server.token_expire_duration"].(string))
	dur, err := tools.AnalyzeTokenExpireDuration(tokenExpireDuration)
	if err != nil {
		resp.BadRequest(ctx, "Token过期时间配置错误", err.Error())
		return
	}
	serverConfig.TokenExpireDuration = dur

	// 解析跨域源配置
	domain := mapData["server.cors.origins"].(string)
	corsOrigins := []string{
		"http://" + domain,
		"http://www." + domain,
		"https://" + domain,
		"https://www." + domain,
	}
	if err = tools.AnalyzeCorsOrigins(corsOrigins); err != nil {
		resp.BadRequest(ctx, "跨域源配置错误", err.Error())
		return
	}
	serverConfig.Cors.Origins = corsOrigins

	// 保存配置
	config.Server = serverConfig

	resp.Ok(ctx, "配置完成", config.Server)
}

// initUser 用于初始化用户配置
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func initUser(ctx *gin.Context) {
	// 解析请求数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}

	userConfig := config.UserConfigData{}

	// 解析用户名
	username, getErr := tools.GetStringFromRawData(rawData, "user.username")
	if getErr != nil {
		resp.BadRequest(ctx, "用户名配置错误", getErr.Error())
		return
	}
	if username == "" {
		resp.BadRequest(ctx, "用户名不能为空", nil)
		return
	}
	userConfig.Username = username

	// 解析用户邮箱
	userEmail, getErr := tools.GetStringFromRawData(rawData, "user.user_email")
	if getErr != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", getErr.Error())
		return
	}
	if anaErr := tools.AnalyzeEmail(userEmail); anaErr != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", anaErr.Error())
		return
	}
	userConfig.UserEmail = userEmail

	// 解析GitHub地址
	githubAddress, getErr := tools.GetStringFromRawData(rawData, "user.user_github_address")
	if getErr != nil {
		resp.BadRequest(ctx, "GitHub地址配置错误", getErr.Error())
		return
	}
	userConfig.UserGithubAddress = githubAddress

	// 解析用户爱好
	userHobbies, getErr := tools.GetStrListFromRawData(rawData, "user.user_hobbies")
	if getErr != nil {
		resp.BadRequest(ctx, "爱好配置错误", getErr.Error())
		return
	}
	userConfig.UserHobbies = userHobbies

	// 解析打字机内容
	typeWriterContent, getErr := tools.GetStrListFromRawData(rawData, "user.type_writer_content")
	if getErr != nil {
		resp.BadRequest(ctx, "打字机内容配置错误", getErr.Error())
		return
	}
	userConfig.TypeWriterContent = typeWriterContent

	// 保存配置
	config.User = userConfig

	resp.Ok(ctx, "配置完成", config.User)
}

// sendCode 用于发送验证码
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func sendCode(ctx *gin.Context) {
	// 解析请求数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	// 解析用户邮箱
	userEmail := strings.TrimSpace(rawData["user.user_email"].(string))
	if emailErr := tools.AnalyzeEmail(userEmail); emailErr != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", emailErr.Error())
		return
	}

	// 解析SMTP账户
	smtpAccount := strings.TrimSpace(rawData["user.smtp_account"].(string))
	if emailErr := tools.AnalyzeEmail(smtpAccount); emailErr != nil {
		resp.BadRequest(ctx, "系统邮箱配置错误", emailErr.Error())
		return
	}

	// 解析SMTP服务器配置
	smtpAddress := strings.TrimSpace(rawData["user.smtp_address"].(string))
	smtpPort, err := tools.GetUInt16FromRawData(rawData, "user.smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "系统邮箱端口配置错误", err.Error())
		return
	}
	smtpAuthCode := strings.TrimSpace(rawData["user.smtp_auth_code"].(string))

	// 发送验证码
	if err = email.SendVerificationCodeByArgs(ctx, userEmail, smtpAccount, smtpAddress, smtpAuthCode, smtpPort); err != nil {
		resp.BadRequest(ctx, "验证码发送失败", err.Error())
		return
	}

	resp.Ok(ctx, "验证码发送成功", config.User)
}

// initMysql 用于初始化MySQL配置
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func initMysql(ctx *gin.Context) {
	mysqlConfig := config.MySQLConfigData{}

	// 解析请求数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "请求数据解析错误", err.Error())
		return
	}

	// 解析基本配置
	mysqlConfig.User = strings.TrimSpace(rawData["mysql.user"].(string))
	mysqlConfig.Password = strings.TrimSpace(rawData["mysql.password"].(string))

	// 解析主机地址
	host := strings.TrimSpace(rawData["mysql.host"].(string))
	if err = tools.AnalyzeHostAddress(host); err != nil {
		resp.BadRequest(ctx, "数据库主机地址配置错误", err.Error())
		return
	}
	mysqlConfig.Host = host

	// 解析端口
	port, err := tools.GetUInt16FromRawData(rawData, "mysql.port")
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err.Error())
		return
	}
	mysqlConfig.Port = port

	// 解析数据库名
	mysqlConfig.DB = strings.TrimSpace(rawData["mysql.database"].(string))

	// 解析连接池配置
	maxOpen, err := tools.GetUInt16FromRawData(rawData, "mysql.max_open")
	if err != nil {
		resp.BadRequest(ctx, "最大打开连接数配置错误", err.Error())
		return
	}
	mysqlConfig.MaxOpen = maxOpen

	maxIdle, err := tools.GetUInt16FromRawData(rawData, "mysql.max_idle")
	if err != nil {
		resp.BadRequest(ctx, "最大空闲连接数配置错误", err.Error())
		return
	}
	mysqlConfig.MaxIdle = maxIdle

	// 验证连接池配置
	if maxIdle > maxOpen {
		resp.BadRequest(ctx, "最大空闲连接数不能大于最大打开连接数", nil)
		return
	}

	// 验证连接配置
	if err = tools.AnalyzeMySqlConnect(&mysqlConfig); err != nil {
		resp.BadRequest(ctx, "数据库连接配置错误", err.Error())
		return
	}

	// 保存配置
	config.MySQL = mysqlConfig

	resp.Ok(ctx, "配置完成", config.MySQL)
}

// initOss 用于初始化OSS配置
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func initOss(ctx *gin.Context) {
	ossConfig := config.OssConfig{}

	// 解析请求数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}

	// 解析基本配置
	ossConfig.Endpoint = strings.TrimSpace(rawData["oss.endpoint"].(string))
	ossConfig.Region = strings.TrimSpace(rawData["oss.region"].(string))
	ossConfig.AccessKeyId = strings.TrimSpace(rawData["oss.access_key_id"].(string))
	ossConfig.AccessKeySecret = strings.TrimSpace(rawData["oss.access_key_secret"].(string))
	ossConfig.Bucket = strings.TrimSpace(rawData["oss.bucket"].(string))
	if err = tools.AnalyzeOssConfig(&ossConfig); err != nil {
		resp.BadRequest(ctx, "OSS配置错误", err.Error())
		return
	}

	// 解析路径配置
	imageOssPath := strings.TrimSpace(rawData["oss.image_oss_path"].(string))
	if err := tools.AnalyzeOssPath(imageOssPath); err != nil {
		resp.BadRequest(ctx, "图片OSS路径配置错误", err.Error())
		return
	}
	ossConfig.ImageOssPath = imageOssPath

	blogOssPath := strings.TrimSpace(rawData["oss.blog_oss_path"].(string))
	if err := tools.AnalyzeOssPath(blogOssPath); err != nil {
		resp.BadRequest(ctx, "博客OSS路径配置错误", err.Error())
		return
	}
	ossConfig.BlogOssPath = blogOssPath

	// 保存配置
	config.Oss = ossConfig

	resp.Ok(ctx, "配置完成", config.Oss)
}

// initCache 用于初始化缓存配置
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func initCache(ctx *gin.Context) {
	cacheConfig := config.CacheConfig{}

	// 解析请求数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "请求数据解析错误", err.Error())
		return
	}

	// 解析AOF配置
	aofEnable, err := tools.GetBoolFromRawData(rawData, "cache.aof.enable")
	if err != nil {
		resp.BadRequest(ctx, "AOF启用配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Enable = aofEnable

	// 解析AOF路径
	projPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(rawData["cache.aof.path"].(string)))
	if err != nil {
		resp.BadRequest(ctx, "AOF路径配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Path = filepath.Join(projPath, "aof", "h2blog.aof")

	// 解析AOF文件配置
	maxSize, err := tools.GetUInt16FromRawData(rawData, "cache.aof.max_size")
	if err != nil {
		resp.BadRequest(ctx, "AOF最大文件大小配置错误", err.Error())
		return
	}
	cacheConfig.Aof.MaxSize = maxSize

	aofCompress, err := tools.GetBoolFromRawData(rawData, "cache.aof.compress")
	if err != nil {
		resp.BadRequest(ctx, "AOF文件压缩配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Compress = aofCompress

	// 保存配置
	config.Cache = cacheConfig

	resp.Ok(ctx, "配置完成", config.Cache)
}

// initLogger 用于初始化日志配置
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func initLogger(ctx *gin.Context) {
	loggerConfig := config.LoggerConfigData{}

	// 解析请求数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	// 解析日志级别
	level := strings.TrimSpace(rawData["logger.level"].(string))
	if anaErr := tools.AnalyzeLoggerLevel(level); anaErr != nil {
		resp.BadRequest(ctx, "日志级别配置错误", anaErr.Error())
		return
	}
	loggerConfig.Level = level

	// 解析日志路径
	projPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(rawData["logger.path"].(string)))
	if err != nil {
		resp.BadRequest(ctx, "日志路径配置错误", err.Error())
		return
	}
	loggerConfig.Path = filepath.Join(projPath, "log", "h2blog.log")

	// 解析日志文件配置
	maxSize, err := tools.GetUInt16FromRawData(rawData, "logger.max_size")
	if err != nil {
		resp.BadRequest(ctx, "日志最大文件大小配置错误", err.Error())
		return
	}
	loggerConfig.MaxSize = maxSize

	maxBackups, err := tools.GetUInt16FromRawData(rawData, "logger.max_backups")
	if err != nil {
		resp.BadRequest(ctx, "日志最大备份数量配置错误", err.Error())
		return
	}
	loggerConfig.MaxBackups = maxBackups

	maxAge, err := tools.GetUInt16FromRawData(rawData, "logger.max_age")
	if err != nil {
		resp.BadRequest(ctx, "日志最大保存天数配置错误", err.Error())
		return
	}
	loggerConfig.MaxAge = maxAge

	compress, err := tools.GetBoolFromRawData(rawData, "logger.compress")
	if err != nil {
		resp.BadRequest(ctx, "日志文件压缩配置错误", err.Error())
		return
	}
	loggerConfig.Compress = compress

	// 保存配置
	config.Logger = loggerConfig

	resp.Ok(ctx, "配置完成", config.Logger)
}

// completeInit 用于完成所有配置并保存
// 参数:
//   - ctx: *gin.Context Gin框架的上下文对象
func completeInit(ctx *gin.Context) {
	// 整合所有配置
	projConfig := config.ProjectConfig{
		User:   config.User,
		Server: config.Server,
		MySQL:  config.MySQL,
		Oss:    config.Oss,
		Cache:  config.Cache,
		Logger: config.Logger,
	}

	// 保存配置到文件
	err := projConfig.Store()
	if err != nil {
		resp.Err(ctx, "配置保存失败", err.Error())
		return
	}

	resp.Ok(ctx, "完成并保存配置", nil)

	// 发送完成信号
	if env.CurrentEnv == env.InitializedEnv {
		env.CompletedConfigSign <- true
	}
}
