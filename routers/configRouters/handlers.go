package configRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/email"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"
	"h2blog_server/storage"
	"path/filepath"
	"strings"
)

// configBase 是一个用于解析和配置服务器基础信息的函数。
// 参数:
//
//	ctx *gin.Context - Gin 框架的上下文对象，用于处理 HTTP 请求和响应。
//
// 返回值:
//
//	无返回值，但通过 ctx 返回 JSON 格式的响应结果。
func configBase(ctx *gin.Context) {
	serverConfig := config.ServerConfigData{}

	// 配置跨域相关的固定值，这些值不需要前端传入，直接在代码中写死。
	serverConfig.Cors.Headers = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	serverConfig.Cors.Methods = []string{"POST", "PUT", "DELETE", "GET"}

	// 从请求的原始数据中解析出配置信息，并存储为 map 格式。
	mapData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	// 解析服务器端口配置，确保端口号合法。
	port, err := tools.AnalyzePort(mapData["server.port"].(string))
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err.Error())
		return
	}
	serverConfig.Port = port

	// 解析 Token 密钥配置，确保密钥符合要求。
	tokenKey := strings.TrimSpace(mapData["server.token_key"].(string))
	if err = tools.AnalyzeTokenKey(tokenKey); err != nil {
		resp.BadRequest(ctx, "Token 密钥配置错误", err.Error())
		return
	}
	serverConfig.TokenKey = tokenKey

	// 解析 Token 过期时间配置，确保时间格式正确并转换为有效的时间间隔。
	tokenExpireDuration := strings.TrimSpace(mapData["server.token_expire_duration"].(string))
	dur, err := tools.AnalyzeTokenExpireDuration(tokenExpireDuration)
	if err != nil {
		resp.BadRequest(ctx, "Token 过期时间配置错误", err.Error())
		return
	}
	serverConfig.TokenExpireDuration = dur

	// 解析跨域源配置，生成完整的跨域源地址列表并验证其合法性。
	domain := mapData["server.cors.origins"].(string)
	corsOrigins := []string{
		"https://" + domain,
		"https://www." + domain,
	}
	if err = tools.AnalyzeCorsOrigins(corsOrigins); err != nil {
		resp.BadRequest(ctx, "跨域源配置错误", err.Error())
		return
	}
	serverConfig.Cors.Origins = corsOrigins

	// 将解析完成的服务器配置存储到全局变量中，供后续使用。
	config.Server = serverConfig

	// 返回成功响应，包含配置完成的信息和最终的服务器配置。
	resp.Ok(ctx, "配置完成", config.Server)
}

// configUser 配置用户信息，包括验证验证码和设置用户名。
// 该函数根据当前环境（配置服务器环境或运行时环境）来决定如何处理验证码的验证与清除。
// 参数:
//   - ctx: *gin.Context, Gin框架的上下文对象，包含了请求和响应的信息。
//
// 返回值:
//
//	无直接返回值，通过ctx对象向客户端发送响应。
func configUser(ctx *gin.Context) {
	code := strings.TrimSpace(ctx.PostForm("user.verification_code"))

	// 根据当前环境验证验证码
	switch env.CurrentEnv {
	case env.ConfigServerEnv:
		if env.VerificationCode != code {
			resp.BadRequest(ctx, "验证码错误", nil)
			return
		}
	case env.RuntimeEnv:
		c, err := storage.Storage.Cache.GetString(ctx, "verification-code")
		if err != nil {
			resp.BadRequest(ctx, "验证码过期", err.Error())
			return
		}
		if c != code {
			resp.BadRequest(ctx, "验证码错误", nil)
			return
		}
	}

	// 设置用户名称
	config.User.Username = strings.TrimSpace(ctx.PostForm("user.username"))

	// 清除已使用的验证码
	switch env.CurrentEnv {
	case env.ConfigServerEnv:
		env.VerificationCode = ""
	case env.RuntimeEnv:
		err := storage.Storage.Cache.Delete(ctx, "verification-code")
		if err != nil {
			resp.Err(ctx, "验证码缓存清除失败", err.Error())
			return
		}
	}

	// 向客户端返回成功消息及配置后的用户信息
	resp.Ok(ctx, "配置完成", config.User)
}

