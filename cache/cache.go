package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"h2blog_server/cache/aof"
	"h2blog_server/cache/common"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// cacheItem 表示缓存中的单个条目
type cacheItem struct {
	value    any              // 实际存储的值
	vt       common.ValueType // 存储值的类型信息
	expireAt time.Time        // 过期时间戳（零值表示永不过期）
}

// Cache 实现了一个带分片锁的线程安全内存缓存系统
// 用于提高并发性能。
//
// 字段:
// - items: 存储缓存条目的映射，使用字符串键（推荐格式："type:id"）
// - mu: 用于线程安全操作的读写锁
// - aof: 用于持久化支持的追加文件
type Cache struct {
	items map[string]cacheItem
	mu    sync.RWMutex
	aof   *aof.Aof
}

// NewCache 创建并初始化一个新的缓存实例，使用给定的上下文
// 如果在Cache-config.yaml中配置了AOF，则启用持久化
func NewCache(ctx context.Context) (*Cache, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	c := &Cache{
		items: make(map[string]cacheItem),
	}

	// 如果配置了AOF，则启用
	if config.Cache.Aof.Enable {
		c.aof = aof.NewAof()
		// 从AOF文件加载数据
		if err := c.loadAof(ctx); err != nil {
			// AOF加载失败是致命的
			panic(err)
		}
	}

	return c, nil
}

// loadAof 从AOF文件加载并重放命令以恢复缓存状态
// 它按时间顺序处理SET、DELETE和CLEANUP命令
func (c *Cache) loadAof(ctx context.Context) error {
	if c.aof == nil {
		return nil
	}

	commands, err := c.aof.LoadFile(ctx)
	if err != nil {
		return err
	}

	// 处理命令前检查上下文
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// 重放所有命令
		for _, cmd := range commands {
			// 定期检查上下文
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				switch cmd[0] {
				case common.SET:
					if len(cmd) != 5 {
						continue
					}
					// 解析过期时间
					var expireAt time.Time
					if cmd[4] != "0" {
						expireTs, err := strconv.ParseInt(cmd[4], 10, 64)
						if err != nil {
							continue
						}
						expireAt = time.Unix(expireTs, 0)
					}

					// 解析值类型
					vt, err := strconv.ParseUint(cmd[3], 10, 8)
					if err != nil {
						continue
					}

					// 创建缓存项
					item := cacheItem{
						value:    cmd[2],               // 值
						vt:       common.ValueType(vt), // 转换为ValueType
						expireAt: expireAt,
					}
					c.items[cmd[1]] = item

				case common.DELETE:
					if len(cmd) != 2 {
						continue
					}
					delete(c.items, cmd[1])

				case common.CLEANUP:
					c.Cleanup()
				}
			}
		}
	}

	return nil
}

// Set stores a value in the Cache with no expiration time
func (c *Cache) Set(ctx context.Context, key string, value any) error {
	return c.SetWithExpired(ctx, key, value, 0)
}

// SetWithExpired stores a value in the Cache with an optional TTL
// If the key exists, it will be overwritten and the TTL will be reset
func (c *Cache) SetWithExpired(ctx context.Context, key string, value any, ttl time.Duration) error {
	if len(strings.TrimSpace(key)) == 0 {
		return ErrEmptyKey
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		// Type safety check: Disallow storage of pointer, array, or slice types
		if reflect.TypeOf(value).Kind() == reflect.Ptr ||
			reflect.TypeOf(value).Kind() == reflect.Array ||
			reflect.TypeOf(value).Kind() == reflect.Slice {
			return ErrPointerNotAllowed
		}

		item := cacheItem{
			value: value,
		}

		// Set expiration time
		var expireTs int64
		if ttl > 0 {
			item.expireAt = time.Now().Add(ttl)
			expireTs = item.expireAt.Unix()
		}

		// Set value type
		switch value.(type) {
		case int, int8, int16, int32, int64:
			item.vt = common.INT
		case uint, uint8, uint16, uint32, uint64:
			item.vt = common.UINT
		case float32, float64:
			item.vt = common.FLOAT
		case string:
			item.vt = common.STRING
		default:
			// For other types, serialize as JSON string
			jsonStr, err := json.Marshal(item.value)
			if err != nil {
				return err
			}
			item.vt = common.OBJ
			item.value = jsonStr
		}

		c.items[key] = item

		// Record to AOF
		if c.aof != nil {
			if err := c.aof.Store(ctx, common.SET, key, fmt.Sprint(item.value),
				string(item.vt), fmt.Sprint(expireTs)); err != nil {
				return fmt.Errorf("failed to store in AOF: %w", err)
			}
		}

		return nil
	}
}

// Incr atomically increments an integer value
// ctx    context for cancellation operations
// key    entry key
//
// Returns:
// - int   new value after operation
// - error possible errors:
//   - ErrNotFound key does not exist
//   - ErrTypeMismatch value type is not integer
//   - ErrOutOfRange value overflow
//
// Note:
// - If the key does not exist, it returns ErrNotFound
// - The operation keeps the original TTL time unchanged
func (c *Cache) Incr(ctx context.Context, key string) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// Get existing value, default to 0 if it doesn't exist
		val, err := c.GetInt(ctx, key)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// If the key does not exist, create a new one
				if err = c.Set(ctx, key, 1); err != nil {
					return 0, err
				}
			} else {
				return 0, err
			}
		}

		// Overflow check
		if val == math.MaxInt {
			return 0, ErrOutOfRange
		}

		c.mu.Lock()
		c.items[key] = cacheItem{
			value:    val + 1,
			vt:       common.INT,
			expireAt: c.items[key].expireAt, // If key exists, keep original TTL, otherwise 0 (never expire)
		}
		c.mu.Unlock()

		return val + 1, nil
	}
}

