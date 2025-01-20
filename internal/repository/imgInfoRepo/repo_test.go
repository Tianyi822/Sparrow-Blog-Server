package imgInfoRepo

import (
	"context"
	"h2blog/internal/model/po"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/pkg/utils"
	"h2blog/storage"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	// 加载配置文件
	config.LoadConfig("../../../resources/config/test/repository-config.yaml")
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
		imgInfo po.ImgInfo
		wantErr bool
	}{
		{
			name: "正常添加",
			imgInfo: po.ImgInfo{
				ImgName:    "test.jpg",
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "重复数据",
			imgInfo: po.ImgInfo{
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
			num, err := AddImgInfo(ctx, &tt.imgInfo)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), num)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), num)

				// 验证数据是否正确保存
				var saved po.ImgInfo
				err = tx.Where("img_name = ?", tt.imgInfo.ImgName).First(&saved).Error
				assert.NoError(t, err)
			}
		})
	}
}

func TestImgInfo_FindOneById(t *testing.T) {
	ctx := context.Background()

	// 准备测试数据
	testImg := po.ImgInfo{
		ImgId:      "test123456789012",
		ImgName:    "test_find.jpg",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	_, err := AddImgInfo(ctx, &testImg)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		name    string
		imgId   string
		wantErr bool
	}{
		{
			name:    "查询存在的记录",
			imgId:   testImg.ImgId,
			wantErr: false,
		},
		{
			name:    "查询不存在的记录",
			imgId:   "nonexistent123456",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ii := &po.ImgInfo{
				ImgId: tt.imgId,
			}

			got, err := FindImgById(ctx, tt.imgId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, ii.ImgId, got.ImgId)
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
	testImages := []po.ImgInfo{
		{ImgId: "test111", ImgName: "test_image_1.jpg"},
		{ImgId: "test222", ImgName: "test_image_2.jpg"},
		{ImgId: "test333", ImgName: "other_image.jpg"},
	}

	for _, img := range testImages {
		_, err := AddImgInfo(ctx, &img)
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
			got, err := FindImgByNameLike(ctx, tt.keyword)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLen, len(got))
			}
		})
	}
}

func TestImgInfo_AddBatch(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		images   []po.ImgInfo
		wantRows int64
		wantErr  bool
	}{
		{
			name: "正常批量添加",
			images: []po.ImgInfo{
				{
					ImgId:      "batch_test_1",
					ImgName:    "batch1.jpg",
					CreateTime: time.Now(),
					UpdateTime: time.Now(),
				},
				{
					ImgId:      "batch_test_2",
					ImgName:    "batch2.jpg",
					CreateTime: time.Now(),
					UpdateTime: time.Now(),
				},
			},
			wantRows: 2,
			wantErr:  false,
		},
		{
			name:     "空切片添加",
			images:   []po.ImgInfo{},
			wantRows: 0,
			wantErr:  false,
		},
		{
			name: "重复数据添加",
			images: []po.ImgInfo{
				{
					ImgId:      "batch_test_1",
					ImgName:    "batch1.jpg",
					CreateTime: time.Now(),
					UpdateTime: time.Now(),
				},
			},
			wantRows: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 开启事务
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			// 执行测试
			rows, err := AddImgInfoBatch(ctx, tt.images)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantRows, rows)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, rows)

				// 验证数据是否正确保存
				if len(tt.images) > 0 {
					var saved []po.ImgInfo
					err = tx.Where("img_id IN ?", []string{tt.images[0].ImgId}).Find(&saved).Error
					assert.NoError(t, err)
					assert.Equal(t, len(saved), 1)
					assert.Equal(t, tt.images[0].ImgName, saved[0].ImgName)
				}
			}

			// 删除测试数据
			//for _, img := range tt.images {
			//	tx.Where("img_id = ?", img.ImgId).Delete(&po.ImgInfo{})
			//}
			//tx.Commit()
		})
	}
}

func TestImgInfo_DeleteBatch(t *testing.T) {
	ctx := context.Background()

	// Prepare test data
	testImages := []po.ImgInfo{
		{
			ImgId:      "delete_test_1",
			ImgName:    "delete1.jpg",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
		{
			ImgId:      "delete_test_2",
			ImgName:    "delete2.jpg",
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	}

	_, err := AddImgInfoBatch(ctx, testImages)
	if err != nil {
		t.Fatalf("Failed to prepare test data: %v", err)
	}

	tests := []struct {
		name     string
		ids      []string
		wantRows int64
		wantErr  bool
	}{
		{
			name:     "正常批量删除",
			ids:      []string{"delete_test_1", "delete_test_2"},
			wantRows: 2,
			wantErr:  false,
		},
		{
			name:     "空切片删除",
			ids:      []string{},
			wantRows: 0,
			wantErr:  false,
		},
		{
			name:     "删除不存在的记录",
			ids:      []string{"nonexistent_id"},
			wantRows: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start transaction
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			// Execute test
			rows, err := DeleteImgInfoBatch(ctx, tt.ids)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, rows)

				// Verify records were deleted
				if len(tt.ids) > 0 {
					var count int64
					tx.Model(&po.ImgInfo{}).Where("img_id IN ?", tt.ids).Count(&count)
					assert.Equal(t, int64(0), count)
				}
			}
		})
	}
}

func TestImgInfo_UpdateNameById(t *testing.T) {
	ctx := context.Background()

	// Prepare test data
	testImg := po.ImgInfo{
		ImgId:      "update_test_1",
		ImgName:    "original.jpg",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	_, err := AddImgInfo(ctx, &testImg)
	if err != nil {
		t.Fatalf("Failed to prepare test data: %v", err)
	}

	tests := []struct {
		name     string
		id       string
		newName  string
		wantRows int64
		wantErr  bool
	}{
		{
			name:     "正常更新名称",
			id:       "update_test_1",
			newName:  "updated.jpg",
			wantRows: 1,
			wantErr:  false,
		},
		{
			name:     "更新不存在的记录",
			id:       "nonexistent_id",
			newName:  "new.jpg",
			wantRows: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start transaction
			tx := storage.Storage.Db.Begin()
			defer tx.Rollback()

			// Execute test
			rows, err := UpdateImgNameById(ctx, tt.id, tt.newName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRows, rows)

				// Verify name was updated
				if tt.wantRows > 0 {
					var updated po.ImgInfo
					tx.Where("img_id = ?", tt.id).First(&updated)
					assert.Equal(t, tt.newName, updated.ImgName)
				}
			}
		})
	}
}
