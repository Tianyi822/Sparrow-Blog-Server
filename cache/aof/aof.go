package aof

import (
	"h2blog/cache/libs"
	"h2blog/pkg/config"
	"sync"
)

type Aof struct {
	File libs.FileOp
	mu   sync.RWMutex
}

func NewAof(filePath string) Aof {
	foConfig := libs.FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}
	return Aof{
		File: libs.CreateFileOp(foConfig),
	}
}
