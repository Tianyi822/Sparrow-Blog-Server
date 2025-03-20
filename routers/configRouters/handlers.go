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
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	code := strings.TrimSpace(rawData["user.verification_code"].(string))

	// 根据当前环境验证验证码
	switch env.CurrentEnv {
	case env.InitializedEnv:
		if env.VerificationCode != code {
			resp.BadRequest(ctx, "验证码错误", nil)
			return
		}
	case env.RuntimeEnv:
		c, err := storage.Storage.Cache.GetString(ctx, env.VerificationCodeKey)
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
	config.User.Username = strings.TrimSpace(rawData["user.username"].(string))

	// 清除已使用的验证码
	switch env.CurrentEnv {
	case env.InitializedEnv:
		env.VerificationCode = ""
	case env.RuntimeEnv:
		err := storage.Storage.Cache.Delete(ctx, env.VerificationCodeKey)
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

	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	// 从请求中获取并验证用户邮箱
	userEmail := strings.TrimSpace(rawData["user.user_email"].(string))
	if err := tools.AnalyzeEmail(userEmail); err != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", err.Error())
		return
	}
	userConfig.UserEmail = userEmail

	// 从请求中获取并验证系统邮箱
	smtpAccount := strings.TrimSpace(rawData["user.smtp_account"].(string))
	if err := tools.AnalyzeEmail(smtpAccount); err != nil {
		resp.BadRequest(ctx, "系统邮箱配置错误", err.Error())
		return
	}
	userConfig.SmtpAccount = smtpAccount

	// 获取SMTP服务器地址
	userConfig.SmtpAddress = strings.TrimSpace(rawData["user.smtp_address"].(string))

	// 从请求中获取并验证SMTP端口
	smtpPort, err := tools.GetUInt16FromRawData(rawData, "user.smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "系统邮箱端口配置错误", err.Error())
		return
	}
	userConfig.SmtpPort = smtpPort

	// 获取SMTP授权码
	userConfig.SmtpAuthCode = strings.TrimSpace(rawData["user.smtp_auth_code"].(string))

	// 将验证后的配置添加到全局配置
	config.User = userConfig

	// 发送验证码邮件
	if err = email.SendVerificationCodeEmail(ctx, userConfig.UserEmail); err != nil {
		resp.BadRequest(ctx, "验证码发送失败", err.Error())
		return
	}

	// 响应客户端验证码发送成功
	resp.Ok(ctx, "验证码发送成功", config.User)
}

// configMysql 配置MySQL数据库连接信息。
// 该函数从HTTP请求上下文中提取MySQL配置数据，验证并解析这些数据，然后更新全局MySQL配置。
// 参数:
//
//	ctx *gin.Context: HTTP请求上下文，用于处理响应和获取请求数据。
func configMysql(ctx *gin.Context) {
	// 初始化MySQL配置结构体。
	mysqlConfig := config.MySQLConfigData{}

	// 从请求中获取原始数据并解析为map形式。
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		// 如果数据解析失败，返回400错误响应。
		resp.BadRequest(ctx, "请求数据解析错误", err.Error())
		return
	}

	// 从原始数据中提取并修剪MySQL用户和密码。
	mysqlConfig.User = strings.TrimSpace(rawData["mysql.user"].(string))
	mysqlConfig.Password = strings.TrimSpace(rawData["mysql.password"].(string))

	// 提取并验证MySQL主机地址。
	host := strings.TrimSpace(rawData["mysql.host"].(string))
	if err = tools.AnalyzeHostAddress(host); err != nil {
		// 如果主机地址配置错误，返回400错误响应。
		resp.BadRequest(ctx, "数据库主机地址配置错误", err.Error())
		return
	}
	mysqlConfig.Host = host

	// 解析MySQL端口。
	port, err := tools.GetUInt16FromRawData(rawData, "mysql.port")
	if err != nil {
		// 如果端口配置错误，返回400错误响应。
		resp.BadRequest(ctx, "端口配置错误", err.Error())
		return
	}
	mysqlConfig.Port = port

	// 提取MySQL数据库名称。
	mysqlConfig.DB = strings.TrimSpace(rawData["mysql.database"].(string))

	// 解析最大打开连接数。
	maxOpen, err := tools.GetUInt16FromRawData(rawData, "mysql.max_open")
	if err != nil {
		// 如果最大打开连接数配置错误，返回400错误响应。
		resp.BadRequest(ctx, "最大打开连接数配置错误", err.Error())
		return
	}
	mysqlConfig.MaxOpen = maxOpen

	// 解析最大空闲连接数。
	maxIdle, err := tools.GetUInt16FromRawData(rawData, "mysql.max_idle")
	if err != nil {
		// 如果最大空闲连接数配置错误，返回400错误响应。
		resp.BadRequest(ctx, "最大空闲连接数配置错误", err.Error())
		return
	}
	mysqlConfig.MaxIdle = maxIdle

	// 检查最大空闲连接数是否大于最大打开连接数。
	if maxIdle > maxOpen {
		resp.BadRequest(ctx, "最大空闲连接数不能大于最大打开连接数", nil)
		return
	}

	// 验证MySQL连接配置。
	if err = tools.AnalyzeMySqlConnect(&mysqlConfig); err != nil {
		// 如果连接配置验证失败，返回400错误响应。
		resp.BadRequest(ctx, "数据库连接配置错误", err.Error())
		return
	}

	// 完成配置，将配置添加到全局。
	config.MySQL = mysqlConfig

	// 返回成功响应，通知客户端配置完成。
	resp.Ok(ctx, "配置完成", config.MySQL)
}

