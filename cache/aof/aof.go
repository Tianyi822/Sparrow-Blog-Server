package aof

import (
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

func NewAof(filePath string) Aof {
	foConfig := FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}
	return Aof{
		file: CreateFileOp(foConfig),
	}
}

func (aof *Aof) LoadFile() ([][]string, error) {

	foConfig := FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}

	fileOp := CreateFileOp(foConfig)
	if err := fileOp.ready(); err != nil {
		msg := fmt.Sprintf("aof 文件准备失败: %s", err.Error())
		logger.Error(msg)
		return nil, err
	}

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
	scanner := fileOp.GetScanner()
	var commands [][]string
	for scanner.Scan() {
		command := strings.Split(scanner.Text(), ";;")
		switch strings.TrimSpace(command[0]) {
		case core.SET:
			if len(command) != 5 {
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
				continue
			}
			// 保存 DELETE 命令
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]),
			})
		case core.INCR:
			if len(command) != 3 {
				continue
			}
			// 保存 INCR 命令
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
				strings.TrimSpace(command[1]),
				strings.TrimSpace(command[2]),
			})
		case core.CLEANUP:
			// 保存 CLEANUP 命令
			commands = append(commands, []string{
				strings.TrimSpace(command[0]),
			})
		}
	}

	return commands, nil
}
