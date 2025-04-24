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
		defer func(name string) {
			_ = os.Remove(name)
		}(testData.configFile)
	}

	if testData.shouldPanic {
		defer func() {
			if r := recover(); r == nil {
				t.Error("LoadConfig() should panic")
			}
		}()
	}

	_ = LoadConfig()
}

func TestMySQLConfig(t *testing.T) {
	_ = LoadConfig()
	if MySQL.User != "root" {
		t.Errorf("MySQL.User should be 'root', but got %s", MySQL.User)
	}
	fmt.Println(MySQL)
}

func TestServerConfig(t *testing.T) {
	_ = LoadConfig()
	if Server.Port != 2233 {
		t.Errorf("Server.Port should be 8080, but got %d", Server.Port)
	}
	fmt.Println(Server)

	t.Logf("SMTP 服务配置: %v", Server.SmtpAccount)
}

func TestOssConfig(t *testing.T) {
	_ = LoadConfig()
	if Oss.Region != "cn-guangzhou" {
		t.Errorf("Oss.Region should be 'cn-guangzhou', but got %s", Oss.Region)
	}
	fmt.Println(Oss)
}

func TestUserConfig(t *testing.T) {
	err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
		return
	}

	t.Logf("用户爱好: %#v", User.UserHobbies)
	t.Logf("打字机内容: %#v", User.TypeWriterContent)
}

func TestCacheConfig(t *testing.T) {
	_ = LoadConfig()
	if Cache.Aof.MaxSize != 3 {
		t.Errorf("Cache.Aof.MaxSize should be 3, but got %d", Cache.Aof.MaxSize)
	}
	fmt.Println(Cache)
}

func TestGetBackgroundImgName(t *testing.T) {
	_ = LoadConfig()
	fmt.Println(User.BackgroundImage)
}
