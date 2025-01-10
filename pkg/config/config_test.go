package config

import (
	"fmt"
	"os"
	"testing"
)

func TestValidConfig(t *testing.T) {
	testData := struct {
		configFile  string
		configData  string
		shouldPanic bool
	}{
		configFile: "test_config.yml",
		configData: `
# 日志配置
logger:
  # 日志级别
  level: debug
  # 日志文件路径
  path: resources/config/test/logger-config.yaml
  # 日志文件最大大小，单位-MB
  max_size: 3
  # 日志文件最大备份数量
  max_backups: 30
  # 日志文件最大保存时间，单位-天
  max_age: 7
  # 是否压缩日志文件
  compress: true`,
		shouldPanic: false,
	}

	testInstance(t, testData)
}

func TestNonexistentConfig(t *testing.T) {
	testData := struct {
		configFile  string
		configData  string
		shouldPanic bool
	}{
		configFile:  "nonexistent.yml",
		configData:  "",
		shouldPanic: true,
	}

	testInstance(t, testData)
}

func TestInvalidConfig(t *testing.T) {
	testData := struct {
		configFile  string
		configData  string
		shouldPanic bool
	}{
		configFile:  "invalid_config.yml",
		configData:  "invalid: {[yaml",
		shouldPanic: true,
	}

	testInstance(t, testData)
}

func testInstance(t *testing.T, testData struct {
	configFile  string
	configData  string
	shouldPanic bool
}) {
	if testData.configData != "" {
		err := os.WriteFile(testData.configFile, []byte(testData.configData), 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(testData.configFile)
	}

	if testData.shouldPanic {
		defer func() {
			if r := recover(); r == nil {
				t.Error("LoadConfig() should panic")
			}
		}()
	}

	LoadConfig(testData.configFile)
}

func TestMySQLConfig(t *testing.T) {
	LoadConfig("../../resources/config/test/storage-config.yaml")
	if MySQLConfig.User != "root" {
		t.Errorf("MySQLConfig.User should be 'root', but got %s", MySQLConfig.User)
	}
	fmt.Println(MySQLConfig)
}

func TestServerConfig(t *testing.T) {
	LoadConfig("../../resources/config/test/web-config.yaml")
	if ServerConfig.Port != 2233 {
		t.Errorf("ServerConfig.Port should be 8080, but got %d", ServerConfig.Port)
	}
	fmt.Println(ServerConfig)
}

func TestOssConfig(t *testing.T) {
	LoadConfig("../../resources/config/test/storage-config.yaml")
	if OSSConfig.Region != "cn-guangzhou" {
		t.Errorf("OSSConfig.Region should be 'cn-guangzhou', but got %s", OSSConfig.Region)
	}
	fmt.Println(OSSConfig)
}

func TestUserConfig(t *testing.T) {
	LoadConfig("../../resources/config/test/web-config.yaml")
	if UserConfig.Username != "chentyit" {
		t.Errorf("UserConfig.Username should be 'chentyit', but got %s", UserConfig.Username)
	}
	// 结构化输出
	fmt.Println(UserConfig)
}
