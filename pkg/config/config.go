package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

// projectConfig 定义了整个项目的配置数据结构
type projectConfig struct {
	User   userConfig       `yaml:"user"`   // 用户配置
	Server serverConfigData `yaml:"server"` // 服务器配置
	Logger loggerConfigData `yaml:"logger"` // 日志配置
	MySQL  mySQLConfigData  `yaml:"mysql"`  // MySQL数据库配置
	OSS    ossConfig        `yaml:"oss"`    // OSS配置
}

// User 定义了用户的配置信息
type userConfig struct {
	Username      string     `yaml:"username"`        // Username 是用户的用户名
	Email         string     `yaml:"email"`           // Email 是用户的电子邮件地址
	ImageOssPath  string     `yaml:"image_oss_path"`  // ImagePath 是用户的图片路径
	AvatarOssPath string     `yaml:"avatar_oss_path"` // AvatarPath 是用户的头像路径
	BlogOssPath   string     `yaml:"blog_oss_path"`   // BlogPath 是用户的博客路径
	WebP          webPConfig `yaml:"webp"`            // WebP 是用户的 WebP 配置
}

// WebPConfig 定义了 WebP 图片的配置信息
type webPConfig struct {
	Enable  bool    `yaml:"enable"`  // Enable 表示是否启用 WebP 格式
	Quality float32 `yaml:"quality"` // Quality 表示 WebP 图片的质量，取值范围为 1-100
	Size    float32 `yaml:"size"`    // Size 表示 WebP 图片的最大大小，单位为 MB
}

// serverConfigData 定义了服务器相关的配置数据结构
type serverConfigData struct {
	Port                uint16         `yaml:"port"`                  // 服务器端口号
	TokenKey            string         `yaml:"token_key"`             // 用于签发和验证JWT的密钥
	TokenExpireDuration int            `yaml:"token_expire_duration"` // token 过期时间，单位：天
	Cors                corsConfigData `yaml:"cors"`                  // CORS配置
	ResourcesPath       string         `yaml:"resources_path"`        // 资源文件路径
}

// corsConfigData 定义了CORS（跨域资源共享）的配置数据结构
type corsConfigData struct {
	Origins []string `yaml:"origins"` // 允许的源列表
	Headers []string `yaml:"headers"` // 允许的头部列表
	Methods []string `yaml:"methods"` // 允许的方法列表
}

// loggerConfigData 定义了日志相关的配置数据结构
type loggerConfigData struct {
	Level      string `yaml:"level"`       // 日志级别
	Path       string `yaml:"path"`        // 日志文件路径
	MaxAge     int    `yaml:"max_age"`     // 日志文件最大保存天数
	MaxSize    int    `yaml:"max_size"`    // 日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups"` // 日志文件最大备份个数
	Compress   bool   `yaml:"compress"`    // 是否压缩日志文件
}

// mySQLConfigData 定义了MySQL数据库相关的配置数据结构
type mySQLConfigData struct {
	User     string `yaml:"user"`     // 数据库用户名
	Password string `yaml:"password"` // 数据库密码
	Host     string `yaml:"host"`     // 数据库主机地址
	Port     uint16 `yaml:"port"`     // 数据库端口号
	DB       string `yaml:"database"` // 数据库名称
	MaxOpen  int    `yaml:"max_open"` // 最大打开的数据库连接数
	MaxIdle  int    `yaml:"max_idle"` // 最大空闲的数据库连接数
}

type ossConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
	AccessKeyId     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Bucket          string `yaml:"bucket"`
}

// loadConfigLock 用于确保配置文件只被加载一次
var loadConfigLock sync.Once

// UserConfig 全局用户配置变量
var UserConfig = &userConfig{}

// ServerConfig 全局服务器配置变量
var ServerConfig = new(serverConfigData)

// LoggerConfig 全局日志配置变量
var LoggerConfig = new(loggerConfigData)

// MySQLConfig 全局MySQL数据库配置变量
var MySQLConfig = new(mySQLConfigData)

// OSSConfig 全局OSS配置变量
var OSSConfig = new(ossConfig)

// LoadConfig 加载配置文件
func LoadConfig(configFilePath string) {
	loadConfigLock.Do(
		func() {
			// 读取配置文件内容
			data, err := os.ReadFile(configFilePath)
			if err != nil {
				msg := fmt.Sprintf("Load config file error: %v", err)
				panic(msg) // 如果读取配置文件失败，则抛出异常
			}

			// 解析配置文件内容
			conf := &projectConfig{}
			err = yaml.Unmarshal(data, &conf)
			if err != nil {
				panic("Reflect config to Struct error") // 如果解析配置文件失败，则抛出异常
			}

			// 将解析后的配置赋值给全局变量
			UserConfig = &conf.User
			ServerConfig = &conf.Server
			LoggerConfig = &conf.Logger
			MySQLConfig = &conf.MySQL
			OSSConfig = &conf.OSS
		},
	)
}