// configOss 从请求中解析并配置 OSS（对象存储服务）相关参数。
// 参数:
//
//	ctx *gin.Context - HTTP 请求上下文，包含请求数据和响应方法。
//
// 返回值:
//
//	无直接返回值，但通过 ctx 返回 JSON 响应，指示配置成功或失败的原因。
func configOss(ctx *gin.Context) {
	ossConfig := config.OssConfig{}

	// 从请求中提取原始数据并解析为 map
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "请求数据解析错误", err.Error())
		return
	}

	// 配置 OSS 基础信息，包括 endpoint、region、access key 等
	ossConfig.Endpoint = strings.TrimSpace(rawData["oss.endpoint"].(string))
	ossConfig.Region = strings.TrimSpace(rawData["oss.region"].(string))
	ossConfig.AccessKeyId = strings.TrimSpace(rawData["oss.access_key_id"].(string))
	ossConfig.AccessKeySecret = strings.TrimSpace(rawData["oss.access_key_secret"].(string))
	ossConfig.Bucket = strings.TrimSpace(rawData["oss.bucket"].(string))
	if err = tools.AnalyzeOssConfig(&ossConfig); err != nil {
		resp.BadRequest(ctx, "OSS 配置错误", err.Error())
		return
	}

	// 配置 OSS 图片路径，并验证路径的合法性
	imageOssPath := strings.TrimSpace(rawData["oss.image_oss_path"].(string))
	if err := tools.AnalyzeOssPath(imageOssPath); err != nil {
		resp.BadRequest(ctx, "图片 OSS 路径配置错误", err.Error())
		return
	}
	ossConfig.ImageOssPath = imageOssPath

	// 配置 OSS 博客路径，并验证路径的合法性
	blogOssPath := strings.TrimSpace(rawData["oss.blog_oss_path"].(string))
	if err := tools.AnalyzeOssPath(blogOssPath); err != nil {
		resp.BadRequest(ctx, "博客 OSS 路径配置错误", err.Error())
		return
	}
	ossConfig.BlogOssPath = blogOssPath

	// 配置 WebP 相关参数，包括启用状态、压缩质量和压缩后大小
	webpEnable, err := tools.GetUInt16FromRawData(rawData, "oss.webp.enable")
	if err != nil {
		resp.BadRequest(ctx, "WebP 启用配置错误", err.Error())
		return
	}
	ossConfig.WebP.Enable = webpEnable == 1

	webpQuality, err := tools.GetFloatFromRawData(rawData, "oss.webp.quality")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩质量配置错误", err.Error())
		return
	}
	ossConfig.WebP.Quality = webpQuality

	webpSize, err := tools.GetFloatFromRawData(rawData, "oss.webp.size")
	if err != nil {
		resp.BadRequest(ctx, "WebP 压缩后大小配置错误", err.Error())
		return
	}
	ossConfig.WebP.Size = webpSize

	// 将配置完成的 OSS 配置对象保存到全局变量中
	config.Oss = ossConfig

	// 返回成功响应，包含配置完成的 OSS 配置信息
	resp.Ok(ctx, "配置完成", config.Oss)
}