// sendVerificationCode 处理发送验证码的请求。
// 该函数从请求中获取用户邮箱、SMTP账户等信息，验证这些信息的有效性，并将有效的配置保存到全局配置中。
// 最后，通过电子邮件发送验证码。如果过程中出现任何错误，将返回相应的错误信息。
// 参数:
//   - ctx: *gin.Context, 用于处理HTTP请求的上下文。
//
// 返回值:
//
//	无直接返回值，但会通过ctx对象响应客户端。
func sendVerificationCode(ctx *gin.Context) {
	userConfig := config.UserConfigData{}

	// 从请求中获取并验证用户邮箱
	userEmail := strings.TrimSpace(ctx.PostForm("user.user_email"))
	if err := tools.AnalyzeEmail(userEmail); err != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", err.Error())
		return
	}
	userConfig.UserEmail = userEmail

	// 从请求中获取并验证系统邮箱
	smtpAccount := strings.TrimSpace(ctx.PostForm("user.smtp_account"))
	if err := tools.AnalyzeEmail(smtpAccount); err != nil {
		resp.BadRequest(ctx, "系统邮箱配置错误", err.Error())
		return
	}
	userConfig.SmtpAccount = smtpAccount

	// 获取SMTP服务器地址
	userConfig.SmtpAddress = strings.TrimSpace(ctx.PostForm("user.smtp_address"))

	// 从请求中获取并验证SMTP端口
	smtpPort, err := tools.GetIntFromPostForm(ctx, "user.smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "系统邮箱端口配置错误", err.Error())
		return
	}
	userConfig.SmtpPort = smtpPort

	// 获取SMTP授权码
	userConfig.SmtpAuthCode = strings.TrimSpace(ctx.PostForm("user.smtp_auth_code"))

	// 发送验证码邮件
	if err = email.SendVerificationCodeEmail(ctx, userConfig.UserEmail); err != nil {
		resp.BadRequest(ctx, "验证码发送失败", err.Error())
		return
	}

	// 将验证后的配置添加到全局配置
	config.User = userConfig

	// 响应客户端验证码发送成功
	resp.Ok(ctx, "验证码发送成功", config.User)
}

func configMysql(ctx *gin.Context) {
	mysqlConfig := config.MySQLConfigData{}

	mysqlConfig.User = strings.TrimSpace(ctx.PostForm("mysql.user"))
	mysqlConfig.Password = strings.TrimSpace(ctx.PostForm("mysql.password"))

	host := strings.TrimSpace(ctx.PostForm("mysql.host"))
	if err := tools.AnalyzeHostAddress(host); err != nil {
		resp.BadRequest(ctx, "数据库主机地址配置错误", err.Error())
		return
	}
	mysqlConfig.Host = host

	// 解析端口
	port, err := tools.GetUInt16FromPostForm(ctx, "mysql.port")
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err.Error())
		return
	}
	mysqlConfig.Port = port

	mysqlConfig.DB = strings.TrimSpace(ctx.PostForm("mysql.database"))

	// 解析最大打开连接数
	maxOpen, err := tools.GetUInt16FromPostForm(ctx, "mysql.max_open")
	if err != nil {
		resp.BadRequest(ctx, "最大打开连接数配置错误", err.Error())
		return
	}
	mysqlConfig.MaxOpen = maxOpen

	// 解析最大空闲连接数
	maxIdle, err := tools.GetUInt16FromPostForm(ctx, "mysql.max_idle")
	if err != nil {
		resp.BadRequest(ctx, "最大空闲连接数配置错误", err.Error())
		return
	}
	mysqlConfig.MaxIdle = maxIdle

	// 检查连接数和空闲数
	if maxIdle > maxOpen {
		resp.BadRequest(ctx, "最大空闲连接数不能大于最大打开连接数", nil)
		return
	}

	if err = tools.AnalyzeMySqlConnect(&mysqlConfig); err != nil {
		resp.BadRequest(ctx, "数据库连接配置错误", err.Error())
		return
	}

	// 完成配置，将配置添加到全局
	config.MySQL = mysqlConfig

	resp.Ok(ctx, "配置完成", config.MySQL)
}