// IncrUint atomically increments an unsigned integer value
// ctx    context for cancellation operations
// key    entry key
//
// Returns:
// - uint   new value after operation
// - error possible errors:
//   - ErrNotFound key does not exist
//   - ErrTypeMismatch value type is not unsigned integer
//   - ErrOutOfRange value overflow
//
// Note:
// - If the key does not exist, it returns ErrNotFound
// - The operation keeps the original TTL time unchanged
func (c *Cache) IncrUint(ctx context.Context, key string) (uint, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		// Get existing value, default to 0 if it doesn't exist
		item, exists := c.items[key]
		var val uint = 0
		if exists {
			// Type check
			v, err := c.GetUint(ctx, key)
			if err != nil {
				return 0, err
			}
			val = v
		}

		// Overflow check
		if val == math.MaxUint {
			return 0, ErrOutOfRange
		}

		newVal := val + 1
		c.items[key] = cacheItem{
			value:    newVal,
			vt:       common.UINT,
			expireAt: item.expireAt, // If key exists, keep original TTL, otherwise 0 (never expire)
		}

		return newVal, nil
	}
}

// Get retrieves the original value of a Cache entry
// ctx  context for cancellation operations
// key  entry key to retrieve
//
// Returns:
// - any    original stored value
// - error  encountered errors during the operation
func (c *Cache) Get(ctx context.Context, key string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Try to minimize lock range
		c.mu.RLock()
		item, exists := c.items[key]
		c.mu.RUnlock()

		if !exists {
			return nil, NewNotFoundError("键不存在：" + key)
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

// GetInt retrieves an int value (automatically handles type conversion)
func (c *Cache) GetInt(ctx context.Context, key string) (int, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int: // Direct return for native int type
		return v, nil
	case int8, int16, int32: // Safe conversion for small integers
		return int(v.(int32)), nil // Type assertion and conversion
	case int64: // 64-bit integers need range check
		if v > math.MaxInt || v < math.MinInt {
			return 0, NewOutOfRangeError("值超出int范围")
		}
		return int(v), nil
	case uint, uint8, uint16, uint32, uint64:
		u := reflect.ValueOf(v).Uint()
		if u > math.MaxInt {
			return 0, NewOutOfRangeError("值超出int范围")
		}
		return int(u), nil
	default:
		return 0, NewTypeMismatchError("无法将类型转换为int")
	}
}

// GetUint retrieves an unsigned integer value (automatically handles type conversion)
// ctx  context for cancellation operations
// key  entry key to retrieve
//
// Returns:
// - uint  converted unsigned integer value
// - error conversion error (ErrTypeMismatch/ErrOutOfRange) or operation error
//
// Supported type conversions:
// - All unsigned integers (uint8/16/32/64)
// - Signed integers (int8/16/32/64) must be non-negative
// - Floating point (float32/64) must be in the [0, math.MaxUint64] range
func (c *Cache) GetUint(ctx context.Context, key string) (uint, error) {
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

// GetFloat retrieves a floating point value (automatically converts to float64)
// ctx  context for cancellation operations
// key  entry key to retrieve
//
// Returns:
// - float64  converted floating point value
// - error    conversion error (ErrTypeMismatch) or operation error
//
// Supported type conversions:
// - All integers (int/uint series) and floating point
// - Other types return ErrTypeMismatch
func (c *Cache) GetFloat(ctx context.Context, key string) (float64, error) {
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

// GetBool retrieves a boolean value
// ctx  context for cancellation operations
// key  entry key to retrieve
//
// Returns:
// - bool   converted boolean value
// - error  conversion error (ErrTypeMismatch) or operation error
//
// Note:
// - Only supports native bool type, does not support string/number to bool conversion
func (c *Cache) GetBool(ctx context.Context, key string) (bool, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return false, err
	}

	if b, ok := val.(bool); ok {
		return b, nil
	}
	return false, ErrTypeMismatch
}

// GetString retrieves a string value
// ctx  context for cancellation operations
// key  entry key to retrieve
//
// Returns:
// - string  converted string value
// - error   conversion error (ErrTypeMismatch) or operation error
//
// Note:
// - Only supports native string type, does not support automatic type conversion
func (c *Cache) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", ErrTypeMismatch
}

// Delete removes an entry from the Cache regardless of its expiration status
// It returns no error if the key doesn't exist
func (c *Cache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.items, key)

		// Record to AOF
		if c.aof != nil {
			if err := c.aof.Store(ctx, common.DELETE, key); err != nil {
				return fmt.Errorf("failed to store in AOF: %w", err)
			}
		}

		return nil
	}
}

// Cleanup removes all expired entries from the Cache
// This operation blocks all read/write operations while running
// It's recommended to run this during low-traffic periods
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.expireAt.IsZero() && now.After(item.expireAt) {
			delete(c.items, key)
		}
	}

	// Record to AOF
	if c.aof != nil {
		// Use background context because this is internal call
		if err := c.aof.Store(context.Background(), common.CLEANUP); err != nil {
			// Only log, does not affect cleanup operation
			logger.Error("failed to store cleanup in AOF: %v", err)
		}
	}
}

// CleanAll removes all entries from the Cache regardless of expiration status
func (c *Cache) CleanAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		delete(c.items, key) // map delete operation
	}
}
