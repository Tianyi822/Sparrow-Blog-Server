package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"sparrow_blog_server/env"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/routers"
	"sparrow_blog_server/routers/adminrouter"
	"sparrow_blog_server/routers/webrouter"
	"sparrow_blog_server/searchengine"
	"sparrow_blog_server/storage"
	"syscall"
	"time"
)

var Args map[string]string

// loadComponent 加载基础组件
func loadComponent(ctx context.Context) {
	// 设置当前环境
	env.CurrentEnv = Args["env"]
	// 初始化日志组件
	err := logger.InitLogger(ctx)
	if err != nil {
		panic("日志模块初始化失败，请检查配置文件是否有误")
	}
	// 初始化数据层
	err = storage.InitStorage(ctx)
	if err != nil {
		panic("数据层初始化失败，请检查配置文件是否有误")
	}
	// 加载搜索引擎
	err = searchengine.LoadingIndex(ctx)
	if err != nil {
		panic("搜索引擎初始化失败，请检查配置文件是否有误")
	}
}

// runServer 启动服务
func runServer() *http.Server {
	logger.Info("加载路由信息")
	routers.IncludeOpts(
		webrouter.Router,
		adminrouter.Routers,
	)
	logger.Info("路由信息加载完成")

	logger.Info("配置路由")
	r := routers.InitRouter()
	logger.Info("路由配置完成")

	logger.Info("启动服务中")
	port := fmt.Sprintf(":%v", config.Server.Port)
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}
	// 开启一个goroutine启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("监听端口: %s\n", err)
		}
	}()
	logger.Info("启动服务完成, 监听端口: %s", port)

	return srv
}

// closeWebServer 监听系统信号，优雅关闭服务
func closeWebServer(srv *http.Server) {
	// 创建一个接收信号的通道
	quit := make(chan os.Signal, 1)

	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 在此停顿
	<-quit

	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 先关闭数据层
	logger.Info("关闭数据层")
	storage.Storage.Close(ctx)
	logger.Info("数据层已关闭")

	// 关闭搜索引擎
	logger.Info("关闭搜索引擎")
	searchengine.CloseIndex()
	logger.Info("搜索引擎已关闭")

	logger.Info("正在关闭服务")
	// 定时优雅关闭服务（将未处理完的请求处理完再关闭服务），超时就退出
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务关闭超时, 哥们已强制关闭: ", err)
	}
	logger.Info("服务已退出")
}

// getArgsFromTerminal 从终端获取参数
func getArgsFromTerminal() {
	Args = make(map[string]string)

	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--env" {
			Args["env"] = os.Args[i+1]
			i++
		}
	}

	if _, ok := Args["env"]; !ok {
		Args["env"] = env.ProdEnv
	}
}

func main() {
	getArgsFromTerminal()

	// 加载配置文件
	config.LoadConfig()

	// 加载基础组件
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	loadComponent(ctx)
	cancel()

	// 选择启动模式
	if env.CurrentEnv == env.ProdEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	// 启动服务
	srv := runServer()

	// 监听系统信号
	closeWebServer(srv)
}
