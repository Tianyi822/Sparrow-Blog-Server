package main

import (
	"context"
	"errors"
	"fmt"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/pkg/markdown"
	"h2blog/pkg/webp"
	"h2blog/routers"
	blogrouters "h2blog/routers/blogRouters"
	imgrouters "h2blog/routers/imgRouters"
	"h2blog/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// loadComponent 加载基础组件
func loadComponent() {
	// 从指定路径加载配置信息
	config.LoadConfig("resources/config/test/web-config.yaml")
	// 初始化日志组件
	err := logger.InitLogger()
	if err != nil {
		panic("日志模块初始化失败，请检查配置文件是否有误")
	}
	// 初始化数据层
	storage.InitStorage()
	// 初始化 Markdown 渲染器
	markdown.InitRenderer()
	// 初始化图片转换器
	webp.InitConverter()
}

// runServer 启动服务
func runServer() *http.Server {
	logger.Info("加载路由信息")
	routers.IncludeOpts(blogrouters.Routers) // 添加博客路由
	routers.IncludeOpts(imgrouters.Routers)  // 添加图片路由
	logger.Info("路由信息加载完成")

	logger.Info("配置路由")
	r := routers.InitRouter()
	logger.Info("路由配置完成")

	logger.Info("启动服务中")
	port := fmt.Sprintf(":%v", config.ServerConfig.Port)
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

// listenSysSignal 监听系统信号，优雅关闭服务
func listenSysSignal(srv *http.Server) {
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

	// 先关闭数据层连接
	logger.Info("关闭数据库连接")
	storage.Storage.CloseDbConnect(ctx)
	logger.Info("数据库连接已关闭")

	// 关闭图片压缩器
	logger.Info("关闭图片压缩器")
	webp.Converter.Shutdown()
	logger.Info("图片压缩器已关闭")

	logger.Info("正在关闭服务")
	// 定时优雅关闭服务（将未处理完的请求处理完再关闭服务），超时就退出
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("服务关闭超时, 哥们已强制关闭: ", err)
	}
	logger.Info("服务已退出")
}

func main() {
	// 加载基础组件
	loadComponent()

	// 启动服务
	srv := runServer()

	// 监听系统信号
	listenSysSignal(srv)
}
