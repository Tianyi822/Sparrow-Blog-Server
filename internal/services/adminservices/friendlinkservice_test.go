package adminservices

import (
	"context"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage"
	"testing"
)

func init() {
	// 加载配置文件
	config.LoadConfig()
	// 初始化 Logger 组件
	err := logger.InitLogger(context.Background())
	if err != nil {
		return
	}
	// 初始化数据库组件
	_ = storage.InitStorage(context.Background())
}

// 测试用的友链ID，用于更新和删除测试
var testFriendLinkId string = "f4e5dd52f2c572df"

// TestCreateFriendLink 测试创建友链
func TestCreateFriendLink(t *testing.T) {
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkName:  "测试友链",
		FriendLinkUrl:   "https://test.example.com",
		FriendAvatarUrl: "https://test.example.com/avatar.jpg",
		FriendDescribe:  "这是一个测试友链",
	}

	err := CreateFriendLink(context.Background(), friendLinkDto)
	if err != nil {
		t.Errorf("创建友链失败: %v", err)
		return
	}

	// 保存生成的ID，用于后续测试
	testFriendLinkId = friendLinkDto.FriendLinkId
	t.Logf("创建友链成功，ID: %s", testFriendLinkId)
}

// TestGetAllFriendLinks 测试获取所有友链
func TestGetAllFriendLinks(t *testing.T) {
	friendLinks, err := GetAllFriendLinks(context.Background())
	if err != nil {
		t.Errorf("获取所有友链失败: %v", err)
		return
	}

	t.Logf("获取友链成功，共 %d 个友链", len(friendLinks))
	for i, friendLink := range friendLinks {
		t.Logf("友链 %d: ID=%s, Name=%s, URL=%s, Avatar=%s, Describe=%s",
			i+1, friendLink.FriendLinkId, friendLink.FriendLinkName, friendLink.FriendLinkUrl,
			friendLink.FriendAvatarUrl, friendLink.FriendDescribe)
	}
}

// TestUpdateFriendLink 测试更新友链
func TestUpdateFriendLink(t *testing.T) {
	if testFriendLinkId == "" {
		t.Skip("跳过更新测试，因为没有可用的测试友链ID")
		return
	}

	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkId:    testFriendLinkId,
		FriendLinkName:  "更新后的测试友链",
		FriendLinkUrl:   "https://updated.example.com",
		FriendAvatarUrl: "https://updated.example.com/avatar.jpg",
		FriendDescribe:  "这是一个更新后的测试友链",
	}

	err := UpdateFriendLink(context.Background(), friendLinkDto)
	if err != nil {
		t.Errorf("更新友链失败: %v", err)
		return
	}

	t.Logf("更新友链成功，ID: %s", testFriendLinkId)
}

// TestDeleteFriendLinkById 测试删除友链
func TestDeleteFriendLinkById(t *testing.T) {
	if testFriendLinkId == "" {
		t.Skip("跳过删除测试，因为没有可用的测试友链ID")
		return
	}

	err := DeleteFriendLinkById(context.Background(), testFriendLinkId)
	if err != nil {
		t.Errorf("删除友链失败: %v", err)
		return
	}

	t.Logf("删除友链成功，ID: %s", testFriendLinkId)
}

