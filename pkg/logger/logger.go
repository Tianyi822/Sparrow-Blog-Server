package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"sparrow_blog_server/env"
	"sparrow_blog_server/pkg/config"
	"strings"
	"sync"
)

var (
	logger     *zap.SugaredLogger
	loggerLock sync.Once
)

// InitLogger 初始化日志系统
func InitLogger(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		var err error
		loggerLock.Do(func() {
			err = initLogger()
		})
		return err
	}
}

func GetLogger() *zap.SugaredLogger {
	return logger
}

// initLogger 实际的初始化逻辑
func initLogger() error {
	loggerConf := config.Logger
	if len(loggerConf.Path) == 0 {
		return fmt.Errorf("日志配置路径为空，请检查配置文件是否有误")
	}

	// 创建日志文件写入器
	var writers = make([]zapcore.WriteSyncer, 0)
	// 创建日志写入器
	if env.CurrentEnv == env.DebugEnv {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 添加文件输出
	if loggerConf.Path != "" {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   loggerConf.Path,
			MaxSize:    int(loggerConf.MaxSize),
			MaxBackups: int(loggerConf.MaxBackups),
			MaxAge:     int(loggerConf.MaxAge),
			Compress:   loggerConf.Compress,
		})
		writers = append(writers, fileWriter)
	}

	// 创建多重写入器
	writeSyncer := zapcore.NewMultiWriteSyncer(writers...)

	// 创建编码器
	encoder := getEncoder()

	// 创建核心记录器
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		getLogLevel(loggerConf.Level),
	)

	// 创建Logger
	zapLog := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	logger = zapLog.Sugar()
	return nil
}

// getEncoder 获取日志编码器
func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func Debug(format string, args ...any) {
	logger.Debugf(format, args...)
}

func Info(format string, args ...any) {
	logger.Infof(format, args...)
}

func Warn(format string, args ...any) {
	logger.Warnf(format, args...)
}

func Error(format string, args ...any) {
	logger.Errorf(format, args...)
}

func Panic(format string, args ...any) {
	logger.Fatalf(format, args...)
}

func Fatal(format string, args ...any) {
	logger.Fatalf(format, args...)
}
