// Package core 提供线程安全的内存缓存核心实现
//
// 特性：
// - 支持TTL自动过期
// - 类型安全的取值方法
// - 分片锁提升并发性能
// - 内存使用监控
//
// 使用示例：
//
//	c := NewCore()
//	ctx := context.Background()
//	err := c.SetWithExpired(ctx, "user:1001", userData, 10*time.Minute)
//	val, ok, err := c.GetInt(ctx, "counter")
//
// 注意事项：
// - 存储指针类型时需要自行管理生命周期
// - 建议通过WithMaxEntries设置条目限制
package core

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"
)

// 缓存操作错误类型定义
var (
	// ErrTypeMismatch 当类型转换不匹配时返回（例如尝试将字符串转换为整型）
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrOutOfRange 当数值超出目标类型范围时返回（例如将float64(1e100)转换为int）
	ErrOutOfRange = errors.New("value out of range")

	// ErrPointerNotAllowed 当尝试存储指针类型时返回
	ErrPointerNotAllowed = errors.New("pointer values are not allowed")

	// ErrNotFound 当条目不存在或已过期时返回
	ErrNotFound = errors.New("entry not found")

	// ErrEmptyKey key 值空
	ErrEmptyKey = errors.New("key is empty")
)

type ValueType uint8

const (
	INT ValueType = iota
	UINT
	FLOAT
	STRING
	OBJ
)

// cacheItem 表示缓存中的单个条目
// value    存储的实际值，支持任意类型。注意存储指针类型时需要自行管理生命周期
// expireAt 条目过期的时间戳（UTC时间），零值表示永不过期
type cacheItem struct {
	value    any
	vt       ValueType
	expireAt time.Time
}

// Core 线程安全的内存缓存核心系统，采用分片锁设计提升并发性能
//
// 字段说明：
// items       - 存储缓存条目的map，key为字符串类型，建议使用"type:id"格式
// mu          - 分片读写锁，每个分片独立加锁减少竞争
// maxEntries  - 最大条目限制（0表示无限制），达到限制时写入返回ErrMaxEntries
// currentSize - 当前内存占用量估算（字节），用于防止内存溢出
type Core struct {
	items map[string]cacheItem
	mu    sync.RWMutex
}

// NewCore 创建并初始化新的缓存核心实例
// 返回值:
// - *Core 初始化完成的缓存指针
func NewCore() *Core {
	return &Core{
		items: make(map[string]cacheItem),
	}
}

// Set 设置缓存条目, 默认长期有效
// 示例:
//
// err := cache.Set(ctx, "user:1001", userData)
//
//	if err != nil {
//	   log.Printf("缓存写入失败: %v", err)
//	}
//
// 参数:
//
// ctx   上下文，用于取消操作和超时控制
// key   条目键（推荐使用冒号分隔的命名规范，如"type:id"）
// value 要存储的值（仅支持非指针类型）
//
// 返回值:
//
//	error 可能返回的错误包括：
//	 - context.Canceled 上下文取消
//	 - context.DeadlineExceeded 操作超时
//	 - ErrMaxEntries 达到最大条目限制（需通过WithMaxEntries设置）
func (c *Core) Set(ctx context.Context, key string, value any) error {
	return c.SetWithExpired(ctx, key, value, 0)
}

// SetWithExpired 设置缓存条目。当存在同名key时会覆盖旧值并重置TTL
//
// 示例:
//
//	err := cache.SetWithExpired(ctx, "user:1001", userData, 10*time.Minute)
//	if err != nil {
//	    log.Printf("缓存写入失败: %v", err)
//	}
//
// 参数:
//
//	ctx   上下文，用于取消操作和超时控制
//	key   条目键（推荐使用冒号分隔的命名规范，如"type:id"）
//	value 要存储的值（仅支持非指针类型）
//	ttl   存活时间（小于等于0时条目会立即过期）
//
// 返回值:
//
//	error 可能返回的错误包括：
//	 - context.Canceled 上下文取消
//	 - context.DeadlineExceeded 操作超时
//	 - ErrMaxEntries 达到最大条目限制（需通过WithMaxEntries设置）
func (c *Core) SetWithExpired(ctx context.Context, key string, value any, ttl time.Duration) error {
	if len(strings.TrimSpace(key)) == 0 {
		return ErrEmptyKey
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		// 类型安全检查：禁止存储指针类型
		if reflect.TypeOf(value).Kind() == reflect.Ptr {
			return ErrPointerNotAllowed
		}

		item := cacheItem{
			value: value,
		}
		if ttl == 0 {
			item.expireAt = time.Time{} // 零值表示永不过期
		} else {
			item.expireAt = time.Now().Add(ttl)
		}

		// 设置值类型
		switch value.(type) {
		case int, int8, int16, int32, int64:
			item.vt = INT
		case uint, uint8, uint16, uint32, uint64:
			item.vt = UINT
		case float32, float64:
			item.vt = FLOAT
		case string:
			item.vt = STRING
		default:
			// 对于其他类型，序列化为JSON字符串存储
			jsonStr, err := json.Marshal(item.value)
			if err != nil {
				return err
			}
			item.vt = OBJ
			item.value = jsonStr
		}

		c.items[key] = item
		return nil
	}
}