// configCache 从请求上下文中解析缓存配置，并将其存储到全局配置中。
// 参数:
//
//	ctx - *gin.Context: HTTP 请求上下文，用于获取请求数据和返回响应。
//
// 返回值:
//
//	无直接返回值，但通过 ctx 返回 HTTP 响应。
func configCache(ctx *gin.Context) {
	cacheConfig := config.CacheConfig{}

	// 从请求中提取原始数据并解析为 map
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "请求数据解析错误", err.Error())
		return
	}

	// 解析 AOF 启用配置
	aofEnable, err := tools.GetUInt16FromRawData(rawData, "cache.aof.enable")
	if err != nil {
		resp.BadRequest(ctx, "AOF 启用配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Enable = aofEnable == 1

	// 解析 AOF 文件路径，并生成绝对路径
	projPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(rawData["cache.aof.path"].(string)))
	if err != nil {
		resp.BadRequest(ctx, "AOF 路径配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Path = filepath.Join(projPath, "aof", "h2blog.aof")

	// 解析 AOF 文件的最大大小配置
	maxSize, err := tools.GetUInt16FromRawData(rawData, "cache.aof.max_size")
	if err != nil {
		resp.BadRequest(ctx, "AOF 最大文件大小配置错误", err.Error())
		return
	}
	cacheConfig.Aof.MaxSize = maxSize

	// 解析 AOF 文件压缩配置
	aofCompress, err := tools.GetUInt16FromRawData(rawData, "cache.aof.compress")
	if err != nil {
		resp.BadRequest(ctx, "AOF 文件压缩配置错误", err.Error())
		return
	}
	cacheConfig.Aof.Compress = aofCompress == 1

	// 将解析完成的缓存配置存储到全局配置中
	config.Cache = cacheConfig

	// 返回成功响应，包含配置信息
	resp.Ok(ctx, "配置完成", config.Cache)
}

// configLogger 从请求上下文中解析日志配置参数，并将其设置为全局日志配置。
// 参数:
//
//	ctx - gin.Context，包含请求的上下文信息，用于解析请求数据和返回响应。
//
// 返回值:
//
//	无直接返回值，但通过 ctx 返回 HTTP 响应，指示配置成功或失败的具体原因。
func configLogger(ctx *gin.Context) {
	loggerConfig := config.LoggerConfigData{}

	// 从请求中提取原始数据并解析为 map 结构
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "配置解析错误", err.Error())
		return
	}

	// 解析日志级别配置，并验证其合法性
	level := strings.TrimSpace(rawData["logger.level"].(string))
	if err := tools.AnalyzeLoggerLevel(level); err != nil {
		resp.BadRequest(ctx, "日志级别配置错误", err.Error())
		return
	}
	loggerConfig.Level = level

	// 解析日志路径配置，并生成绝对路径
	projPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(rawData["logger.path"].(string)))
	if err != nil {
		resp.BadRequest(ctx, "日志路径配置错误", err.Error())
		return
	}
	loggerConfig.Path = filepath.Join(projPath, "log", "h2blog.log")

	// 解析日志文件最大大小配置
	maxSize, err := tools.GetUInt16FromRawData(rawData, "logger.max_size")
	if err != nil {
		resp.BadRequest(ctx, "日志最大文件大小配置错误", err.Error())
		return
	}
	loggerConfig.MaxSize = maxSize

	// 解析日志文件最大备份数量配置
	maxBackups, err := tools.GetUInt16FromRawData(rawData, "logger.max_backups")
	if err != nil {
		resp.BadRequest(ctx, "日志最大备份数量配置错误", err.Error())
		return
	}
	loggerConfig.MaxBackups = maxBackups

	// 解析日志文件最大保存天数配置
	maxAge, err := tools.GetUInt16FromRawData(rawData, "logger.max_age")
	if err != nil {
		resp.BadRequest(ctx, "日志最大保存天数配置错误", err.Error())
		return
	}
	loggerConfig.MaxAge = maxAge

	// 解析日志文件是否启用压缩配置
	compress, err := tools.GetUInt16FromRawData(rawData, "logger.compress")
	if err != nil {
		resp.BadRequest(ctx, "日志文件压缩配置错误", err.Error())
		return
	}
	loggerConfig.Compress = compress == 1

	// 将解析完成的日志配置设置为全局配置
	config.Logger = loggerConfig

	// 返回成功响应，包含配置完成的信息
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

	resp.Ok(ctx, "完成并保存配置", nil)

	// 若当前为初始化环境，则发送信号通知关闭配置服务
	if env.CurrentEnv == env.InitializedEnv {
		env.CompletedConfigSign <- true
	}
}
