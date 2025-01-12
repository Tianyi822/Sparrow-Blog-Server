package entity

import (
	"context"
	"github.com/stretchr/testify/assert"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/pkg/utils"
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

func TestImgInfo_AddOne(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		imgInfo ImgInfo
		wantErr bool
	}{
		{
			name: "正常添加",
			imgInfo: ImgInfo{
				ImgName:    "test.jpg",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "重复数据",
			imgInfo: ImgInfo{
				ImgName:    "test.jpg",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 开启事务
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			id, err := utils.HashWithLength(tt.imgInfo.ImgName, 16)
			if err != nil {
				t.Error(err)
			}
			tt.imgInfo.ImgId = id

			// 执行测试
			num, err := tt.imgInfo.AddOne(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), num)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), num)

				// 验证数据是否正确保存
				var saved ImgInfo
				err = tx.Where("img_name = ?", tt.imgInfo.ImgName).First(&saved).Error
				assert.NoError(t, err)
			}
		})
	}
}

func TestImgInfo_FindOneById(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据
	testImg := ImgInfo{
		ImgId:      "test123456789012",
		ImgName:    "test_find.jpg",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	_, err := testImg.AddOne(ctx)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name    string
		imgId   string
		want    int64
		wantErr bool
	}{
		{
			name:    "查询存在的记录",
			imgId:   testImg.ImgId,
			want:    1,
			wantErr: false,
		},
		{
			name:    "查询不存在的记录",
			imgId:   "nonexistent123456",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ii := &ImgInfo{
				ImgId: tt.imgId,
			}

			got, err := ii.FindOneById(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)

				if tt.want > 0 {
					assert.Equal(t, testImg.ImgName, ii.ImgName)
				}
			}
		})
	}
}

func TestImgInfo_FindByNameLike(t *testing.T) {
	ctx := context.Background()

	// 开启事务
	tx := storage.Storage.Db.Begin()
	defer tx.Rollback()

	// 准备测试数据
	testImages := []ImgInfo{
		{ImgId: "test111", ImgName: "test_image_1.jpg"},
		{ImgId: "test222", ImgName: "test_image_2.jpg"},
		{ImgId: "test333", ImgName: "other_image.jpg"},
	}

	for _, img := range testImages {
		_, err := img.AddOne(ctx)
		if err != nil {
			t.Error(err)
		}
	}

	tests := []struct {
		name    string
		keyword string
		wantLen int
		wantErr bool
	}{
		{
			name:    "查找test关键字",
			keyword: "test",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "查找other关键字",
			keyword: "other",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "查找不存在的关键字",
			keyword: "nonexist",
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ii := &ImgInfo{ImgName: tt.keyword}
			got, err := ii.FindByNameLike(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(got))
			}
		})
	}
}
