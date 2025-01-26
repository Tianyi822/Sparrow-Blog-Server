package aof

import (
	"h2blog/pkg/config"
	"sync"
)

type Aof struct {
	File FileOp
	mu   sync.RWMutex
}

func NewAof(filePath string) Aof {
	foConfig := FoConfig{
		NeedCompress: config.CacheConfig.Aof.Compress,
		Path:         config.CacheConfig.Aof.Path,
		MaxSize:      config.CacheConfig.Aof.MaxSize,
	}
	return Aof{
		File: CreateFileOp(foConfig),
	}
}
