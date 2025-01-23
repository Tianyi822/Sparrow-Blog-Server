package cache

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
	"time"
)

func TestDelete(t *testing.T) {
	c := NewCache()
	ctx := context.Background()

	t.Run("basic deletion", func(t *testing.T) {
		// Set and verify item exists
		err := c.Set(ctx, "key1", "value1", time.Hour)
		if err != nil {
			t.Fatalf("设置失败：%v", err)
		}

		// Delete the item
		err = c.Delete(ctx, "key1")
		if err != nil {
			t.Errorf("删除失败：%v", err)
		}

		// Verify item is gone
		_, err = c.Get(ctx, "key1")
		if err == nil {
			t.Error("删除后项目仍然存在")
		}
	})

	t.Run("delete non-existent", func(t *testing.T) {
		err := c.Delete(ctx, "nonexistent")
		if err != nil {
			t.Errorf("删除不存在的键不应报错：%v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := c.Delete(ctx, "key")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("期望 context.Canceled，实际得到 %v", err)
		}
	})
}

func TestConcurrentDeleteAndGet(t *testing.T) {
	c := NewCache()
	ctx := context.Background()
	keys := []string{"k1", "k2", "k3", "k4", "k5"}

	// Setup initial data
	for _, k := range keys {
		_ = c.Set(ctx, k, k, time.Hour)
	}

	var wg sync.WaitGroup
	iterations := 1000

	// Start readers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			for _, k := range keys {
				_, _ = c.Get(ctx, k)
			}
		}
	}()

	// Start deletes
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			for _, k := range keys {
				_ = c.Delete(ctx, k)
				_ = c.Set(ctx, k, k, time.Hour)
			}
		}
	}()

	wg.Wait()
}

func TestEdgeCases(t *testing.T) {
	c := NewCache()
	ctx := context.Background()

	t.Run("empty string key", func(t *testing.T) {
		err := c.Set(ctx, "", "value", time.Hour)
		if err != nil {
			t.Errorf("设置空字符串键失败：%v", err)
		}

		val, err := c.GetString(ctx, "")
		if err != nil || val != "value" {
			t.Errorf("获取空字符串键失败: err=%v val=%v", err, val)
		}
	})

	t.Run("zero TTL", func(t *testing.T) {
		err := c.Set(ctx, "zero", "value", 0)
		if err != nil {
			t.Errorf("设置零有效期失败：%v", err)
		}

		_, err = c.Get(ctx, "zero")
		if err != nil {
			t.Error("Zero TTL item should be never expired")
		}
	})

	t.Run("type conversion edge cases", func(t *testing.T) {
		testCases := []struct {
			name     string
			value    interface{}
			testFunc func() error
		}{
			{
				name:  "float64 to int overflow",
				value: math.MaxFloat64,
				testFunc: func() error {
					_ = c.Set(ctx, "test", math.MaxFloat64, time.Hour)
					_, err := c.GetInt(ctx, "test")
					return err
				},
			},
			{
				name:  "negative to uint",
				value: -1,
				testFunc: func() error {
					_ = c.Set(ctx, "test", -1, time.Hour)
					_, err := c.GetUint(ctx, "test")
					return err
				},
			},
			{
				name:  "string to bool",
				value: "true",
				testFunc: func() error {
					_ = c.Set(ctx, "test", "true", time.Hour)
					_, err := c.GetBool(ctx, "test")
					return err
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.testFunc()
				if err == nil {
					t.Error("Expected error, got nil")
				}
			})
		}
	})
}

func TestCleanupBehavior(t *testing.T) {
	c := NewCache()
	ctx := context.Background()

	// Setup test data
	_ = c.Set(ctx, "exp1", 1, -time.Hour)      // Already expired
	_ = c.Set(ctx, "exp2", 2, time.Nanosecond) // Will expire immediately
	_ = c.Set(ctx, "valid", 3, time.Hour)      // Valid

	time.Sleep(time.Millisecond) // Ensure exp2 expires

	c.Cleanup()

	// Verify cleanup behavior
	tests := []struct {
		key string
		err error
	}{
		{"exp1", ErrNotFound},
		{"exp2", ErrNotFound},
		{"valid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			_, err := c.Get(ctx, tt.key)
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
		})
	}
}
