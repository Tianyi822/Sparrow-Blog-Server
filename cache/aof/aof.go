package aof

import (
	"context"
	"errors"
	"fmt"
	"h2blog/cache/core"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"strings"
	"sync"
)

type Aof struct {
	file FileOp
	mu   sync.RWMutex
}

// NewAof 创建并返回一个新的 AOF 实例，使用给定的配置
func NewAof() Aof {
	foConfig := FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}
	return Aof{
		file: CreateFileOp(foConfig),
	}
}

// LoadFile 读取并解析 AOF 文件，返回命令字符串切片
// 每个命令都被分割成多个组件（操作类型、键、值等）
func (aof *Aof) LoadFile(ctx context.Context) ([][]string, error) {
	select {
	case <-ctx.Done():
		logger.Warn("加载 AOF 文件被取消")
		return nil, ctx.Err()
	default:
		// 开始扫描 AOF 文件
		// 文件每行数据为: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
		// 例：
		// - SET;;key:int;;1;;int;;0
		// - SET;;key:string;;hello world;;string;;0
		// - DELETE;;key
		// - INCR;;key;;int
		// - INCR;;key;;uint
		// - CLEANUP
		//
		// 注意:
		// - CLEANALL 之所以没有是因为这个操作会清空所有缓存，直接清空所有 AOF 文件，不需要进行保存
		// - CLEANUP 是清理过期数据，这个操作需要保存，因为可能存在过期数据没有及时清理的情况
		scanner, err := aof.file.GetScanner()
		if err != nil {
			logger.Error("获取 AOF 文件扫描器失败: %v", err)
			return nil, err
		}

		aof.mu.Lock()
		defer aof.mu.Unlock()

		logger.Info("开始加载 AOF 文件")
		var commands [][]string
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			command := strings.Split(scanner.Text(), ";;")
			switch strings.TrimSpace(command[0]) {
			case core.SET:
				if len(command) != 5 {
					logger.Warn("第 %d 行 SET 命令格式错误，跳过: %v", lineNum, command)
					continue
				}
				// 保存 SET 命令
				commands = append(commands, []string{
					strings.TrimSpace(command[0]),
					strings.TrimSpace(command[1]),
					strings.TrimSpace(command[2]),
					strings.TrimSpace(command[3]),
					strings.TrimSpace(command[4]),
				})
			case core.DELETE:
				if len(command) != 2 {
					logger.Warn("第 %d 行 DELETE 命令格式错误，跳过: %v", lineNum, command)
					continue
				}
				commands = append(commands, []string{
					strings.TrimSpace(command[0]),
					strings.TrimSpace(command[1]),
				})
			case core.INCR:
				if len(command) != 3 {
					logger.Warn("第 %d 行 INCR 命令格式错误，跳过: %v", lineNum, command)
					continue
				}
				commands = append(commands, []string{
					strings.TrimSpace(command[0]),
					strings.TrimSpace(command[1]),
					strings.TrimSpace(command[2]),
				})
			case core.CLEANUP:
				commands = append(commands, []string{
					strings.TrimSpace(command[0]),
				})
			default:
				logger.Warn("第 %d 行包含未知命令，跳过: %v", lineNum, command[0])
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Error("扫描 AOF 文件出错: %v", err)
			return nil, fmt.Errorf("扫描文件出错: %w", err)
		}

		logger.Info("AOF 文件加载完成，共读取 %d 行，有效命令 %d 个", lineNum, len(commands))
		return commands, nil
	}
}

// Store 保存命令到 AOF 文件
// 文件每行数据为: OPERATE;;KEY;;VALUE;;VALUETYPE;;EXPIRED
// 例：
// - SET;;key:int;;1;;int;;0
// - SET;;key:string;;hello world;;string;;0
// - DELETE;;key
// - INCR;;key;;int
// - INCR;;key;;uint
// - CLEANUP
func (aof *Aof) Store(ctx context.Context, cmd string, args ...string) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	select {
	case <-ctx.Done():
		logger.Warn("保存命令被取消")
		return ctx.Err()
	default:
		switch cmd {
		case core.SET:
			if len(args) != 4 {
				msg := fmt.Sprintf("SET 命令需要4个参数 (key=%s, value=%s, type=%s, expired=%s)，但收到 %d 个",
					safeGet(args, 0), safeGet(args, 1), safeGet(args, 2), safeGet(args, 3), len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("保存 SET 命令: key=%s, type=%s", args[0], args[2])
			return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s;;%s;;%s", cmd, args[0], args[1], args[2], args[3])))
		case core.DELETE:
			if len(args) != 1 {
				msg := fmt.Sprintf("DELETE 命令需要1个参数 (key=%s)，但收到 %d 个", safeGet(args, 0), len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("保存 DELETE 命令: key=%s", args[0])
			return aof.file.Write([]byte(fmt.Sprintf("%s;;%s", cmd, args[0])))
		case core.INCR:
			if len(args) != 2 {
				msg := fmt.Sprintf("INCR 命令需要2个参数 (key=%s, type=%s)，但收到 %d 个",
					safeGet(args, 0), safeGet(args, 1), len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("保存 INCR 命令: key=%s, type=%s", args[0], args[1])
			return aof.file.Write([]byte(fmt.Sprintf("%s;;%s;;%s", cmd, args[0], args[1])))
		case core.CLEANUP:
			if len(args) != 0 {
				msg := fmt.Sprintf("CLEANUP 命令不需要参数，但收到 %d 个", len(args))
				logger.Error(msg)
				return errors.New(msg)
			}
			logger.Debug("保存 CLEANUP 命令")
			return aof.file.Write([]byte(cmd))
		default:
			msg := fmt.Sprintf("不支持的命令类型: %s", cmd)
			logger.Error(msg)
			return errors.New(msg)
		}
	}
}

// safeGet 安全获取切片元素，避免越界
func safeGet(slice []string, index int) string {
	if index < 0 || index >= len(slice) {
		return "<nil>"
	}
	return slice[index]
}
