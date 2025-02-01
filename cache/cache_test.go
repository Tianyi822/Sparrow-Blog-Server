package cache

import (
	"context"
	"errors"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"testing"
	"time"
)

func init() {
	config.LoadConfig("../resources/config/test/cache-config.yaml")
	_ = logger.InitLogger()
	InitCache(context.Background())
}

func TestCore_Basic(t *testing.T) {
	ctx := context.Background()

	// 测试 Set 和 Get
	t.Run("基本的 Set 和 Get 操作", func(t *testing.T) {
		err := Cache.Set(ctx, "test:string", "hello")
		if err != nil {
			t.Errorf("Set 错误: %v", err)
		}

		val, err := Cache.GetString(ctx, "test:string")
		if err != nil {
			t.Errorf("GetString 错误: %v", err)
		}
		if val != "hello" {
			t.Errorf("期望值为 'hello'，实际得到 '%s'", val)
		}
	})

	// 测试空键
	t.Run("空键处理", func(t *testing.T) {
		err := Cache.Set(ctx, "", "value")
		if !errors.Is(err, ErrEmptyKey) {
			t.Errorf("期望得到 ErrEmptyKey，实际得到 %v", err)
		}
	})

	// 测试指针值
	t.Run("指针值拒绝", func(t *testing.T) {
		x := "test"
		err := Cache.Set(ctx, "test:pointer", &x)
		if !errors.Is(err, ErrPointerNotAllowed) {
			t.Errorf("期望得到 ErrPointerNotAllowed，实际得到 %v", err)
		}
	})
}

func TestCore_TypedOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("整数操作", func(t *testing.T) {
		// 测试整数存储和获取
		err := Cache.Set(ctx, "test:int", 42)
		if err != nil {
			t.Errorf("设置整数错误: %v", err)
		}

		val, err := Cache.GetInt(ctx, "test:int")
		if err != nil {
			t.Errorf("获取整数错误: %v", err)
		}
		if val != 42 {
			t.Errorf("期望值为 42，实际得到 %d", val)
		}

		// 测试 Incr
		newVal, err := Cache.Incr(ctx, "test:int")
		if err != nil {
			t.Errorf("Incr 错误: %v", err)
		}
		if newVal != 43 {
			t.Errorf("Incr 后期望值为 43，实际得到 %d", newVal)
		}
	})

	t.Run("无符号整数操作", func(t *testing.T) {
		err := Cache.Set(ctx, "test:uint", uint(10))
		if err != nil {
			t.Errorf("设置无符号整数错误: %v", err)
		}

		val, err := Cache.GetUint(ctx, "test:uint")
		if err != nil {
			t.Errorf("获取无符号整数错误: %v", err)
		}
		if val != 10 {
			t.Errorf("期望值为 10，实际得到 %d", val)
		}
	})

	t.Run("浮点数操作", func(t *testing.T) {
		err := Cache.Set(ctx, "test:float", 3.14)
		if err != nil {
			t.Errorf("设置浮点数错误: %v", err)
		}

		val, err := Cache.GetFloat(ctx, "test:float")
		if err != nil {
			t.Errorf("获取浮点数错误: %v", err)
		}
		if val != 3.14 {
			t.Errorf("期望值为 3.14，实际得到 %f", val)
		}
	})
}

func TestCore_Expiration(t *testing.T) {
	ctx := context.Background()

	t.Run("TTL 过期", func(t *testing.T) {
		// 设置 100ms 的 TTL
		err := Cache.SetWithExpired(ctx, "test:expire", "temporary", 100*time.Millisecond)
		if err != nil {
			t.Errorf("SetWithExpired 错误: %v", err)
		}

		// 应该能立即获取
		_, err = Cache.Get(ctx, "test:expire")
		if err != nil {
			t.Errorf("过期前获取错误: %v", err)
		}

		// 等待过期
		time.Sleep(150 * time.Millisecond)

		// 过期后应返回未找到
		_, err = Cache.Get(ctx, "test:expire")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("过期后期望得到 ErrNotFound，实际得到 %v", err)
		}
	})
}

func TestCore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("删除操作", func(t *testing.T) {
		// 设置值
		err := Cache.Set(ctx, "test:delete", "value")
		if err != nil {
			t.Errorf("设置错误: %v", err)
		}

		// 删除它
		err = Cache.Delete(ctx, "test:delete")
		if err != nil {
			t.Errorf("删除错误: %v", err)
		}

		// 验证它已被删除
		_, err = Cache.Get(ctx, "test:delete")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("删除后期望得到 ErrNotFound，实际得到 %v", err)
		}
	})
}

func TestCore_Clean(t *testing.T) {
	ctx := context.Background()

	t.Run("清理过期条目", func(t *testing.T) {
		// 设置不同 TTL 的条目
		_ = Cache.SetWithExpired(ctx, "test:expire1", "val1", 50*time.Millisecond)
		_ = Cache.SetWithExpired(ctx, "test:expire2", "val2", 150*time.Millisecond)
		_ = Cache.Set(ctx, "test:persistent", "val3") // 无 TTL

		// 等待第一个条目过期
		time.Sleep(100 * time.Millisecond)

		// 运行清理
		Cache.Cleanup()

		// 检查结果
		_, err1 := Cache.Get(ctx, "test:expire1")
		_, err2 := Cache.Get(ctx, "test:expire2")
		_, err3 := Cache.Get(ctx, "test:persistent")

		if !errors.Is(err1, ErrNotFound) {
			t.Error("期望过期条目1已被清理")
		}
		if errors.Is(err2, ErrNotFound) {
			t.Error("期望未过期条目2仍然存在")
		}
		if err3 != nil {
			t.Error("期望永久条目仍然存在")
		}
	})

	t.Run("清理所有条目", func(t *testing.T) {
		// 设置一些条目
		_ = Cache.Set(ctx, "test:clean1", "val1")
		_ = Cache.Set(ctx, "test:clean2", "val2")
		_ = Cache.Set(ctx, "test:clean3", "val3")

		// 清理所有条目
		Cache.CleanAll()

		// 验证所有条目都被删除
		_, err := Cache.Get(ctx, "test:clean1")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("期望 test:clean1 被删除，实际得到 %v", err)
		}

		_, err = Cache.Get(ctx, "test:clean2")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("期望 test:clean2 被删除，实际得到 %v", err)
		}

		_, err = Cache.Get(ctx, "test:clean3")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("期望 test:clean3 被删除，实际得到 %v", err)
		}
	})
}