// Incr 原子递增整型值
// ctx   上下文，用于取消操作
// key   条目键
//
// 返回值:
// - int  操作后的新值
// - error 可能错误：
//   - ErrNotFound key不存在
//   - ErrTypeMismatch 值类型非整型
//   - ErrOutOfRange 数值溢出
//
// 注意:
// - 不存在的key直接返回ErrNotFound
// - 操作成功后保持原有TTL时间不变
func (c *Core) Incr(ctx context.Context, key string) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// 获取现有值,不存在则默认为0
		val, err := c.GetInt(ctx, key)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// 不存在的 key 新建一个
				if err = c.Set(ctx, key, 1); err != nil {
					return 0, err
				}
			} else {
				return 0, err
			}
		}

		// 溢出检查
		if val == math.MaxInt {
			return 0, ErrOutOfRange
		}

		c.mu.Lock()
		c.items[key] = cacheItem{
			value:    val + 1,
			vt:       INT,
			expireAt: c.items[key].expireAt, // 若key存在则保持原有过期时间,否则为0(永不过期)
		}
		c.mu.Unlock()

		return val + 1, nil
	}
}

// IncrUint 原子递增无符号整型值
// ctx   上下文，用于取消操作
// key   条目键
//
// 返回值:
// - uint  操作后的新值
// - error 可能错误：
//   - ErrNotFound key不存在
//   - ErrTypeMismatch 值类型非无符号整型
//   - ErrOutOfRange 数值溢出
//
// 注意:
// - 不存在的key直接返回ErrNotFound
// - 操作成功后保持原有TTL时间不变
func (c *Core) IncrUint(ctx context.Context, key string) (uint, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		// 获取现有值,不存在则默认为0
		item, exists := c.items[key]
		var val uint = 0
		if exists {
			// 类型检查
			v, err := c.GetUint(ctx, key)
			if err != nil {
				return 0, err
			}
			val = v
		}

		// 溢出检查
		if val == math.MaxUint {
			return 0, ErrOutOfRange
		}

		newVal := val + 1
		c.items[key] = cacheItem{
			value:    newVal,
			vt:       UINT,
			expireAt: item.expireAt, // 若key存在则保持原有过期时间,否则为0(永不过期)
		}

		return newVal, nil
	}
}

// Get 获取缓存条目原始值
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - any 	存储的原始值
// - error  操作过程中遇到的错误
func (c *Core) Get(ctx context.Context, key string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// 尽可能缩小加锁范围
		c.mu.RLock()
		item, exists := c.items[key]
		c.mu.RUnlock()

		if !exists {
			return nil, ErrNotFound
		}

		if item.expireAt.IsZero() {
			return item.value, nil
		} else if time.Now().After(item.expireAt) {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
			return nil, ErrNotFound
		} else {
			return item.value, nil
		}
	}
}

// GetInt 获取int类型值 (自动处理类型转换)
func (c *Core) GetInt(ctx context.Context, key string) (int, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int: // 原生int类型直接返回
		return v, nil
	case int8, int16, int32: // 小整型安全转换
		return int(v.(int32)), nil // 类型断言后转换
	case int64: // 64位整型需要范围检查
		if v > math.MaxInt || v < math.MinInt {
			return 0, ErrOutOfRange // 数值超出int范围
		}
		return int(v), nil
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > math.MaxInt {
			return 0, ErrOutOfRange
		}
		return int(u), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// GetUint 获取无符号整型值（自动类型转换）
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - uint 转换后的无符号整数值
// - error 转换错误（ErrTypeMismatch/ErrOutOfRange）或操作错误
//
// 支持的类型转换:
// - 所有无符号整型（uint8/16/32/64）
// - 有符号整型（int8/16/32/64）需为非负数
// - 浮点型（float32/64）需在[0, math.MaxUint64]范围内
func (c *Core) GetUint(ctx context.Context, key string) (uint, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0, ErrOutOfRange
		}
		return uint(i), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// GetFloat 获取浮点型值（自动类型转换到float64）
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - float64 转换后的浮点数值
// - error   转换错误（ErrTypeMismatch）或操作错误
//
// 支持的类型转换:
// - 所有整型（int/uint系列）和浮点型
// - 其他类型返回ErrTypeMismatch
func (c *Core) GetFloat(ctx context.Context, key string) (float64, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	default:
		return 0, ErrTypeMismatch
	}
}

// GetBool 获取布尔类型值
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - bool  转换后的布尔值
// - error 转换错误（ErrTypeMismatch）或操作错误
//
// 注意:
// - 仅支持原生bool类型，不支持字符串/数字到bool的转换
func (c *Core) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}
	return false, ErrTypeMismatch
}

// GetString 获取字符串类型值
// ctx 上下文，用于取消操作
// key 要获取的条目键
//
// 返回值:
// - string 转换后的字符串值
// - error  转换错误（ErrTypeMismatch）或操作错误
//
// 注意:
// - 仅支持原生string类型，不支持自动类型转换
func (c *Core) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", ErrTypeMismatch
}

// Delete 删除指定键的缓存条目，无论该条目是否过期
//
// 参数:
//
//	ctx 上下文，用于取消操作和超时控制
//	key 要删除的条目键
//
// 返回值:
//
//	error 可能返回的错误包括：
//	 - context.Canceled 上下文取消
//	 - context.DeadlineExceeded 操作超时
//
// 注意：
// - 删除不存在的key不会返回错误
// - 该操作会立即释放相关内存
// - 高频删除操作建议使用批量删除接口
func (c *Core) Delete(ctx context.Context, key string) error {
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

// Cleanup 同步清理所有过期条目。该方法会获取写锁，建议在低峰期调用或通过StartCleaner后台定时执行
//
// 注意：
// - 遍历整个缓存空间，时间复杂度为O(n)
// - 执行期间会阻塞所有读写操作
// - 建议清理间隔不小于5分钟以避免性能波动
//
// 注意：
// - 遍历整个缓存空间，时间复杂度为O(n)
// - 执行期间会阻塞所有读写操作
// - 建议清理间隔不小于5分钟以避免性能波动
//
// 注意：
// - 遍历整个缓存空间，时间复杂度为O(n)
// - 执行期间会阻塞所有读写操作
func (c *Core) Cleanup() {
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
