package cache

import (
	"context"
	"errors"
	"math"
	"reflect"
	"sync"
	"time"
)

// cacheItem 表示缓存中的单个条目
// value    存储的实际值，支持任意类型
// expireAt 条目过期的时间戳（UTC时间）
type cacheItem struct {
	value    any
	expireAt time.Time
}

// Cache 线程安全的内存缓存系统
// items 使用map存储所有缓存条目，key为字符串类型
// mu     读写锁保证并发安全
type Cache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

// NewCache 创建并初始化新的缓存实例
// 返回值:
// - *Cache 初始化完成的缓存指针
func NewCache() *Cache {
	return &Cache{
		items: make(map[string]cacheItem),
	}
}

// Set 设置缓存条目
// ctx   上下文，用于取消操作
// key   条目的唯一标识键
// value 要存储的值（支持任意类型）
// ttl   条目的存活时间（time.Duration类型）
//
// 返回值:
// - error 操作过程中遇到的错误（包括上下文取消）
func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		c.items[key] = cacheItem{
			value:    value,
			expireAt: time.Now().Add(ttl),
		}
		return nil
	}
}

// Get 获取缓存条目原始值
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - any 	存储的原始值
// - bool   是否找到有效条目
// - error  操作过程中遇到的错误
func (c *Cache) Get(ctx context.Context, key string) (any, bool, error) {
	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	default:
		c.mu.RLock()
		defer c.mu.RUnlock()

		item, exists := c.items[key]
		if !exists {
			return nil, false, nil
		}

		if time.Now().After(item.expireAt) {
			delete(c.items, key)
			return nil, false, nil
		}

		return item.value, true, nil
	}
}

// 新增类型安全访问方法
var (
	ErrTypeMismatch = errors.New("type mismatch")
	ErrOutOfRange   = errors.New("value out of range")
)

// GetInt 获取int类型值 (自动处理类型转换)
func (c *Cache) GetInt(ctx context.Context, key string) (int, bool, error) {
	val, ok, err := c.Get(ctx, key)
	if !ok || err != nil {
		return 0, ok, err
	}

	switch v := val.(type) {
	case int: // 原生int类型直接返回
		return v, true, nil
	case int8, int16, int32: // 小整型安全转换
		return int(v.(int32)), true, nil // 类型断言后转换
	case int64: // 64位整型需要范围检查
		if v > math.MaxInt || v < math.MinInt {
			return 0, true, ErrOutOfRange // 数值超出int范围
		}
		return int(v), true, nil
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > math.MaxInt {
			return 0, true, ErrOutOfRange
		}
		return int(u), true, nil
	case float32:
		if v > math.MaxInt || v < math.MinInt {
			return 0, true, ErrOutOfRange
		}
		return int(v), true, nil
	case float64:
		if v > math.MaxInt || v < math.MinInt {
			return 0, true, ErrOutOfRange
		}
		return int(v), true, nil
	default:
		return 0, true, ErrTypeMismatch
	}
}

// GetUint 获取无符号整型值（自动类型转换）
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - uint 转换后的无符号整数值
// - bool 是否找到有效条目
// - error 转换错误（ErrTypeMismatch/ErrOutOfRange）或操作错误
//
// 支持的类型转换:
// - 所有无符号整型（uint8/16/32/64）
// - 有符号整型（int8/16/32/64）需为非负数
// - 浮点型（float32/64）需在[0, math.MaxUint64]范围内
func (c *Cache) GetUint(ctx context.Context, key string) (uint, bool, error) {
	val, ok, err := c.Get(ctx, key)
	if !ok || err != nil {
		return 0, ok, err
	}

	switch v := val.(type) {
	case uint:
		return v, true, nil
	case uint8:
		return uint(v), true, nil
	case uint16:
		return uint(v), true, nil
	case uint32:
		return uint(v), true, nil
	case uint64:
		return uint(v), true, nil
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0, true, ErrOutOfRange
		}
		return uint(i), true, nil
	case float32: // 32位浮点数需要双重检查
		// 先转换为float64进行精确范围验证
		if v < 0 || float64(v) > math.MaxUint64 {
			return 0, true, ErrOutOfRange // 值超出无符号整型范围
		}
		return uint(v), true, nil // 安全转换为uint
	case float64: // 64位浮点数直接检查
		if v < 0 || v > math.MaxUint64 {
			return 0, true, ErrOutOfRange // 值超出无符号整型范围
		}
		return uint(v), true, nil // 直接转换（可能丢失小数部分）
	default:
		return 0, true, ErrTypeMismatch
	}
}

// GetFloat 获取浮点型值（自动类型转换到float64）
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - float64 转换后的浮点数值
// - bool    是否找到有效条目
// - error   转换错误（ErrTypeMismatch）或操作错误
//
// 支持的类型转换:
// - 所有整型（int/uint系列）和浮点型
// - 其他类型返回ErrTypeMismatch
func (c *Cache) GetFloat(ctx context.Context, key string) (float64, bool, error) {
	val, ok, err := c.Get(ctx, key)
	if !ok || err != nil {
		return 0, ok, err
	}

	switch v := val.(type) {
	case float32:
		return float64(v), true, nil
	case float64:
		return v, true, nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), true, nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), true, nil
	default:
		return 0, true, ErrTypeMismatch
	}
}

// GetBool 获取布尔类型值
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - bool  转换后的布尔值
// - bool  是否找到有效条目
// - error 转换错误（ErrTypeMismatch）或操作错误
//
// 注意:
// - 仅支持原生bool类型，不支持字符串/数字到bool的转换
func (c *Cache) GetBool(ctx context.Context, key string) (bool, bool, error) {
	val, ok, err := c.Get(ctx, key)
	if !ok || err != nil {
		return false, ok, err
	}

	if b, ok := val.(bool); ok {
		return b, true, nil
	}
	return false, true, ErrTypeMismatch
}

// GetString 获取字符串类型值
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - string 转换后的字符串值
// - bool   是否找到有效条目
// - error  转换错误（ErrTypeMismatch）或操作错误
//
// 注意:
// - 仅支持原生string类型，不支持自动类型转换
func (c *Cache) GetString(ctx context.Context, key string) (string, bool, error) {
	val, ok, err := c.Get(ctx, key)
	if !ok || err != nil {
		return "", ok, err
	}

	if s, ok := val.(string); ok {
		return s, true, nil
	}
	return "", true, ErrTypeMismatch
}

// Delete 删除键
func (c *Cache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.items, key) // 直接删除，即使key不存在也是安全的
		return nil
	}
}

// Cleanup 清理过期的键值对
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now() // 获取当前UTC时间
	// 遍历所有缓存条目
	for key, item := range c.items {
		// 检查过期时间（UTC时间比较）
		if now.After(item.expireAt) {
			// 同步删除过期条目
			delete(c.items, key) // map删除操作
		}
	}
}
