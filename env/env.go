package env

import (
	"os"
	"os/user"
	"path/filepath"
)

// Env 运行时的环境名称
const (
	DebugEnv = "debug"
	ProdEnv  = "prod"
)

// CurrentEnv 全局环境变量
var CurrentEnv = DebugEnv

// SparrowBlogHome 项目数据目录路径
var SparrowBlogHome string

// InitSparrowBlogHome 初始化项目数据目录路径
// 检查环境变量 SPARROW_BLOG_HOME，如果不存在则使用默认路径 ~/.sparrow_blog
//
// 返回值:
//   - string: 最终使用的项目数据目录路径
//   - error: 初始化过程中遇到的任何错误
func InitSparrowBlogHome() (string, error) {
	// 先检查环境变量
	if home := os.Getenv("SPARROW_BLOG_HOME"); home != "" {
		SparrowBlogHome = home
		return SparrowBlogHome, nil
	}

	// 获取当前用户信息
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	// 使用默认路径 ~/.sparrow_blog
	defaultPath := filepath.Join(usr.HomeDir, ".sparrow_blog")
	SparrowBlogHome = defaultPath

	// 设置环境变量，确保其他组件能够使用
	os.Setenv("SPARROW_BLOG_HOME", SparrowBlogHome)

	return SparrowBlogHome, nil
}

// GetSparrowBlogHome 获取项目数据目录路径
//
// 返回值:
//   - string: 项目数据目录路径
func GetSparrowBlogHome() string {
	return SparrowBlogHome
}
