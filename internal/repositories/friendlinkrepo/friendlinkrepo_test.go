package friendlinkrepo

import (
	"context"
	"fmt"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"
)

func init() {
	_ = config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		panic(err)
	}
	// 初始化数据库组件
	err = storage.InitStorage(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestCreateFriendLink(t *testing.T) {
	ctx := context.Background()
	friendLinkDto := dto.FriendLinkDto{
		FriendLinkName: "chentyit",
		FriendLinkUrl:  "https://chentyit.github.io",
	}

	num, err := CreateFriendLink(ctx, friendLinkDto)
	if err != nil {
		t.Errorf("CreateFriendLink() error = %v", err)
	}

	fmt.Printf("添加 %v 条数据\n", num)
}

func TestGetFriendLinkByName(t *testing.T) {
	ctx := context.Background()
	friendLinkDto, err := GetFriendLinkByNameLike(ctx, "chentyit")
	if err != nil {
		t.Errorf("GetFriendLinkByName() error = %v", err)
	}

	fmt.Printf("friendLinkDto = %v\n", friendLinkDto)
}

func TestUpdateFriendLinkByID(t *testing.T) {
	ctx := context.Background()
	num, err := UpdateFriendLinkByID(ctx, dto.FriendLinkDto{
		FriendLinkId:   "eefe040262ec2915",
		FriendLinkName: "chentyit666",
		FriendLinkUrl:  "https://chentyit.github.io",
	})
	if err != nil {
		t.Errorf("UpdateFriendLinkByID() error = %v", err)
	}
	fmt.Printf("更新 %v", num)
}

func TestDeleteFriendLinkById(t *testing.T) {
	ctx := context.Background()
	num, err := DeleteFriendLinkById(ctx, "eefe040262ec2915")
	if err != nil {
		t.Errorf("DeleteFriendLinkById() error = %v", err)
	}
	fmt.Printf("删除 %v 条数据\n", num)
}
