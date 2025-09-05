package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sparrow_blog_server/cache/aof"
	"sparrow_blog_server/cache/common"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
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
						return err
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

	// 加载完数据后，需要将当前内存中的数据持久化到磁盘，保证缓存启动时，磁盘与内存中的数据一致
	for k, v := range c.items {
		if err := c.aof.Store(
			ctx,
			common.SET,
			k,
			fmt.Sprint(v.value),
			fmt.Sprint(v.vt),
			fmt.Sprint(v.expireAt.Unix()),
		); err != nil {
			return fmt.Errorf("failed to store in AOF: %w", err)
		}
	}

	return nil
}

// Set 在缓存中存储一个值，不设置过期时间
func (c *Cache) Set(ctx context.Context, key string, value any) error {
	return c.SetWithExpired(ctx, key, value, 0)
}

// SetWithExpired 在缓存中存储一个带有可选TTL的值
// 如果键已存在，将被覆盖，TTL将被重置
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

		// 类型安全检查：不允许存储指针、数组或切片类型
		if reflect.TypeOf(value).Kind() == reflect.Ptr ||
			reflect.TypeOf(value).Kind() == reflect.Array ||
			reflect.TypeOf(value).Kind() == reflect.Slice {
			return ErrPointerNotAllowed
		}

		item := cacheItem{
			value: value,
		}

		// 设置过期时间
		var expireTs int64
		if ttl > 0 {
			item.expireAt = time.Now().Add(ttl)
			expireTs = item.expireAt.Unix()
		}

		// 设置值类型
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
			// 对于其他类型，序列化为JSON字符串
			jsonStr, err := json.Marshal(item.value)
			if err != nil {
				return err
			}
			item.vt = common.OBJ
			item.value = jsonStr
		}

		c.items[key] = item

		// 记录到AOF
		if c.aof != nil {
			if err := c.aof.Store(
				ctx,
				common.SET,
				key,
				fmt.Sprint(item.value),
				fmt.Sprint(item.vt),
				fmt.Sprint(expireTs),
			); err != nil {
				return fmt.Errorf("failed to store in AOF: %w", err)
			}
		}

		return nil
	}
}

// Incr 原子递增一个整数值
// ctx    用于取消操作的上下文
// key    条目键
//
// 返回:
// - int   操作后的新值
// - error 可能的错误:
//   - ErrNotFound 键不存在
//   - ErrTypeMismatch 值类型不是整数
//   - ErrOutOfRange 值溢出
//
// 注意:
// - 如果键不存在，它将返回ErrNotFound
// - 操作会保持原始TTL时间不变
func (c *Cache) Incr(ctx context.Context, key string) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		// 获取现有值，如果不存在则默认为0
		val, err := c.GetInt(ctx, key)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				// 如果键不存在，创建一个新的
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
			vt:       common.INT,
			expireAt: c.items[key].expireAt, // 如果键存在，保持原始TTL，否则为0（永不过期）
		}
		c.mu.Unlock()

		return val + 1, nil
	}
}

// IncrUint 原子递增一个无符号整数值
// ctx    用于取消操作的上下文
// key    条目键
//
// 返回:
// - uint  操作后的新值
// - error 可能的错误:
//   - ErrNotFound 键不存在
//   - ErrTypeMismatch 值类型不是无符号整数
//   - ErrOutOfRange 值溢出
//
// 注意:
// - 如果键不存在，它将返回ErrNotFound
// - 操作会保持原始TTL时间不变
func (c *Cache) IncrUint(ctx context.Context, key string) (uint, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		// 获取现有值，如果不存在则默认为0
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
			vt:       common.UINT,
			expireAt: item.expireAt, // 如果键存在，保持原始TTL，否则为0（永不过期）
		}

		return newVal, nil
	}
}

// Get 检索缓存条目的原始值
// ctx  用于取消操作的上下文
// key  要检索的条目键
//
// 返回:
// - any    原始存储的值
// - error  操作过程中遇到的错误
func (c *Cache) Get(ctx context.Context, key string) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// 尝试最小化锁定范围
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