func configOss(ctx *gin.Context) {
	ossConfig := config.OssConfig{}

	// OSS 基础配置
	ossConfig.Endpoint = strings.TrimSpace(ctx.PostForm("oss.endpoint"))
	ossConfig.Region = strings.TrimSpace(ctx.PostForm("oss.region"))
	ossConfig.AccessKeyId = strings.TrimSpace(ctx.PostForm("oss.access_key_id"))
	ossConfig.AccessKeySecret = strings.TrimSpace(ctx.PostForm("oss.access_key_secret"))
	ossConfig.Bucket = strings.TrimSpace(ctx.PostForm("oss.bucket"))
	if err := tools.AnalyzeOssConfig(&ossConfig); err != nil {
		resp.BadRequest(ctx, "OSS 配置错误", err.Error())
		return
	}

	// OSS 路径配置
	imageOssPath := strings.TrimSpace(ctx.PostForm("oss.image_oss_path"))
	if err := tools.AnalyzeOssPath(imageOssPath); err != nil {
		resp.BadRequest(ctx, "图片 OSS 路径配置错误", err.Error())
		return
	}
	ossConfig.ImageOssPath = imageOssPath

	blogOssPath := strings.TrimSpace(ctx.PostForm("oss.blog_oss_path"))
	if err := tools.AnalyzeOssPath(blogOssPath); err != nil {
		resp.BadRequest(ctx, "博客 OSS 路径配置错误", err.Error())
		return
	}
	ossConfig.BlogOssPath = blogOssPath

	// OSS 下的 webp 文件配置
	webpEnable, err := tools.GetIntFromPostForm(ctx, "oss.webp.enable")
	if err != nil {
		resp.BadRequest(ctx, "WebP 启用配置错误", err.Error())
		return
	}
	ossConfig.WebP.Enable = webpEnable == 1

	webpQuality, err := tools.GetFloatFromPostForm(ctx, "oss.webp.quality")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩质量配置错误", err.Error())
		return
	}
	ossConfig.WebP.Quality = webpQuality

	webpSize, err := tools.GetFloatFromPostForm(ctx, "oss.webp.size")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩后大小配置错误", err.Error())
		return
	}
	ossConfig.WebP.Size = webpSize

	// 完成配置，并将配置添加到全局
	config.Oss = ossConfig

	resp.Ok(ctx, "配置完成", config.Oss)
}

func configCache(ctx *gin.Context) {
	cacheConfig := config.CacheConfig{}

	aofEnable, err := tools.GetIntFromPostForm(ctx, "cache.aof.enable")
	if err != nil {
		resp.BadRequest(ctx, "AOF 启用配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Enable = aofEnable == 1

	aofPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(ctx.PostForm("cache.aof.path")))
	if err != nil {
		resp.BadRequest(ctx, "AOF 路径配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Path = filepath.Join(aofPath, "aof.log")

	// 解析最大 AOF 文件大小
	maxSize, err := tools.GetUInt16FromPostForm(ctx, "cache.aof.max_size")
	if err != nil {
		resp.BadRequest(ctx, "AOF 最大文件大小配置错误", err.Error())
		return
	}
	cacheConfig.Aof.MaxSize = maxSize

	// 解析 AOF 文件压缩
	aofCompress, err := tools.GetIntFromPostForm(ctx, "cache.aof.compress")
	if err != nil {
		resp.BadRequest(ctx, "AOF 文件压缩配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Compress = aofCompress == 1

	// 完成配置，将配置添加到全局
	config.Cache = cacheConfig

	resp.Ok(ctx, "配置完成", config.Cache)
}

func configLogger(ctx *gin.Context) {
	loggerConfig := config.LoggerConfigData{}

	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	level := strings.TrimSpace(rawData["logger.level"].(string))
	if err := tools.AnalyzeLoggerLevel(level); err != nil {
		resp.BadRequest(ctx, "日志级别配置错误", err.Error())
		return
	}
	loggerConfig.Level = level

	projPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(rawData["logger.path"].(string)))
	if err != nil {
		resp.BadRequest(ctx, "日志路径配置错误", err.Error())
		return
	}
	loggerConfig.Path = filepath.Join(projPath, "log", "h2blog.log")

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

	compress, err := tools.GetIntFromRawData(rawData, "logger.compress")
	if err != nil {
		resp.BadRequest(ctx, "日志文件压缩配置错误", err.Error())
		return
	}
	loggerConfig.Compress = compress == 1

	// 完成配置，将配置添加到全局
	config.Logger = loggerConfig

	resp.Ok(ctx, "配置完成", config.Logger)
}

// completeConfig 完成配置，将配置保存到本地文件中
func completeConfig(ctx *gin.Context) {

	projConfig := config.ProjectConfig{
		User:   config.User,
		Server: config.Server,
		MySQL:  config.MySQL,
		Oss:    config.Oss,
		Cache:  config.Cache,
		Logger: config.Logger,
	}

	err := projConfig.Store()
	if err != nil {
		resp.Err(ctx, "配置保存失败", err.Error())
		return
	}

	resp.Ok(ctx, "关闭配置服务", nil)
	// 配置完成，通知关闭配置服务
	env.CompletedConfigSign <- true
}
