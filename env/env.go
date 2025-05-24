package env

// Env 运行时的环境名称
const (
	DebugEnv = "debug"
	ProdEnv  = "prod"
)

// CurrentEnv 全局环境变量
var CurrentEnv = DebugEnv
