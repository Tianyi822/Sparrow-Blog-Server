package webjwt

import (
	"context"
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

func TestGenerateJWTToken(t *testing.T) {
	token, err := GenerateJWTToken()
	if err != nil {
		t.Errorf("生成 Token 失败: %v", err.Error())
	} else {
		t.Logf("生成 Token 成功: %v", token)
	}

	customClaims, err := ParseJWTToken(token)
	if err != nil {
		t.Errorf("解析 Token 失败: %v", err.Error())
	} else {
		t.Logf("解析 Token 成功: %v", customClaims.UserEmail)
	}
}
