package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"h2blog/pkg/config"
	"os"
	"strings"
	"sync"
)

var (
	logger     *zap.SugaredLogger
	loggerLock sync.Once
)

// InitLogger 初始化日志系统
func InitLogger() error {
	var err error
	loggerLock.Do(func() {
		err = initLogger()
	})
	return err
}

func GetLogger() *zap.SugaredLogger {
	return logger
}

// initLogger 实际的初始化逻辑
func initLogger() error {
	loggerConf := config.LoggerConfig
	if loggerConf == nil {
		return fmt.Errorf("logger config is nil")
	}

	// 创建日志写入器
	// TODO: 后面根据运行环境，需要对输出进行判断，现在暂时不动
	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}

	// 添加文件输出
	if loggerConf.Path != "" {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   loggerConf.Path,
			MaxSize:    loggerConf.MaxSize,
			MaxBackups: loggerConf.MaxBackups,
			MaxAge:     loggerConf.MaxAge,
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

// 简单的日志接口
func logf(level func(template string, args ...any), format string, args ...any) {
	if logger == nil {
		if err := InitLogger(); err != nil {
			fmt.Printf("Failed to initialize logger: %v\n", err)
			return
		}
	}
	level(format, args...)
}

func Debug(format string, args ...any) {
	logf(logger.Debugf, format, args...)
}

func Info(format string, args ...any) {
	logf(logger.Infof, format, args...)
}

func Warn(format string, args ...any) {
	logf(logger.Warnf, format, args...)
}

func Error(format string, args ...any) {
	logf(logger.Errorf, format, args...)
}

func Panic(format string, args ...any) {
	logf(logger.Panicf, format, args...)
}

func Fatal(format string, args ...any) {
	logf(logger.Fatalf, format, args...)
}