// TestFriendLinkServiceIntegration 集成测试：测试完整的CRUD流程
func TestFriendLinkServiceIntegration(t *testing.T) {
	ctx := context.Background()

	// 1. 创建友链
	t.Log("=== 开始集成测试：创建友链 ===")
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkName:  "集成测试友链",
		FriendLinkUrl:   "https://integration.test.com",
		FriendAvatarUrl: "https://integration.test.com/avatar.jpg",
		FriendDescribe:  "这是一个集成测试友链",
	}

	err := CreateFriendLink(ctx, friendLinkDto)
	if err != nil {
		t.Fatalf("集成测试-创建友链失败: %v", err)
	}
	createdId := friendLinkDto.FriendLinkId
	t.Logf("集成测试-创建友链成功，ID: %s", createdId)

	// 2. 获取所有友链，验证刚创建的友链是否存在
	t.Log("=== 验证友链创建成功 ===")
	allFriendLinks, err := GetAllFriendLinks(ctx)
	if err != nil {
		t.Fatalf("集成测试-获取所有友链失败: %v", err)
	}

	found := false
	for _, fl := range allFriendLinks {
		if fl.FriendLinkId == createdId {
			found = true
			if fl.FriendLinkName != "集成测试友链" {
				t.Errorf("集成测试-友链名称不匹配，期望: %s, 实际: %s", "集成测试友链", fl.FriendLinkName)
			}
			if fl.FriendLinkUrl != "https://integration.test.com" {
				t.Errorf("集成测试-友链URL不匹配，期望: %s, 实际: %s", "https://integration.test.com", fl.FriendLinkUrl)
			}
			if fl.FriendAvatarUrl != "https://integration.test.com/avatar.jpg" {
				t.Errorf("集成测试-友链头像URL不匹配，期望: %s, 实际: %s", "https://integration.test.com/avatar.jpg", fl.FriendAvatarUrl)
			}
			if fl.FriendDescribe != "这是一个集成测试友链" {
				t.Errorf("集成测试-友链描述不匹配，期望: %s, 实际: %s", "这是一个集成测试友链", fl.FriendDescribe)
			}
			break
		}
	}
	if !found {
		t.Fatalf("集成测试-创建的友链未找到，ID: %s", createdId)
	}
	t.Log("集成测试-友链创建验证成功")

	// 3. 更新友链
	t.Log("=== 开始更新友链 ===")
	updateDto := &dto.FriendLinkDto{
		FriendLinkId:    createdId,
		FriendLinkName:  "更新后的集成测试友链",
		FriendLinkUrl:   "https://updated.integration.test.com",
		FriendAvatarUrl: "https://updated.integration.test.com/avatar.jpg",
		FriendDescribe:  "这是一个更新后的集成测试友链",
	}

	err = UpdateFriendLink(ctx, updateDto)
	if err != nil {
		t.Fatalf("集成测试-更新友链失败: %v", err)
	}
	t.Log("集成测试-友链更新成功")

	// 4. 验证更新结果
	t.Log("=== 验证友链更新成功 ===")
	allFriendLinks, err = GetAllFriendLinks(ctx)
	if err != nil {
		t.Fatalf("集成测试-获取所有友链失败: %v", err)
	}

	found = false
	for _, fl := range allFriendLinks {
		if fl.FriendLinkId == createdId {
			found = true
			if fl.FriendLinkName != "更新后的集成测试友链" {
				t.Errorf("集成测试-更新后友链名称不匹配，期望: %s, 实际: %s", "更新后的集成测试友链", fl.FriendLinkName)
			}
			if fl.FriendLinkUrl != "https://updated.integration.test.com" {
				t.Errorf("集成测试-更新后友链URL不匹配，期望: %s, 实际: %s", "https://updated.integration.test.com", fl.FriendLinkUrl)
			}
			if fl.FriendAvatarUrl != "https://updated.integration.test.com/avatar.jpg" {
				t.Errorf("集成测试-更新后友链头像URL不匹配，期望: %s, 实际: %s", "https://updated.integration.test.com/avatar.jpg", fl.FriendAvatarUrl)
			}
			if fl.FriendDescribe != "这是一个更新后的集成测试友链" {
				t.Errorf("集成测试-更新后友链描述不匹配，期望: %s, 实际: %s", "这是一个更新后的集成测试友链", fl.FriendDescribe)
			}
			break
		}
	}
	if !found {
		t.Fatalf("集成测试-更新的友链未找到，ID: %s", createdId)
	}
	t.Log("集成测试-友链更新验证成功")

	// 5. 删除友链
	t.Log("=== 开始删除友链 ===")
	err = DeleteFriendLinkById(ctx, createdId)
	if err != nil {
		t.Fatalf("集成测试-删除友链失败: %v", err)
	}
	t.Log("集成测试-友链删除成功")

	// 6. 验证删除结果
	t.Log("=== 验证友链删除成功 ===")
	allFriendLinks, err = GetAllFriendLinks(ctx)
	if err != nil {
		t.Fatalf("集成测试-获取所有友链失败: %v", err)
	}

	for _, fl := range allFriendLinks {
		if fl.FriendLinkId == createdId {
			t.Fatalf("集成测试-友链删除失败，友链仍然存在，ID: %s", createdId)
		}
	}
	t.Log("集成测试-友链删除验证成功")
	t.Log("=== 集成测试完成 ===")
}

// TestCreateFriendLinkWithEmptyData 测试创建友链时的边界情况
func TestCreateFriendLinkWithEmptyData(t *testing.T) {
	// 测试空名称
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkName: "",
		FriendLinkUrl:  "https://test.example.com",
	}

	err := CreateFriendLink(context.Background(), friendLinkDto)
	if err == nil {
		t.Log("注意：创建空名称友链未返回错误，可能需要在业务层添加验证")
		// 清理创建的数据
		if friendLinkDto.FriendLinkId != "" {
			_ = DeleteFriendLinkById(context.Background(), friendLinkDto.FriendLinkId)
		}
	} else {
		t.Logf("创建空名称友链正确返回错误: %v", err)
	}

	// 测试空URL
	friendLinkDto2 := &dto.FriendLinkDto{
		FriendLinkName: "测试友链",
		FriendLinkUrl:  "",
	}

	err = CreateFriendLink(context.Background(), friendLinkDto2)
	if err == nil {
		t.Log("注意：创建空URL友链未返回错误，可能需要在业务层添加验证")
		// 清理创建的数据
		if friendLinkDto2.FriendLinkId != "" {
			_ = DeleteFriendLinkById(context.Background(), friendLinkDto2.FriendLinkId)
		}
	} else {
		t.Logf("创建空URL友链正确返回错误: %v", err)
	}
}

// TestDeleteNonExistentFriendLink 测试删除不存在的友链
func TestDeleteNonExistentFriendLink(t *testing.T) {
	nonExistentId := "non_existent_friend_link_id"
	err := DeleteFriendLinkById(context.Background(), nonExistentId)
	if err != nil {
		t.Logf("删除不存在的友链正确返回错误: %v", err)
	} else {
		t.Log("删除不存在的友链未返回错误，这可能是正常的（取决于业务逻辑）")
	}
}

// TestUpdateNonExistentFriendLink 测试更新不存在的友链
func TestUpdateNonExistentFriendLink(t *testing.T) {
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkId:   "non_existent_friend_link_id",
		FriendLinkName: "不存在的友链",
		FriendLinkUrl:  "https://nonexistent.example.com",
	}

	err := UpdateFriendLink(context.Background(), friendLinkDto)
	if err != nil {
		t.Logf("更新不存在的友链正确返回错误: %v", err)
	} else {
		t.Log("更新不存在的友链未返回错误，这可能是正常的（取决于业务逻辑）")
	}
}
