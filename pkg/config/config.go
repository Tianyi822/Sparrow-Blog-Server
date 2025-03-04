package config

import (
	"bufio"
	"fmt"
	"h2blog_server/pkg/fileTool"
	"h2blog_server/pkg/utils"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ProjectConfig 定义了所有配置数据的结构
type ProjectConfig struct {
	User   UserConfigData   `yaml:"user"`   // 用户配置
	Server ServerConfigData `yaml:"server"` // 服务器配置
	Logger LoggerConfigData `yaml:"logger"` // 日志配置
	MySQL  MySQLConfigData  `yaml:"mysql"`  // MySQL数据库配置
	Oss    OssConfig        `yaml:"oss"`    // OSS对象存储配置
	Cache  CacheConfig      `yaml:"cache"`  // 缓存配置
}

func (pc *ProjectConfig) Store() error {
	h2BlogHomePath, err := getH2BlogDir()
	if err != nil {
		return fmt.Errorf("获取或创建 H2Blog 目录失败: %w", err)
	}

	// 将 ProjectConfig 结构体转换为 YAML 格式的字节数组
	yamlData, err := yaml.Marshal(pc)
	if err != nil {
		return fmt.Errorf("将配置数据转换为YAML失败: %w", err)
	}

	// 将 YAML 数据写入到文件中
	file, err := fileTool.CreateFile(filepath.Join(h2BlogHomePath, "config", "h2blog_config.yaml"))
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = file.Write(yamlData)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// UserConfigData 用户配置
type UserConfigData struct {
	Username        string `yaml:"username"`         // 用户名
	UserEmail       string `yaml:"user_email"`       // 用户邮箱
	SmtpAccount     string `yaml:"smtp_account"`     // 邮箱 SMTP 账号
	SmtpAddress     string `yaml:"smtp_address"`     // 邮箱 SMTP 服务器地址
	SmtpPort        int    `yaml:"smtp_port"`        // 邮箱 SMTP 端口
	SmtpAuthCode    string `yaml:"smtp_auth_code"`   // 邮箱 SMTP 密码
	BackgroundImage string `yaml:"background_image"` // Background image name
}

// WebPConfigData 定义了WebP图片转换设置
type WebPConfigData struct {
	Enable  bool    `yaml:"enable"`  // 是否启用WebP转换
	Quality float32 `yaml:"quality"` // WebP图片质量(1-100)
	Size    float32 `yaml:"size"`    // WebP图片最大大小(MB)
}

// ServerConfigData 定义了服务器相关配置
type ServerConfigData struct {
	Port                uint16         `yaml:"port"`                  // 服务器端口号
	TokenKey            string         `yaml:"token_key"`             // JWT签名和验证密钥
	TokenExpireDuration uint8          `yaml:"token_expire_duration"` // Token过期时间(天)
	Cors                CorsConfigData `yaml:"cors"`                  // CORS跨域配置
}

// CorsConfigData 定义了跨域资源共享配置
type CorsConfigData struct {
	Origins []string `yaml:"origins"` // 允许的源
	Headers []string `yaml:"headers"` // 允许的请求头
	Methods []string `yaml:"methods"` // 允许的请求方法
}

// LoggerConfigData 定义了日志配置
type LoggerConfigData struct {
	Level      string `yaml:"level"`       // 日志级别
	Path       string `yaml:"path"`        // 日志文件路径
	MaxAge     uint16 `yaml:"max_age"`     // 日志文件保留最大天数
	MaxSize    uint16 `yaml:"max_size"`    // 日志文件最大大小(MB)
	MaxBackups uint16 `yaml:"max_backups"` // 日志备份文件最大数量
	Compress   bool   `yaml:"compress"`    // 是否压缩日志文件
}

// MySQLConfigData 定义了MySQL数据库配置
type MySQLConfigData struct {
	User     string `yaml:"user"`     // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
	Host     string `yaml:"host"`     // 数据库主机地址
	Port     uint16 `yaml:"port"`     // 数据库端口号
	DB       string `yaml:"database"` // 数据库名称
	MaxOpen  uint16 `yaml:"max_open"` // 最大打开连接数
	MaxIdle  uint16 `yaml:"max_idle"` // 最大空闲连接数
}

// OssConfig 定义了对象存储服务配置
type OssConfig struct {
	Endpoint        string         `yaml:"endpoint"`          // OSS 服务的访问域名
	Region          string         `yaml:"region"`            // OSS 服务的地域
	AccessKeyId     string         `yaml:"access_key_id"`     // OSS 访问密钥ID
	AccessKeySecret string         `yaml:"access_key_secret"` // OSS 访问密钥密文
	Bucket          string         `yaml:"bucket"`            // OSS 存储空间名称
	ImageOssPath    string         `yaml:"image_oss_path"`    // 图片存储路径
	BlogOssPath     string         `yaml:"blog_oss_path"`     // 博客内容存储路径
	WebP            WebPConfigData `yaml:"webp"`              // WebP图片配置
}

// CacheConfig 定义了缓存系统配置
type CacheConfig struct {
	Aof AofConfig `yaml:"aof"` // AOF持久化配置
}

// AofConfig 定义了追加文件持久化配置
type AofConfig struct {
	Enable   bool   `yaml:"enable"`   // 是否启用AOF持久化
	Path     string `yaml:"path"`     // AOF文件路径
	MaxSize  uint16 `yaml:"max_size"` // AOF文件最大大小(MB)
	Compress bool   `yaml:"compress"` // 是否压缩AOF文件
}

// 全局配置变量
var (
	// loadConfigLock 确保配置只被加载一次
	loadConfigLock sync.Once

	// User 保存全局用户配置
	User UserConfigData

	// Server 保存全局服务器配置
	Server ServerConfigData

	// Logger 保存全局日志配置
	Logger LoggerConfigData

	// MySQL 保存全局MySQL数据库配置
	MySQL MySQLConfigData

	// Oss 保存全局OSS配置
	Oss OssConfig

	// Cache 保存全局缓存配置
	Cache CacheConfig
)

// LoadConfig 加载配置文件
func LoadConfig() error {
	userHomePath, err := getH2BlogDir()
	if err != nil {
		panic(err)
	}

	if !fileTool.IsExist(filepath.Join(userHomePath, "config", "h2blog_config.yaml")) {
		return NewNoConfigFileErr("配置文件不存在")
	}

	loadConfigLock.Do(
		func() {
			err = loadConfigFromFile()
			if err != nil {
				err = loadConfigFromTerminal()
			}

			// If both loading from file and terminal failed, then panic
			if err != nil {
				panic(err)
			}
		},
	)

	return nil
}

// getH2BlogDir 返回h2blog配置和数据的基础目录
// 它使用用户的主目录作为基础路径
//
// 返回:
//   - string: h2blog目录的完整路径
//   - error: 获取用户主目录时遇到的任何错误
func getH2BlogDir() (string, error) {
	// 从环境变量中获取 H2BLOG_HOME
	if h2blogHome := os.Getenv("H2BLOG_HOME"); h2blogHome != "" {
		return h2blogHome, nil
	}

	// 获取当前用户信息
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("获取当前用户失败: %w", err)
	}

	// 将主目录与.h2blog连接
	return filepath.Join(usr.HomeDir, ".h2blog"), nil
}

// loadConfigFromFile 从文件加载配置数据
func loadConfigFromFile() error {
	// 首先尝试用户的主目录
	h2blogDir, err := getH2BlogDir()
	if err != nil {
		return fmt.Errorf("获取h2blog目录失败: %w", err)
	}

	// 首先尝试主目录
	configPath := filepath.Join(h2blogDir, "config", "h2blog_config.yaml")
	if fileTool.IsExist(configPath) {
		return loadConfigFromPath(configPath)
	}

	// 如果在主目录中未找到，则尝试当前目录
	currentDirPath := filepath.Join(".h2blog", "config", "h2blog_config.yaml")
	if fileTool.IsExist(currentDirPath) {
		return loadConfigFromPath(currentDirPath)
	}

	return fmt.Errorf("config file not found")
}

// loadConfigFromPath 从指定路径加载配置
func loadConfigFromPath(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("load config file error: %w", err)
	}

	conf := &ProjectConfig{}
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return fmt.Errorf("reflect config to struct error: %w", err)
	}

	// 设置全局配置
	User = conf.User
	Server = conf.Server
	Logger = conf.Logger
	MySQL = conf.MySQL
	Oss = conf.Oss
	Cache = conf.Cache

	return nil
}

