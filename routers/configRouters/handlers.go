package configRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"
	"path/filepath"
	"strings"
)

func configBase(ctx *gin.Context) {
	serverConfig := &config.ServerConfigData{}

	// 以下两项不需要前端传入配置，直接写死
	serverConfig.Cors.Headers = []string{"Content-Type", "Authorization", "X-CSRF-Token"}
	serverConfig.Cors.Methods = []string{"POST", "PUT", "DELETE", "GET"}

	// 解析端口
	portStr := strings.TrimSpace(ctx.PostForm("server.port"))
	port, err := tools.AnalyzePort(portStr)
	if err != nil {
		resp.BadRequest(ctx, "端口配置错误", err.Error())
		return
	}
	serverConfig.Port = port

	// 解析 Token 密钥
	tokenKey := strings.TrimSpace(ctx.PostForm("server.token_key"))
	if err = tools.AnalyzeTokenKey(tokenKey); err != nil {
		resp.BadRequest(ctx, "Token 密钥配置错误", err.Error())
		return
	}
	serverConfig.TokenKey = tokenKey

	tokenExpireDuration := strings.TrimSpace(ctx.PostForm("server.token_expire_duration"))
	dur, err := tools.AnalyzeTokenExpireDuration(tokenExpireDuration)
	if err != nil {
		resp.BadRequest(ctx, "Token 过期时间配置错误", err.Error())
		return
	}
	serverConfig.TokenExpireDuration = dur

	corsOrigins := ctx.PostFormArray("server.cors.origins")
	if err = tools.AnalyzeCorsOrigins(corsOrigins); err != nil {
		resp.BadRequest(ctx, "跨域源配置错误", err.Error())
		return
	}
	serverConfig.Cors.Origins = corsOrigins

	// 完成配置，将配置添加到全局
	config.Server = serverConfig

	resp.Ok(ctx, "配置完成", config.Server)
}

func configUser(ctx *gin.Context) {
	userConfig := &config.UserConfigData{}

	userConfig.Username = strings.TrimSpace(ctx.PostForm("user.username"))

	userEmail := strings.TrimSpace(ctx.PostForm("user.user_email"))
	if err := tools.AnalyzeEmail(userEmail); err != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", err.Error())
		return
	}
	userConfig.UserEmail = userEmail

	smtpAccount := strings.TrimSpace(ctx.PostForm("user.smtp_account"))
	if err := tools.AnalyzeEmail(smtpAccount); err != nil {
		resp.BadRequest(ctx, "系统邮箱配置错误", err.Error())
		return
	}
	userConfig.SmtpAccount = smtpAccount

	userConfig.SmtpAddress = strings.TrimSpace(ctx.PostForm("user.smtp_address"))

	smtpPort, err := tools.GetIntFromPostForm(ctx, "user.smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "系统邮箱端口配置错误", err.Error())
		return
	}
	userConfig.SmtpPort = smtpPort

	userConfig.SmtpAuthCode = strings.TrimSpace(ctx.PostForm("user.smtp_auth_code"))

	// 完成配置，将配置添加到全局
	config.User = userConfig

	resp.Ok(ctx, "配置完成", config.User)
}

func verifyEmail(ctx *gin.Context) {
	resp.Ok(ctx, "邮箱验证成功", nil)
}

func configMysql(ctx *gin.Context) {
	mysqlConfig := &config.MySQLConfigData{}

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

	if err = tools.AnalyzeMySqlConnect(mysqlConfig); err != nil {
		resp.BadRequest(ctx, "数据库连接配置错误", err.Error())
		return
	}

	// 完成配置，将配置添加到全局
	config.MySQL = mysqlConfig

	resp.Ok(ctx, "配置完成", config.MySQL)
}

func configOss(ctx *gin.Context) {
	ossConfig := &config.OssConfig{}

	// OSS 基础配置
	ossConfig.Endpoint = strings.TrimSpace(ctx.PostForm("oss.endpoint"))
	ossConfig.Region = strings.TrimSpace(ctx.PostForm("oss.region"))
	ossConfig.AccessKeyId = strings.TrimSpace(ctx.PostForm("oss.access_key_id"))
	ossConfig.AccessKeySecret = strings.TrimSpace(ctx.PostForm("oss.access_key_secret"))
	ossConfig.Bucket = strings.TrimSpace(ctx.PostForm("oss.bucket"))
	if err := tools.AnalyzeOssConfig(ossConfig); err != nil {
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
	cacheConfig := &config.CacheConfig{}

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
	loggerConfig := &config.LoggerConfigData{}

	level := strings.TrimSpace(ctx.PostForm("logger.level"))
	if err := tools.AnalyzeLoggerLevel(level); err != nil {
		resp.BadRequest(ctx, "日志级别配置错误", err.Error())
		return
	}
	loggerConfig.Level = level

	logPath, err := tools.AnalyzeAbsolutePath(strings.TrimSpace(ctx.PostForm("logger.path")))
	if err != nil {
		resp.BadRequest(ctx, "日志路径配置错误", err.Error())
		return
	}
	loggerConfig.Path = filepath.Join(logPath, "h2blog.log")

	maxSize, err := tools.GetUInt16FromPostForm(ctx, "logger.max_size")
	if err != nil {
		resp.BadRequest(ctx, "日志最大文件大小配置错误", err.Error())
		return
	}
	loggerConfig.MaxSize = maxSize

	maxBackups, err := tools.GetUInt16FromPostForm(ctx, "logger.max_backups")
	if err != nil {
		resp.BadRequest(ctx, "日志最大备份数量配置错误", err.Error())
		return
	}
	loggerConfig.MaxBackups = maxBackups

	maxAge, err := tools.GetUInt16FromPostForm(ctx, "logger.max_age")
	if err != nil {
		resp.BadRequest(ctx, "日志最大保存天数配置错误", err.Error())
		return
	}
	loggerConfig.MaxAge = maxAge

	compress, err := tools.GetIntFromPostForm(ctx, "logger.compress")
	if err != nil {
		resp.BadRequest(ctx, "日志文件压缩配置错误", err.Error())
		return
	}
	loggerConfig.Compress = compress == 1

	// 完成配置，将配置添加到全局
	config.Logger = loggerConfig

	resp.Ok(ctx, "配置完成", config.Logger)
}

// closeConfigServer 关闭配置服务
func closeConfigServer(ctx *gin.Context) {
	resp.Ok(ctx, "关闭配置服务", nil)
	// 配置完成，通知关闭配置服务
	env.CompletedConfigSign <- true
}
