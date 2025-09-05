package main

import (
	// 标准库
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"sparrow_blog_server/env"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/routers"
	"sparrow_blog_server/routers/adminrouter"
	"sparrow_blog_server/routers/webrouter"
	"sparrow_blog_server/searchengine"
	"sparrow_blog_server/storage"
)

// Args 存储从命令行解析得到的参数配置
var Args map[string]string

// initializeApplicationComponents 初始化应用程序核心基础组件
// @param ctx 上下文对象，用于控制初始化超时和取消操作
func initializeApplicationComponents(ctx context.Context) {
	// 设置全局运行环境变量，影响后续所有组件的初始化行为
	env.CurrentEnv = Args["env"]

	// 优先初始化日志系统，确保后续初始化过程可以被记录
	err := logger.InitLogger(ctx)
	if err != nil {
		panic("日志模块初始化失败，请检查配置文件是否有误")
	}

	// 初始化数据存储层，包括数据库连接池和缓存系统
	err = storage.InitStorage(ctx)
	if err != nil {
		panic("数据层初始化失败，请检查配置文件是否有误")
	}

	// 初始化 Bleve 搜索引擎，加载中文分词索引
	err = searchengine.LoadingIndex(ctx)
	if err != nil {
		panic("搜索引擎初始化失败，请检查配置文件是否有误")
	}
}

// startWebServer 启动 Web 服务器并配置完整的路由系统
// @return *http.Server HTTP 服务器实例，用于后续的优雅关闭操作
func startWebServer() *http.Server {
	// 注册所有路由模块，包括前台和后台管理路由
	logger.Info("加载路由信息")
	routers.IncludeOpts(
		webrouter.Router,    // 前台博客路由
		adminrouter.Routers, // 后台管理路由
	)
	logger.Info("路由信息加载完成")

	// 初始化 Gin 路由引擎，应用中间件和路由配置
	logger.Info("配置路由")
	routerEngine := routers.InitRouter()
	logger.Info("路由配置完成")

	// 构建服务器监听地址
	serverPort := fmt.Sprintf(":%v", config.Server.Port)

	// 创建 HTTP 服务器实例
	logger.Info("启动服务中")
	webServer := &http.Server{
		Addr:    serverPort,
		Handler: routerEngine,
	}

	// 在独立 goroutine 中启动服务器，避免阻塞主线程
	go func() {
		var err error

		// 根据环境变量选择协议：生产环境使用 HTTPS，开发环境使用 HTTP
		if env.CurrentEnv == env.ProdEnv {
			// 生产环境：启动 HTTPS 服务
			certificateFile := config.Server.SSL.CertFile
			privateKeyFile := config.Server.SSL.KeyFile

			// 验证 SSL 证书文件配置
			if certificateFile == "" || privateKeyFile == "" {
				logger.Fatal("生产环境需要配置SSL证书文件路径")
				return
			}

			logger.Info("启动 HTTPS 服务，监听端口: %s", serverPort)
			err = webServer.ListenAndServeTLS(certificateFile, privateKeyFile)
		} else {
			// 开发环境：启动 HTTP 服务
			logger.Info("启动 HTTP 服务，监听端口: %s", serverPort)
			err = webServer.ListenAndServe()
		}

		// 处理服务器启动错误，忽略正常关闭信号
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("监听端口失败: %s\n", err)
		}
	}()

	// 记录服务启动完成状态
	if env.CurrentEnv == env.ProdEnv {
		logger.Info("HTTPS 服务启动完成, 监听端口: %s", serverPort)
	} else {
		logger.Info("HTTP 服务启动完成, 监听端口: %s", serverPort)
	}

	return webServer
}

// gracefulShutdown 实现服务的优雅关闭机制
// @param webServer HTTP 服务器实例，需要被优雅关闭
func gracefulShutdown(webServer *http.Server) {
	// 创建缓冲为1的信号通道，避免信号丢失
	signalChannel := make(chan os.Signal, 1)

	// 注册需要监听的系统信号
	// SIGTERM: kill 命令发送的终止信号（优雅关闭）
	// SIGINT: Ctrl+C 产生的中断信号（用户主动停止）
	// 注意: SIGKILL(-9) 信号无法被捕获，会强制终止进程
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待系统信号，程序将在此处暂停直到接收到信号
	<-signalChannel

	// 创建带超时的上下文，10秒内必须完成所有关闭操作
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 按照依赖关系的逆序关闭组件，确保数据一致性

	// 第一步: 关闭数据存储层（数据库连接池、缓存系统等）
	// 优先关闭数据层，确保所有数据写入完成
	logger.Info("关闭数据层")
	storage.Storage.Close(shutdownCtx)
	logger.Info("数据层已关闭")

	// 第二步: 关闭搜索引擎，停止索引操作
	logger.Info("关闭搜索引擎")
	searchengine.CloseIndex()
	logger.Info("搜索引擎已关闭")

	// 第三步: 优雅关闭 Web 服务器，停止接受新请求并等待现有请求完成
	logger.Info("正在关闭服务")
	if err := webServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("服务关闭超时, 已强制关闭: ", err)
	}
	logger.Info("服务已退出")
}

// parseCommandLineArgs 解析命令行参数并设置默认值
func parseCommandLineArgs() {
	// 初始化全局参数存储映射
	Args = make(map[string]string)

	// 遍历命令行参数，查找 --env 标志
	for i := 0; i < len(os.Args); i++ {
		// 检查当前参数是否为 --env 且存在对应的值
		if os.Args[i] == "--env" && i+1 < len(os.Args) {
			Args["env"] = os.Args[i+1]
			i++ // 跳过已处理的环境值参数
		}
	}

	// 如果未指定环境参数，设置默认值为生产环境
	if _, exists := Args["env"]; !exists {
		Args["env"] = env.ProdEnv
	}
}

// main 是应用程序的主入口函数，负责完整的生命周期管理
func main() {
	// 阶段1: 解析命令行参数，获取运行环境等配置
	parseCommandLineArgs()

	// 阶段2: 加载 YAML 配置文件，初始化全局配置
	config.LoadConfig()

	// 阶段3: 初始化应用程序核心组件，设置1分钟超时
	// 包括日志系统、数据存储层、搜索引擎等关键组件
	initializationCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	initializeApplicationComponents(initializationCtx)
	cancel() // 及时释放上下文资源

	// 阶段4: 根据运行环境设置 Gin 框架模式
	// 生产环境使用 Release 模式以获得最佳性能
	if env.CurrentEnv == env.ProdEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	// 阶段5: 启动 Web 服务器，开始处理 HTTP 请求
	webServer := startWebServer()

	// 阶段6: 进入信号监听状态，等待优雅关闭信号
	// 程序将在此处阻塞，直到接收到 SIGINT 或 SIGTERM 信号
	gracefulShutdown(webServer)
}
