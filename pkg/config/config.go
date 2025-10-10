package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sparrow_blog_server/pkg/filetool"
	"sync"

	"gopkg.in/yaml.v3"
)

// 全局配置变量
var (
	// loadConfigOnce 确保配置只被加载一次
	loadConfigOnce sync.Once

	// IsFirstRun 标识是否为首次运行（配置文件不存在）
	IsFirstRun bool

	// User 保存全局用户配置
	User UserConfigData

	// Server 保存全局服务器配置
	Server ServerConfigData

	// Logger 保存全局日志配置
	Logger LoggerConfigData

	// SearchEngine 搜索引擎配置
	SearchEngine SearchEngineData

	// Sqlite 配置
	Sqlite SqliteConfig

	// Oss 保存全局OSS配置
	Oss OssConfig

	// Cache 保存全局缓存配置
	Cache CacheConfig
)

// LoadConfig 加载配置文件。
// 该函数的主要功能是检查并加载 H2Blog 的配置文件，确保配置文件存在并正确加载。
// 如果配置文件不存在或加载过程中发生错误，会返回相应的错误或触发 panic。
func LoadConfig() {
	// 使用 sync.Once 确保配置文件只加载一次。
	// 如果加载过程中发生错误，直接触发 panic。
	loadConfigOnce.Do(func() {
		// 判断项目目录是否有配置文件
		// 若有，则加载该文件
		// 若没有，则按照 items.go 结构创建配置文件并保存在项目目录中
		configFilePath, err := getConfigFilePath()
		if err != nil {
			panic(fmt.Errorf("获取配置文件路径失败: %w", err))
		}

		if !filetool.IsExist(configFilePath) {
			// 设置首次运行标识
			IsFirstRun = true

			// 创建带有默认值的配置文件
			f, createErr := filetool.CreateFile(configFilePath)
			if createErr != nil {
				panic(createErr)
			}
			pc, defaultErr := createDefaultConfig()
			if defaultErr != nil {
				panic(fmt.Errorf("创建默认配置失败: %w", defaultErr))
			}
			yamlData, marshalErr := yaml.Marshal(pc)
			if marshalErr != nil {
				panic(fmt.Errorf("将配置数据转换为 YAML 失败: %w", marshalErr))
			}
			_, wErr := f.Write(yamlData)
			if wErr != nil {
				panic(fmt.Errorf("写入 YAML 数据到文件失败: %w", wErr))
			}
			closeErr := f.Close()
			if closeErr != nil {
				panic(fmt.Errorf("关闭文件失败: %w", closeErr))
			}
		} else {
			// 配置文件存在，不是首次运行
			IsFirstRun = false
		}

		// 从配置文件中加载配置
		// 读取配置文件内容
		data, readErr := os.ReadFile(configFilePath)
		if readErr != nil {
			// 如果读取配置文件时发生错误，返回错误信息
			panic(fmt.Errorf("load config file error: %w", readErr))
		}
		// 解析配置数据
		conf := &ProjectConfig{}
		if unErr := yaml.Unmarshal(data, &conf); unErr != nil {
			// 如果解析配置到结构体时发生错误，返回错误信息
			panic(fmt.Errorf("reflect config to struct error: %w", unErr))
		}

		// 设置全局配置
		User = conf.User
		Server = conf.Server
		Logger = conf.Logger
		SearchEngine = conf.SearchEngine
		Sqlite = conf.Sqlite
		Oss = conf.Oss
		Cache = conf.Cache
	})
}

// createDefaultConfig 创建带有默认值的配置
//
// 返回值:
//   - *ProjectConfig: 包含默认值的配置对象
//   - error: 创建过程中遇到的任何错误
func createDefaultConfig() (*ProjectConfig, error) {
	projDir, err := getProjDir()
	if err != nil {
		return nil, fmt.Errorf("获取项目目录失败: %w", err)
	}

	return &ProjectConfig{
		User: UserConfigData{},
		Server: ServerConfigData{
			Port: 8080,
		},
		Logger: LoggerConfigData{
			Level:      "info",
			Path:       filepath.Join(projDir, "log", "sparrow_blog.log"),
			MaxAge:     7,
			MaxSize:    10,
			MaxBackups: 3,
			Compress:   true,
		},
		SearchEngine: SearchEngineData{
			IndexPath: filepath.Join(projDir, "index", "sparrow_blog.bleve"),
		},
		Sqlite: SqliteConfig{
			Path: filepath.Join(projDir, "data", "sparrow_blog.db"),
		},
		Oss: OssConfig{},
		Cache: CacheConfig{
			Aof: AofConfig{
				Enable:   true,
				Path:     filepath.Join(projDir, "aof", "sparrow_blog.aof"),
				MaxSize:  10,
				Compress: true,
			},
		},
	}, nil
}

// getConfigFilePath 获取配置文件的完整路径
//
// 返回值:
//   - string: 配置文件的完整路径
//   - error: 获取过程中遇到的任何错误
func getConfigFilePath() (string, error) {
	projDir, err := getProjDir()
	if err != nil {
		return "", fmt.Errorf("获取项目目录失败: %w", err)
	}
	return filepath.Join(projDir, "config", "sparrow_blog_config.yaml"), nil
}

// Store 将配置保存到文件中
//
// 返回值:
// - error: 如果保存过程中发生错误，则返回相应的错误信息
func (pc *ProjectConfig) Store() error {
	h2BlogHomePath, err := getProjDir()
	if err != nil {
		return fmt.Errorf("获取或创建 H2Blog 目录失败: %w", err)
	}

	// 将 ProjectConfig 结构体转换为 YAML 格式的字节数组
	yamlData, err := yaml.Marshal(pc)
	if err != nil {
		return fmt.Errorf("将配置数据转换为YAML失败: %w", err)
	}

	// 生成配置文件路径
	configPath := filepath.Join(h2BlogHomePath, "config", "sparrow_blog_config.yaml")

	// 删除原有的配置文件
	if filetool.IsExist(configPath) {
		removeErr := os.Remove(configPath)
		if removeErr != nil {
			return fmt.Errorf("删除旧的配置文件失败: %w", removeErr)
		}
	}

	// 将新的 YAML 数据写入到文件中
	file, err := filetool.CreateFile(configPath)
	if err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	_, err = file.Write(yamlData)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("同步配置文件失败: %w", err)
	}

	return nil
}

// getProjDir 返回h2blog配置和数据的基础目录
// 它使用用户的主目录作为基础路径
//
// 返回:
//   - string: h2blog目录的完整路径
//   - error: 获取用户主目录时遇到的任何错误
func getProjDir() (string, error) {
	// 从环境变量中获取
	if h2blogHome := os.Getenv("SPARROW_BLOG_HOME"); h2blogHome != "" {
		return h2blogHome, nil
	}

	// 获取当前用户信息
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("获取当前用户失败: %w", err)
	}

	// 将主目录与.h2blog连接
	return filepath.Join(usr.HomeDir, ".sparrow_blog"), nil
}