// checkPortAvailable 尝试绑定端口以检查其可用性
// 如果端口已被占用则返回错误
// 参数:
//   - port: 要检查的端口号
//
// 返回:
//   - error: 如果端口可用则返回nil，如果端口被占用则返回错误信息
func checkPortAvailable(port uint16) error {
	// Try to listen on the port
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("port %d is not available", port)
	}
	// Close the listener immediately to free the port
	err = listener.Close()
	if err != nil {
		return err
	}
	return nil
}

// loadConfigFromTerminal 从终端输入加载配置
// 提示用户输入必要的配置值并将其保存到配置文件中
func loadConfigFromTerminal() error {
	fmt.Println("未找到配置文件。请输入配置值:")

	// 获取 h2blog 目录
	h2blogDir, err := getH2BlogDir()
	if err != nil {
		return fmt.Errorf("获取 h2blog 目录失败: %w", err)
	}

	// 更新日志和 AOF 的默认路径
	defaultLogPath := filepath.Join(h2blogDir, "logs", "h2blog.log")
	defaultAofPath := filepath.Join(h2blogDir, "aof", "h2blog.aof")

	conf := &ProjectConfig{
		User: UserConfigData{
			Username:  getInput("Username: "),
			UserEmail: getInput("Email: "),
		},

		Server: ServerConfigData{
			Port: func() uint16 {
				for {
					port := getUint16Input("Server port (press Enter to use default '2233'): ", 0, 65535, "2233")

					// Check if the specified port is available
					if err := checkPortAvailable(port); err != nil {
						fmt.Printf("Port %d is not available. Please choose a different port.\n", port)
						continue
					}

					return port
				}
			}(),
			TokenKey:            getInput("JWT token key (press Enter for generating random string as token key): ", utils.GenRandomString(32)),
			TokenExpireDuration: getUint8Input("Token expiration in days (press Enter for default 1 day): ", 1, 365, "1"),
			Cors: CorsConfigData{
				Origins: []string{"*"}, // Default values
				Headers: []string{"Content-Type", "Authorization", "Token"},
				Methods: []string{"GET", "POST", "PUT", "DELETE"},
			},
		},

		Logger: LoggerConfigData{
			Level:      "info",
			Path:       getInput(fmt.Sprintf("Log file path (press Enter to use default '%s'): ", defaultLogPath), defaultLogPath),
			MaxAge:     getUint16Input("Log max age in days (press Enter for default 1 day): ", 1, 365, "1"),
			MaxSize:    getUint16Input("Log max size in MB (press Enter for default 1): ", 1, 100, "1"),
			MaxBackups: getUint16Input("Max log backups (press Enter for default 10): ", 1, 100, "10"),
			Compress:   getBoolInput("Compress logs? (y/n) (press Enter for default yes): ", "y"),
		},

		MySQL: MySQLConfigData{
			User:     getInput("MySQL username: "),
			Password: getInput("MySQL password: "),
			Host:     getInput("MySQL host: "),
			Port:     getUint16Input("MySQL port (1-65535): ", 1, 65535),
			DB:       getInput("MySQL database name: "),
			MaxOpen:  getUint16Input("Max open connections (1-1000) (press Enter for default 10): ", 1, 1000, "10"),
			MaxIdle:  getUint16Input("Max idle connections (1-100) (press Enter for default 10): ", 1, 100, "10"),
		},

		Oss: OssConfig{
			Endpoint:        getInput("OSS endpoint: "),
			Region:          getInput("OSS region: "),
			AccessKeyId:     getInput("OSS access key ID: "),
			AccessKeySecret: getInput("OSS access key secret: "),
			Bucket:          getInput("OSS bucket name: "),
			ImageOssPath:    "images/", // Default values
			BlogOssPath:     "blogs/",
			WebP: WebPConfigData{
				Enable:  getBoolInput("Enable WebP conversion? (y/n) (press Enter for default yes): ", "y"),
				Quality: getFloatInput("WebP quality (1-100) (press Enter for default 75): ", 1, 100, "75"),
				Size:    getFloatInput("Maximum WebP size in MB (press Enter for default 1): ", 0.1, 10, "1"),
			},
		},

		Cache: CacheConfig{
			Aof: AofConfig{
				Enable:   getBoolInput("Enable AOF persistence? (y/n) (press Enter for default yes): ", "y"),
				Path:     getInput(fmt.Sprintf("AOF file path (press Enter to use default '%s'): ", defaultAofPath), defaultAofPath),
				MaxSize:  getUint16Input("AOF max size in MB (press Enter for default 1): ", 1, 10, "1"),
				Compress: getBoolInput("Compress AOF files? (y/n) (press Enter for default yes): ", "y"),
			},
		},
	}

	// Set global variables
	User = conf.User
	Server = conf.Server
	Logger = conf.Logger
	MySQL = conf.MySQL
	Oss = conf.Oss
	Cache = conf.Cache

	// Create config directory in user's home
	configDir := filepath.Join(h2blogDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save configuration to file
	data, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDir, "h2blog_config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to save config file: %w", err)
	}

	fmt.Printf("Configuration saved to %s\n", configPath)
	return nil
}

