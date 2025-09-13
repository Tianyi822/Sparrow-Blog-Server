package config

// ProjectConfig 定义了所有配置数据的结构
type ProjectConfig struct {
	User         UserConfigData   `yaml:"user"`          // 用户配置
	Server       ServerConfigData `yaml:"server"`        // 服务器配置
	Logger       LoggerConfigData `yaml:"logger"`        // 日志配置
	SearchEngine SearchEngineData `yaml:"search_engine"` // 搜索引擎配置
	MySQL        MySQLConfigData  `yaml:"mysql"`         // MySQL数据库配置
	Oss          OssConfig        `yaml:"oss"`           // OSS对象存储配置
	Cache        CacheConfig      `yaml:"cache"`         // 缓存配置
}

// UserConfigData 用户配置
type UserConfigData struct {
	Username          string   `yaml:"user_name"`           // 用户名
	UserEmail         string   `yaml:"user_email"`          // 用户邮箱
	UserGithubAddress string   `yaml:"user_github_address"` // Github 地址
	UserHobbies       []string `yaml:"user_hobbies"`        // 用户爱好
	TypeWriterContent []string `yaml:"type_writer_content"` // 打字机内容
	BackgroundImage   string   `yaml:"background_image"`    // 背景图
	AvatarImage       string   `yaml:"avatar_image"`        // 头像
	WebLogo           string   `yaml:"web_logo"`            // 网站 logo
	ICPFilingNumber   string   `yaml:"icp_filing_number"`   // 网站备案号
}

// ServerConfigData 定义了服务器相关配置
type ServerConfigData struct {
	Port                uint16         `yaml:"port"`                  // 服务器端口号
	TokenKey            string         `yaml:"token_key"`             // JWT签名和验证密钥
	TokenExpireDuration uint8          `yaml:"token_expire_duration"` // Token过期时间(天)
	Cors                CorsConfigData `yaml:"cors"`                  // CORS跨域配置
	SmtpAccount         string         `yaml:"smtp_account"`          // 邮箱 SMTP 账号
	SmtpAddress         string         `yaml:"smtp_address"`          // 邮箱 SMTP 服务器地址
	SmtpPort            uint16         `yaml:"smtp_port"`             // 邮箱 SMTP 端口
	SmtpAuthCode        string         `yaml:"smtp_auth_code"`        // 邮箱 SMTP 密码
	SSL                 SSLConfigData  `yaml:"ssl"`                   // SSL/TLS配置
}

// SSLConfigData 定义了SSL/TLS相关配置
type SSLConfigData struct {
	CertFile string `yaml:"cert_file"` // SSL证书文件路径
	KeyFile  string `yaml:"key_file"`  // SSL私钥文件路径
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

// SearchEngineData 搜索引擎配置
type SearchEngineData struct {
	IndexPath string `yaml:"index_path"` // 搜索索引文件路径
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
