package env

// Env 运行时的环境类型
type Env string

// Env 运行时的环境名称
const (
	RuntimeEnv      Env = "RUNTIME_ENV"
	ConfigServerEnv Env = "CONFIG_SERVER_ENV"
)

// CurrentEnv 全局环境变量
var CurrentEnv Env

var CompletedConfigSign chan bool
