package config

import (
	"fmt"
	"h2blog_server/pkg/filetool"
	"os"
	"os/user"
	"path/filepath"
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
	file, err := filetool.CreateFile(filepath.Join(h2BlogHomePath, "config", "h2blog_config.yaml"))
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
	Username        string `yaml:"user_name"`        // 用户名
	UserEmail       string `yaml:"user_email"`       // 用户邮箱
	SmtpAccount     string `yaml:"smtp_account"`     // 邮箱 SMTP 账号
	SmtpAddress     string `yaml:"smtp_address"`     // 邮箱 SMTP 服务器地址
	SmtpPort        uint16 `yaml:"smtp_port"`        // 邮箱 SMTP 端口
	SmtpAuthCode    string `yaml:"smtp_auth_code"`   // 邮箱 SMTP 密码
	BackgroundImage string `yaml:"background_image"` // 背景图
	AvatarImage     string `yaml:"avatar_image"`     // 头像
	WebLogo         string `yaml:"web_logo"`         // 网站 logo
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
	Endpoint        string `yaml:"endpoint"`          // OSS 服务的访问域名
	Region          string `yaml:"region"`            // OSS 服务的地域
	AccessKeyId     string `yaml:"access_key_id"`     // OSS 访问密钥ID
	AccessKeySecret string `yaml:"access_key_secret"` // OSS 访问密钥密文
	Bucket          string `yaml:"bucket"`            // OSS 存储空间名称
	ImageOssPath    string `yaml:"image_oss_path"`    // 图片存储路径
	BlogOssPath     string `yaml:"blog_oss_path"`     // 博客内容存储路径
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

// LoadConfig 加载配置文件。
// 该函数的主要功能是检查并加载 H2Blog 的配置文件，确保配置文件存在并正确加载。
// 如果配置文件不存在或加载过程中发生错误，会返回相应的错误或触发 panic。
//
// 返回值:
// - error: 如果配置文件不存在，返回 NoConfigFileErr 错误；否则返回 nil。
func LoadConfig() error {
	// 获取 H2Blog 的用户目录路径。如果获取失败，直接触发 panic。
	userHomePath, err := getH2BlogDir()
	if err != nil {
		panic(err)
	}

	// 检查配置文件是否存在。如果不存在，返回 NoConfigFileErr 错误。
	if !filetool.IsExist(filepath.Join(userHomePath, "config", "h2blog_config.yaml")) {
		return NewNoConfigFileErr("配置文件不存在")
	}

	// 使用 sync.Once 确保配置文件只加载一次。
	// 如果加载过程中发生错误，直接触发 panic。
	loadConfigLock.Do(
		func() {
			err = loadConfigFromFile()
			if err != nil {
				panic(err)
			}
		},
	)

	// 如果一切正常，返回 nil 表示加载成功。
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

// loadConfigFromFile 尝试从文件中加载配置。
// 该函数首先检查用户的主目录下是否存在配置文件，如果存在则加载配置。
// 如果在预期路径中未找到配置文件，则返回错误。
//
// 返回值:
// - error: 如果加载配置失败或配置文件不存在，则返回相应的错误信息。
func loadConfigFromFile() error {
	// 尝试获取用户的主目录下的 h2blog 目录路径
	h2blogDir, err := getH2BlogDir()
	if err != nil {
		return fmt.Errorf("获取h2blog目录失败: %w", err)
	}

	// 构造配置文件的路径，并检查该路径下的配置文件是否存在
	configPath := filepath.Join(h2blogDir, "config", "h2blog_config.yaml")
	if filetool.IsExist(configPath) {
		// 如果配置文件存在，则尝试从该路径加载配置
		return loadConfigFromPath(configPath)
	}

	// 如果未找到配置文件，返回错误
	return fmt.Errorf("config file not found")
}

// loadConfigFromPath 从指定路径加载配置文件。
// 参数:
//
//	configPath - 配置文件的路径。
//
// 返回值:
//
//	如果加载或解析配置文件时发生错误，则返回错误。
func loadConfigFromPath(configPath string) error {
	// 读取配置文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 如果读取配置文件时发生错误，返回错误信息
		return fmt.Errorf("load config file error: %w", err)
	}

	// 初始化配置结构体并解析配置文件内容到该结构体
	conf := &ProjectConfig{}
	if err = yaml.Unmarshal(data, &conf); err != nil {
		// 如果解析配置到结构体时发生错误，返回错误信息
		return fmt.Errorf("reflect config to struct error: %w", err)
	}

	// 设置全局配置
	User = conf.User
	Server = conf.Server
	Logger = conf.Logger
	MySQL = conf.MySQL
	Oss = conf.Oss
	Cache = conf.Cache

	// 如果一切正常，返回nil表示成功
	return nil
}