// GetInt 检索一个整数值（自动处理类型转换）
func (c *Cache) GetInt(ctx context.Context, key string) (int, error) {
	val, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case int: // 直接返回原生int类型
		return v, nil
	case int8, int16, int32: // 小整数的安全转换
		return int(v.(int32)), nil // 类型断言和转换
	case int64: // 64位整数需要范围检查
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

// GetUint 检索一个无符号整数值（自动处理类型转换）
// ctx  用于取消操作的上下文
// key  要检索的条目键
//
// 返回:
// - uint  转换后的无符号整数值
// - error 转换错误（ErrTypeMismatch/ErrOutOfRange）或操作错误
//
// 支持的类型转换:
// - 所有无符号整数（uint8/16/32/64）
// - 有符号整数（int8/16/32/64）必须是非负数
// - 浮点数（float32/64）必须在[0, math.MaxUint64]范围内
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

// GetFloat 检索一个浮点值（自动转换为float64）
// ctx  用于取消操作的上下文
// key  要检索的条目键
//
// 返回:
// - float64  转换后的浮点值
// - error    转换错误（ErrTypeMismatch）或操作错误
//
// 支持的类型转换:
// - 所有整数（int/uint系列）和浮点数
// - 其他类型返回ErrTypeMismatch
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

// GetBool 检索一个布尔值
// ctx  用于取消操作的上下文
// key  要检索的条目键
//
// 返回:
// - bool   转换后的布尔值
// - error  转换错误（ErrTypeMismatch）或操作错误
//
// 注意:
// - 仅支持原生bool类型，不支持字符串/数字到布尔值的转换
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

// GetString 检索一个字符串值
// ctx  用于取消操作的上下文
// key  要检索的条目键
//
// 返回:
// - string  转换后的字符串值
// - error   转换错误（ErrTypeMismatch）或操作错误
//
// 注意:
// - 仅支持原生string类型，不支持自动类型转换
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

// GetKeysLike 获取所有键名中包含指定字符串的键
// 参数:
// - ctx: 用于取消操作的上下文
// - matchStr: 要匹配的字符串，支持通配符*
//
// 返回:
// - []string: 所有匹配的键名
// - error: 操作错误（如上下文取消）
func (c *Cache) GetKeysLike(ctx context.Context, matchStr string) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		keys := make([]string, 0)
		for key := range c.items {
			if strings.Contains(key, matchStr) {
				keys = append(keys, key)
			}
		}
		return keys, nil
	}
}

// Delete 从缓存中删除一个条目，无论其过期状态如何
// 如果键不存在，则不返回错误
func (c *Cache) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.items, key)

		// 记录到AOF
		if c.aof != nil {
			if err := c.aof.Store(ctx, common.DELETE, key); err != nil {
				return fmt.Errorf("failed to store in AOF: %w", err)
			}
		}

		return nil
	}
}

// Cleanup 从缓存中删除所有过期的条目
// 此操作在运行时会阻塞所有读/写操作
// 建议在低流量期间运行
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if !item.expireAt.IsZero() && now.After(item.expireAt) {
			delete(c.items, key)
		}
	}

	// 记录到AOF
	if c.aof != nil {
		// 使用后台上下文，因为这是内部调用
		if err := c.aof.Store(context.Background(), common.CLEANUP); err != nil {
			// 仅记录日志，不影响清理操作
			logger.Error("failed to store cleanup in AOF: %v", err)
		}
	}
}

// CleanAll 从缓存中删除所有条目，无论其过期状态如何
func (c *Cache) CleanAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.items {
		delete(c.items, key) // map删除操作
	}
}

// Close 安全地关闭缓存并确保所有数据都被持久化到磁盘。
// 应该在应用程序关闭时调用此方法以防止数据丢失。
//
// 线程安全:
// - 使用互斥锁确保关闭过程中没有并发操作
//
// 返回:
// - error: 关闭操作过程中遇到的任何错误
func (c *Cache) Close() error {
	if c == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果启用了AOF，关闭它以将数据刷新到磁盘
	if c.aof != nil {
		if err := c.aof.Close(); err != nil {
			return fmt.Errorf("failed to close AOF: %w", err)
		}
		c.aof = nil
	}

	// 清理items映射
	c.items = nil

	return nil
}
