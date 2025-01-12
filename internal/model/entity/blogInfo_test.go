package entity

import (
	"context"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/storage"
	"testing"
	"time"
)

func init() {
	// 加载配置文件
	config.LoadConfig("../../../resources/config/test/model-config.yaml")
	// 初始化 Logger 组件
	err := logger.InitLogger()
	if err != nil {
		return
	}
	// 初始化数据库组件
	storage.InitStorage()
}

func TestH2BlogInfo_AddOne(t *testing.T) {
	tests := []struct {
		name            string
		blogInfo        *BlogInfo
		wantAffectedNum int64
		wantErr         bool
	}{
		{
			name: "successful creation",
			blogInfo: &BlogInfo{
				Title:      "Test Blog",
				Brief:      "Test Brief",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantAffectedNum: 1,
			wantErr:         false,
		},
		{
			name: "duplicate title error",
			blogInfo: &BlogInfo{
				Title:      "Test Blog", // Duplicate title
				Brief:      "Another Brief",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantAffectedNum: 0,
			wantErr:         true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAffectedNum, err := tt.blogInfo.AddOne(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("BlogInfo.AddOne() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotAffectedNum != tt.wantAffectedNum {
				t.Errorf("BlogInfo.AddOne() = %v, want %v", gotAffectedNum, tt.wantAffectedNum)
			}
		})
	}

	// 关闭数据库连接
	storage.Storage.CloseDbConnect(ctx)
}

func TestH2BlogInfoPo_DeleteOne(t *testing.T) {
	tests := []struct {
		name            string
		blogInfo        *BlogInfo
		wantAffectedNum int64
		wantErr         bool
	}{
		{
			name: "successful deletion",
			blogInfo: &BlogInfo{
				BlogId: "test-blog-id",
			},
			wantAffectedNum: 1,
			wantErr:         false,
		},
		{
			name: "non-existent ID",
			blogInfo: &BlogInfo{
				BlogId: "non-existent-id",
			},
			wantAffectedNum: 0,
			wantErr:         false,
		},
	}

	ctx := context.Background()

	t1 := tests[0]
	_, _ = t1.blogInfo.AddOne(ctx)
	gotAffectedNum, err := t1.blogInfo.DeleteOneById(ctx)
	if (err != nil) != t1.wantErr {
		t.Errorf("BlogInfo.DeleteOneById() error = %v, wantErr %v", err, t1.wantErr)
		return
	}
	if gotAffectedNum != t1.wantAffectedNum {
		t.Errorf("BlogInfo.DeleteOneById() = %v, want %v", gotAffectedNum, t1.wantAffectedNum)
	}

	t2 := tests[1]
	gotAffectedNum, err = t2.blogInfo.DeleteOneById(ctx)
	if (err != nil) != t2.wantErr {
		t.Errorf("BlogInfo.DeleteOneById() error = %v, wantErr %v", err, t2.wantErr)
	}
	if gotAffectedNum != t2.wantAffectedNum {
		t.Errorf("BlogInfo.DeleteOneById() = %v, want %v", gotAffectedNum, t2.wantAffectedNum)
	}
}
