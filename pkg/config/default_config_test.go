package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateDefaultConfig(t *testing.T) {
	// 测试创建默认配置
	defaultConfig, err := createDefaultConfig()
	if err != nil {
		t.Fatalf("创建默认配置失败: %v", err)
	}

	// 验证日志配置
	if defaultConfig.Logger.Path == "" {
		t.Error("默认日志路径不应为空")
	}
	
	if !strings.Contains(defaultConfig.Logger.Path, "log/sparrow_blog.log") {
		t.Errorf("日志路径应包含 'log/sparrow_blog.log', 但得到: %s", defaultConfig.Logger.Path)
	}
	
	if defaultConfig.Logger.Level != "info" {
		t.Errorf("默认日志级别应为 'info', 但得到: %s", defaultConfig.Logger.Level)
	}
	
	if defaultConfig.Logger.MaxSize != 10 {
		t.Errorf("默认日志最大大小应为 10, 但得到: %d", defaultConfig.Logger.MaxSize)
	}

	// 验证服务器配置
	if defaultConfig.Server.Port != 8080 {
		t.Errorf("默认服务器端口应为 8080, 但得到: %d", defaultConfig.Server.Port)
	}

	// 验证搜索引擎配置
	if defaultConfig.SearchEngine.IndexPath == "" {
		t.Error("默认搜索引擎索引路径不应为空")
	}
	
	if !strings.Contains(defaultConfig.SearchEngine.IndexPath, "index/sparrow_blog.bleve") {
		t.Errorf("搜索引擎索引路径应包含 'index/sparrow_blog.bleve', 但得到: %s", defaultConfig.SearchEngine.IndexPath)
	}

	// 验证缓存配置
	if !defaultConfig.Cache.Aof.Enable {
		t.Error("默认应启用AOF")
	}
	
	if defaultConfig.Cache.Aof.Path == "" {
		t.Error("默认AOF路径不应为空")
	}
	
	if !strings.Contains(defaultConfig.Cache.Aof.Path, "aof/sparrow_blog.aof") {
		t.Errorf("AOF路径应包含 'aof/sparrow_blog.aof', 但得到: %s", defaultConfig.Cache.Aof.Path)
	}

	t.Logf("✅ 默认配置验证通过")
	t.Logf("日志路径: %s", defaultConfig.Logger.Path)
	t.Logf("搜索引擎索引路径: %s", defaultConfig.SearchEngine.IndexPath)
	t.Logf("AOF路径: %s", defaultConfig.Cache.Aof.Path)
}

func TestDefaultConfigFileGeneration(t *testing.T) {
	// 创建一个临时的项目目录用于测试
	tempDir := filepath.Join(os.TempDir(), "sparrow_blog_test")
	defer os.RemoveAll(tempDir)
	
	// 设置临时环境变量
	originalHome := os.Getenv("SPARROW_BLOG_HOME")
	os.Setenv("SPARROW_BLOG_HOME", tempDir)
	defer func() {
		if originalHome != "" {
			os.Setenv("SPARROW_BLOG_HOME", originalHome)
		} else {
			os.Unsetenv("SPARROW_BLOG_HOME")
		}
	}()
	
	// 获取配置文件路径
	configFilePath, err := getConfigFilePath()
	if err != nil {
		t.Fatalf("获取配置文件路径失败: %v", err)
	}
	
	// 确保配置文件不存在
	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		os.Remove(configFilePath)
	}
	
	// 触发配置加载（这应该创建默认配置文件）
	LoadConfig()
	
	// 验证配置文件是否被创建
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		t.Fatal("配置文件应该被创建，但没有找到")
	}
	
	// 验证日志路径不为空
	if Logger.Path == "" {
		t.Fatal("加载的日志路径不应为空")
	}
	
	if !strings.Contains(Logger.Path, "log/sparrow_blog.log") {
		t.Errorf("日志路径应包含 'log/sparrow_blog.log', 但得到: %s", Logger.Path)
	}
	
	t.Logf("✅ 默认配置文件生成测试通过")
	t.Logf("配置文件路径: %s", configFilePath)
	t.Logf("生成的日志路径: %s", Logger.Path)
}