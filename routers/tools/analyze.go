package tools

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"h2blog_server/pkg/config"
	"math"
	"net"
	"net/http"
	"net/url"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// AnalyzePort 分析端口配置
// 分析该端口是否被占用
func AnalyzePort(port string) (uint16, error) {
	// 尝试监听端口
	addr := fmt.Sprintf(":%v", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("port %v 不可用", port)
	}

	// 关闭监听，并释放端口
	err = listener.Close()
	if err != nil {
		return 0, err
	}

	val, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	} else if val < uint64(0) && val > uint64(65535) {
		return 0, fmt.Errorf("port %v 超出范围", port)
	}

	return uint16(val), nil
}

// AnalyzeTokenKey 分析 tokenKey 配置
// 若 tokenKey 的熵小于3，则认为不安全
func AnalyzeTokenKey(tokenKey string) error {

	charCount := make(map[rune]int)
	for _, char := range tokenKey {
		charCount[char]++
	}

	var entropy float64
	for _, count := range charCount {
		p := float64(count) / float64(len(tokenKey))
		entropy -= p * math.Log2(p)
	}

	if entropy < 3 {
		return fmt.Errorf("tokenKey %v 不安全", tokenKey)
	}

	return nil
}

// AnalyzeTokenExpireDuration 分析 token 过期时间
func AnalyzeTokenExpireDuration(tokenExpireDuration string) (uint8, error) {
	val, err := strconv.ParseUint(tokenExpireDuration, 10, 8)
	if err != nil {
		return 0, err
	} else if val < 0 || val > 90 {
		return 0, fmt.Errorf("token 过期时间 %v 超出范围 (0~90)", tokenExpireDuration)
	} else {
		return uint8(val), nil
	}
}

// AnalyzeCorsOrigins 分析跨域请求来源地址
func AnalyzeCorsOrigins(corsOrigins []string) error {
	for index, origin := range corsOrigins {
		ori := strings.TrimSpace(origin)
		// 若为空，则跳过
		if ori == "" {
			continue
		}
		// 若为 *，则报错
		if ori == "*" {
			return fmt.Errorf("跨域请求来源地址 %v 不允许全匹配", origin)
		}
		// 若为 https://xxx 或 http://xxx，则解析
		_, err := url.Parse(ori)
		if err != nil {
			return err
		}
		corsOrigins[index] = ori
	}
	return nil
}

// AnalyzeEmail 分析邮箱格式是否正确
func AnalyzeEmail(email string) error {
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	// 使用正则表达式判断
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	if !matched {
		return fmt.Errorf("邮箱格式不正确: %v", email)
	}
	return nil
}

// AnalyzeOssPath 分析 OSS 路径
func AnalyzeOssPath(ossPath string) error {
	if ossPath == "" {
		return fmt.Errorf("ossPath 不能为空")
	}

	if !strings.HasSuffix(ossPath, "/") {
		return fmt.Errorf("ossPath 必须以 / 结尾")
	}

	if strings.HasPrefix(ossPath, "/") {
		return fmt.Errorf("ossPath 不能以 / 开头")
	}

	return nil
}

// AnalyzeOssConfig 解析 OSS 配置
func AnalyzeOssConfig(ossConfig *config.OssConfig) error {
	// 创建一个静态凭证提供者，使用配置文件中的 AccessKeyId 和 AccessKeySecret
	provider := credentials.NewStaticCredentialsProvider(ossConfig.AccessKeyId, ossConfig.AccessKeySecret)
	// 加载默认配置，并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(ossConfig.Region)

	// 创建 Oss Client
	// 使用配置信息创建一个新的 Oss 客户端
	ossClient := oss.NewClient(cfg)

	// 获取 Bucket 信息
	// 创建一个获取 Bucket 信息的请求，指定要获取的 Bucket 名称
	request := &oss.GetBucketInfoRequest{
		Bucket: oss.Ptr(ossConfig.Bucket),
	}

	// 这里传入一个空的 context，只是用于检查是否连接成功，后续操作还是要传入项目的 context
	// 使用 Oss 客户端发送请求获取 Bucket 信息
	result, err := ossClient.GetBucketInfo(context.Background(), request)
	if err != nil {
		// 如果获取 Bucket 信息失败，记录错误日志
		return fmt.Errorf("获取 bucket 信息失败: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		return fmt.Errorf("获取 bucket 信息失败: %v", result.StatusCode)
	}

	return nil
}

func AnalyzeHostAddress(host string) error {
	// 域名验证正则表达式
	domainRegex := `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
	domainMatch, _ := regexp.MatchString(domainRegex, host)

	// IPv4 地址验证
	ip4 := net.ParseIP(host)
	isIPv4 := ip4 != nil && ip4.To4() != nil

	// IPv6 地址验证
	isIPv6 := ip4 != nil && ip4.To16() != nil && ip4.To4() == nil

	if !(domainMatch || isIPv4 || isIPv6) {
		return fmt.Errorf("host 地址 %v 格式不正确", host)
	}

	return nil
}

func AnalyzeMySqlConnect(mysqlConfig *config.MySQLConfigData) error {
	// 连接 MySQL（不指定库名）
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
			mysqlConfig.User,     // MySQL 用户名
			mysqlConfig.Password, // MySQL 密码
			mysqlConfig.Host,     // MySQL 服务器地址
			mysqlConfig.Port,     // MySQL 服务器端口
		),
	)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			panic(err)
		}
	}(db)

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`, mysqlConfig.DB))
	if err != nil {
		return err
	}

	return nil
}

// AnalyzeLoggerLevel 分析日志级别
// 只允许一下的日志级别
// - INFO
// - DEBUG
// - ERROR
// - WARN
func AnalyzeLoggerLevel(level string) error {
	level = strings.ToUpper(level)
	if level != "INFO" && level != "DEBUG" && level != "ERROR" && level != "WARN" {
		return fmt.Errorf("日志级别 %v 不正确", level)
	}
	return nil
}

// AnalyzeAbsolutePath 分析并验证给定的绝对路径。
// 如果路径为空，则尝试使用当前用户的主目录作为默认路径。
// 参数:
//
//	path - 待分析的路径字符串。如果为空，函数会尝试使用用户的主目录。
//
// 返回值:
//
//	string - 验证通过的绝对路径。
//	error - 如果发生错误（如无法获取用户信息或路径不是绝对路径），返回相应的错误信息。
func AnalyzeAbsolutePath(path string) (string, error) {
	// 如果路径为空，尝试使用用户的主目录作为默认路径
	if path == "" {
		u, err := user.Current()
		if err != nil {
			// 明确指出获取用户信息失败的原因
			return "", fmt.Errorf("无法获取当前用户信息: %w", err)
		}
		// 使用用户的主目录拼接默认路径
		path = filepath.Join(u.HomeDir, ".h2blog")
	}

	// 检查路径是否为绝对路径
	if !filepath.IsAbs(path) {
		// 提供更详细的错误信息，帮助调用者定位问题
		return "", fmt.Errorf("提供的路径 '%s' 不是绝对路径，请提供有效的绝对路径", path)
	}

	return path, nil
}
