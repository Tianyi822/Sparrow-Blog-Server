package main

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/markdown"
	"h2blog_server/pkg/webp"
	"h2blog_server/routers"
	"h2blog_server/routers/configServer/configRouters"
	blogrouters "h2blog_server/routers/webServer/blogRouters"
	imgrouters "h2blog_server/routers/webServer/imgRouters"
	"h2blog_server/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// loadComponent 加载基础组件
func loadComponent(ctx context.Context) {
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
	// 初始化 Markdown 渲染器
	err = markdown.InitRenderer(ctx)
	if err != nil {
		panic("Markdown 渲染器初始化失败，请检查配置文件是否有误")
	}
	// 初始化图片转换器
	err = webp.InitConverter(ctx)
	if err != nil {
		panic("图片转换器初始化失败，请检查配置文件是否有误")
	}
}

// runServer 启动服务
func runServer() *http.Server {
	logger.Info("加载路由信息")
	routers.IncludeOpts(blogrouters.Routers, imgrouters.Routers) // 添加路由
	logger.Info("路由信息加载完成")

	logger.Info("配置路由")
	r := routers.InitRouter(env.RuntimeEnv)
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

func startConfigServer() *http.Server {
	// 加载配置接口
	routers.IncludeOpts(configRouters.Routers)

	// 初始化路由
	r := routers.InitRouter(env.ConfigServerEnv)

	// 配置 HTTP 服务
	srv := &http.Server{
		Addr:    ":2234",
		Handler: r,
	}

	// 开启一个 goroutine 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Sprintf("配置服务报错: %v", err.Error()))
		}
	}()

	env.CompletedConfigSign = make(chan bool)

	return srv
}

func closeConfigServer(srv *http.Server) {
	// 等待配置完成
	sign := <-env.CompletedConfigSign

	if !sign {
		panic("配置服务关闭出错")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	// 定时优雅关闭服务（将未处理完的请求处理完再关闭服务），超时就退出
	if err := srv.Shutdown(ctx); err != nil {
		panic(fmt.Sprintf("配置服务关闭超时, 哥们已强制关闭: %v", err.Error()))
	}

	cancel()
}

func main() {
	// 优先去本地默认路径加载配置文件
	err := config.LoadConfig()
	var configErr *config.Err
	errors.As(err, &configErr)
	if configErr.IsNoConfigFileErr() {
		// 若未找到配置文件，则单独开启配置服务，与业务端口分开使用
		server := startConfigServer()
		// 等待配置服务关闭
		closeConfigServer(server)
	}

	// 加载基础组件
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	loadComponent(ctx)
	cancel()

	// 启动服务
	srv := runServer()

	// 监听系统信号
	closeWebServer(srv)
}
