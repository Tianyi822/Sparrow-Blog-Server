package env

// Env 运行时的环境名称
const (
	RuntimeEnv      = "RUNTIME_ENV"
	ConfigServerEnv = "CONFIG_SERVER_ENV"
)

// CurrentEnv 全局环境变量
var CurrentEnv = ConfigServerEnv

// CompletedConfigSign 全局配置加载完成信号
var CompletedConfigSign chan bool

// VerificationCode 验证码，仅在启动配置服务时使用，运行时会保存在缓存中
// 因为在启动配置服务的时候，缓存还没有初始化，不可以使用缓存，所以使用全局变量
var VerificationCode string
