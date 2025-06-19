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
	config.LoadConfig()
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
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkName: "chentyit",
		FriendLinkUrl:  "https://chentyit.github.io",
	}

	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := CreateFriendLink(tx, friendLinkDto)
	if err != nil {
		tx.Rollback()
		t.Errorf("CreateFriendLink() error = %v", err)
		return
	}

	// 提交事务
	tx.Commit()

	fmt.Printf("添加友链成功，ID: %s\n", friendLinkDto.FriendLinkId)
}

func TestUpdateFriendLinkByID(t *testing.T) {
	ctx := context.Background()
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkId:   "eefe040262ec2915",
		FriendLinkName: "chentyit666",
		FriendLinkUrl:  "https://chentyit.github.io",
	}

	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := UpdateFriendLinkByID(tx, friendLinkDto)
	if err != nil {
		tx.Rollback()
		t.Errorf("UpdateFriendLinkByID() error = %v", err)
		return
	}

	// 提交事务
	tx.Commit()

	fmt.Printf("更新友链成功\n")
}

func TestDeleteFriendLinkById(t *testing.T) {
	ctx := context.Background()

	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := DeleteFriendLinkById(tx, "eefe040262ec2915")
	if err != nil {
		tx.Rollback()
		t.Errorf("DeleteFriendLinkById() error = %v", err)
		return
	}

	// 提交事务
	tx.Commit()

	fmt.Printf("删除友链成功\n")
}

func TestFindAllFriendLinks(t *testing.T) {
	ctx := context.Background()
	friendLinks, err := FindAllFriendLinks(ctx)
	if err != nil {
		t.Errorf("FindAllFriendLinks() error = %v", err)
		return
	}

	fmt.Printf("查询到 %d 个友链:\n", len(friendLinks))
	for _, friendLink := range friendLinks {
		fmt.Printf("- ID: %s, Name: %s, URL: %s\n",
			friendLink.FriendLinkId,
			friendLink.FriendLinkName,
			friendLink.FriendLinkUrl)
	}
}