// Helper functions for getting user input

// getInput 提示用户输入并确保返回非空响应
func getInput(prompt string, defaultValue ...string) string {
	for {
		fmt.Print(prompt)
		var input string

		// Use bufio.NewReader to read full line including spaces
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v. Please try again.\n", err)
			continue
		}

		// Remove trailing newline and spaces
		input = strings.TrimSpace(line)

		// Ensure input is not empty
		if input != "" {
			return input
		}

		// If no input is provided and a default value is provided, use the default value
		if input == "" && len(defaultValue) > 0 {
			return defaultValue[0]
		}

		fmt.Println("Input cannot be empty. Please try again.")
	}
}

// getBoolInput 提示用户输入是/否响应，对于是返回true
func getBoolInput(prompt string, defaultValue ...string) bool {
	for {
		input := strings.ToLower(getInput(prompt, defaultValue...))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

func getUint8Input(prompt string, min, max uint16, defaultValue ...string) uint8 {
	for {
		input := getInput(prompt, defaultValue...)
		val, err := strconv.ParseUint(input, 10, 8)
		if err == nil && val >= uint64(min) && val <= uint64(max) {
			return uint8(val)
		}
		fmt.Printf("Please enter a number between %d and %d\n", min, max)
	}
}

// getUint16Input 提示用户输入指定范围内的uint16数值
func getUint16Input(prompt string, min, max uint16, defaultValue ...string) uint16 {
	for {
		input := getInput(prompt, defaultValue...)
		val, err := strconv.ParseUint(input, 10, 16)
		if err == nil && val >= uint64(min) && val <= uint64(max) {
			return uint16(val)
		}
		fmt.Printf("Please enter a number between %d and %d\n", min, max)
	}
}

// getFloatInput 提示用户输入指定范围内的float32数值
func getFloatInput(prompt string, min, max float32, defaultValue ...string) float32 {
	for {
		input := getInput(prompt, defaultValue...)

		val, err := strconv.ParseFloat(input, 32)
		if err == nil && float32(val) >= min && float32(val) <= max {
			return float32(val)
		}
		fmt.Printf("Please enter a number between %.1f and %.1f\n", min, max)
	}
}
