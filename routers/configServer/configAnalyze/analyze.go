package configAnalyze

import (
	"fmt"
	"math"
	"net"
	"net/url"
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
		// 若为 https://xxx，则解析
		address, err := url.Parse(ori)
		if err != nil {
			return err
		}
		// 若协议不是 https，则报错
		if address.Scheme != "https" {
			return fmt.Errorf("跨域请求来源地址 %v 的协议不是 https", origin)
		}
		// 若域名解析失败，则报错
		if _, err = net.LookupHost(ori); err != nil {
			return fmt.Errorf("跨域请求来源地址 %v 解析失败", origin)
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
