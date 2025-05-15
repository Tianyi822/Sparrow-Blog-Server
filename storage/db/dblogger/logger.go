package dblogger

import (
	"context"
	"errors"
	"sparrow_blog_server/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DbLogger 是一个实现了 gormlogger.Interface 接口的日志记录器，实现自定义数据库日志记录
type DbLogger struct {
	zapLogger     *zap.SugaredLogger
	LogLevel      gormlogger.LogLevel
	SlowThreshold time.Duration
}

func NewDbLogger() *DbLogger {
	return &DbLogger{
		zapLogger:     logger.GetLogger(),
		LogLevel:      gormlogger.Info,
		SlowThreshold: 200 * time.Millisecond,
	}
}

// LogMode 设置日志级别
func (l *DbLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info 打印信息
func (l *DbLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if l.LogLevel >= gormlogger.Info {
		l.zapLogger.Infof(msg, data...)
	}
}

// Warn 打印警告信息
func (l *DbLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if l.LogLevel >= gormlogger.Warn {
		l.zapLogger.Warnf(msg, data...)
	}
}

// Error 打印错误信息
func (l *DbLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if l.LogLevel >= gormlogger.Error {
		l.zapLogger.Errorf(msg, data...)
	}
}

// Trace 打印SQL执行信息
func (l *DbLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 根据错误级别打印日志
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.zapLogger.Errorf("[%.3fms] [rows:%v] %s; %s", float64(elapsed.Nanoseconds())/1e6, rows, sql, err)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		l.zapLogger.Warnf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	default:
		l.zapLogger.Debugf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
	}
}
