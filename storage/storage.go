package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sparrow_blog_server/cache"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage/db/sqlite"
	"sparrow_blog_server/storage/ossstore"
	"sparrow_blog_server/storage/ossstore/aliyun"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"gorm.io/gorm"
)

var Storage *storage
var storageOnce sync.Once

// storage 结构体用于存储数据库和对象存储客户端的实例
type storage struct {
	Db        *gorm.DB     // 数据库连接
	OssClient *oss.Client  // 对象存储客户端
	Cache     *cache.Cache // 缓存客户端
}

// InitStorage 初始化 storage 组件
func InitStorage(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		storageOnce.Do(func() {
			Storage = &storage{}
			// TODO: 可以不止配置 MySql 一种数据库，现在先写死，后面根据数据库配置进行选择
			logger.Info("配置数据库")
			db, err := sqlite.ConnectSqlite(ctx)
			if err != nil {
				msg := fmt.Sprintf("连接数据库失败 %v", err)
				logger.Panic(msg)
				return
			}
			Storage.Db = db

			// TODO: 同样，对象存储桶也可以用其他的，现在先写死用阿里云 Oss
			logger.Info("配置对象存储")
			client, err := aliyun.ConnectOss(ctx)
			if err != nil {
				msg := fmt.Sprintf("连接对象存储失败 %v", err)
				logger.Panic(msg)
				return
			}
			Storage.OssClient = client

			logger.Info("配置缓存")
			c, err := cache.NewCache(ctx)
			if err != nil {
				msg := fmt.Sprintf("连接缓存失败 %v", err)
				logger.Panic(msg)
				return
			}
			Storage.Cache = c
		})

		return nil
	}
}

// DeleteObject 删除对象
// - ctx 上下文对象，用于控制请求的截止时间、取消信号等
// - oldName 要删除的对象名称
func (s *storage) DeleteObject(ctx context.Context, oldName string) error {
	// 构建删除对象的请求
	deleteRequest := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(config.Oss.Bucket), // 存储空间名称
		Key:    oss.Ptr(oldName),           // 要删除的对象名称
	}
	// 执行删除对象的操作
	deleteResult, err := s.OssClient.DeleteObject(ctx, deleteRequest)
	if err != nil {
		msg := fmt.Sprintf("删除 OssClient 对象失败 %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("删除源对象成功: %#v", deleteResult)

	return nil
}

// RenameObject 重命名对象
// - ctx 上下文对象，用于控制请求的截止时间、取消信号等
// - oldPath 原对象路径
// - newPath 新对象路径
func (s *storage) RenameObject(ctx context.Context, oldPath, newPath string) error {
	// 创建文件拷贝器
	c := s.OssClient.NewCopier()

	// 构建拷贝对象的请求
	copyRequest := &oss.CopyObjectRequest{
		Bucket:       oss.Ptr(config.Oss.Bucket), // 目标存储空间名称
		Key:          oss.Ptr(newPath),           // 目标对象名称
		SourceBucket: oss.Ptr(config.Oss.Bucket), // 源存储空间名称
		SourceKey:    oss.Ptr(oldPath),           // 源对象名称
		StorageClass: oss.StorageClassStandard,   // 指定存储类型为归档类型
	}
	// 执行拷贝对象的操作
	result, err := c.Copy(ctx, copyRequest)
	if err != nil {
		// 记录错误信息，并返回自定义错误信息
		msg := fmt.Sprintf("拷贝 OssClient 对象失败 %v", err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("拷贝对象成功: %#v", result)

	// 构建删除对象的请求
	deleteRequest := &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(config.Oss.Bucket), // 存储空间名称
		Key:    oss.Ptr(oldPath),           // 要删除的对象名称
	}
	// 执行删除对象的操作
	deleteResult, err := s.OssClient.DeleteObject(ctx, deleteRequest)
	if err != nil {
		// 记录错误信息，并返回自定义错误信息
		msg := fmt.Sprintf("删除 (%v) 对象失败 %v", oldPath, err)
		logger.Error(msg)
		return errors.New(msg)
	}
	logger.Info("删除源对象成功: %#v", deleteResult)

	return nil
}

// PutContentToOss 上传内容
// - ctx 上下文对象，用于控制请求的截止时间、取消信号等
// - content 要上传的内容
// - objectName 上传到 OssClient 的对象名称
func (s *storage) PutContentToOss(ctx context.Context, content []byte, objectName string) error {

	// 创建上传器
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(config.Oss.Bucket), // 指定上传的 Bucket 名称，使用 config 中的配置
		Key:    oss.Ptr(objectName),        // 指定上传的对象名称
		Body:   bytes.NewReader(content),   // 将内容转换为 Reader，作为上传的内容
	}

	// 使用上下文 ctx 开启上传请求
	result, err := s.OssClient.PutObject(ctx, request)

	// 上传文件失败
	if err != nil {
		msg := fmt.Sprintf("上传 (%v) 失败, 错误信息: %v", objectName, err)
		logger.Error(msg)
		// 返回自定义错误信息，避免暴露敏感信息
		return errors.New(msg)
	}

	// 记录成功信息
	logger.Info("上传 (%v) 文件成功: %#v", objectName, result.Status)
	return nil
}

// IsExist 检查指定的对象是否存在于存储桶中。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和取消操作。
//   - objectName: 要检查的对象名称。
//
// 返回值:
//   - bool: 如果对象存在，则返回 true；否则返回 false。
//   - error: 如果在检查过程中发生错误，则返回错误信息；否则返回 nil。
func (s *storage) IsExist(ctx context.Context, objectName string) (bool, error) {
	// 调用 OSS 客户端的 IsObjectExist 方法检查对象是否存在。
	result, err := s.OssClient.IsObjectExist(ctx, config.Oss.Bucket, objectName)
	if err != nil {
		// 如果检查失败，记录错误日志并返回封装后的错误信息。
		msg := fmt.Sprintf("判断 (%v) 对象是否存在失败 %v", objectName, err)
		logger.Error(msg)
		return false, errors.New(msg)
	}

	// 如果对象存在，返回 true 和 nil 错误。
	if result {
		return true, nil
	}

	// 如果对象不存在，返回 false 和 nil 错误。
	return false, nil
}

// GetContentFromOss 下载内容
// - ctx 上下文对象，用于控制请求的截止时间、取消信号等
// - objectName 下载的对象名称
// 返回下载的内容和错误信息
func (s *storage) GetContentFromOss(ctx context.Context, objectName string) ([]byte, error) {
	// 使用上下文 ctx 开启上传请求
	result, err := s.OssClient.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(config.Oss.Bucket), // 指定下载的 Bucket 名称，使用 config 中的配置
		Key:    oss.Ptr(objectName),        // 指定下载的对象名称
	})
	if err != nil {
		msg := fmt.Sprintf("获取 (%v) 文件失败 %v", objectName, err)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("OssClient 关闭 Reader 失败: " + err.Error())
		}
	}(result.Body)

	// 读取文件内容
	data, err := io.ReadAll(result.Body)
	if err != nil {
		msg := fmt.Sprintf("读取 objectName 文件失败: %s, 错误信息: %v", objectName, err)
		logger.Error(msg)
		return nil, errors.New(msg)
	}
	return data, nil
}

// ListOssDirFiles 列举指定目录下的所有对象
func (s *storage) ListOssDirFiles(ctx context.Context, dir string) ([]string, error) {
	// 创建列出对象的请求
	request := &oss.ListObjectsV2Request{
		Bucket: oss.Ptr(config.Oss.Bucket),
		Prefix: oss.Ptr(dir), // 列举指定目录下的所有对象
	}

	// 创建分页器
	p := s.OssClient.NewListObjectsV2Paginator(request)

	// 初始化结果数组
	var results []string

	// 初始化页码计数器
	var i int
	logger.Info("开始列举 (%v) 下的对象", dir)
	// 遍历分页器中的每一页
	for p.HasNext() {
		i++

		// 获取下一页的数据
		page, err := p.NextPage(ctx)
		if err != nil {
			logger.Error("获取第 %v 页数据失败: %v", i, err)
			return nil, err
		}

		// 打印该页中的每个对象的信息
		for _, obj := range page.Contents {
			results = append(results, oss.ToString(obj.Key))
		}
	}

	// 从第一个截取是因为第一个是目录名称
	return results[1:], nil
}

// GenPreSignUrl 生成一个预签名的 URL，用于访问或上传指定的对象。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和取消操作。
//   - objectName: 对象的名称（通常是文件名），用于标识存储桶中的具体对象。
//   - method: 操作方法，支持 "Get" 或 "Put"，分别表示获取对象和上传对象。
//
// 返回值:
//   - *oss.PresignResult: 包含预签名 URL 和相关元数据的结果对象。
//   - error: 如果发生错误，返回具体的错误信息。
func (s *storage) GenPreSignUrl(ctx context.Context, objectName, fileType, method string, duration time.Duration) (*oss.PresignResult, error) {
	// 根据传入的方法构造对应的请求体
	var request oss.RequestCommonInterface
	switch method {
	case ossstore.Get:
		// 构造获取对象的请求
		request = &oss.GetObjectRequest{
			Bucket: oss.Ptr(config.Oss.Bucket),
			Key:    oss.Ptr(objectName),
		}
	case ossstore.Put:
		// 构造上传对象的请求
		var header string
		switch fileType {
		case ossstore.MarkDown:
			header = ossstore.MarkdownHeader
		case ossstore.Webp:
			header = ossstore.WebpHeader
		default:
			return nil, errors.New("不支持的文件类型")
		}
		request = &oss.PutObjectRequest{
			Bucket:      oss.Ptr(config.Oss.Bucket),
			Key:         oss.Ptr(objectName),
			ContentType: oss.Ptr(header),
		}
	default:
		// 如果传入的方法不被支持，返回错误
		return nil, errors.New("不支持的方法")
	}

	// 调用 OSS 客户端生成预签名 URL，设置过期时间为 10 分钟
	result, err := s.OssClient.Presign(ctx, request, oss.PresignExpires(duration))

	if err != nil {
		// 如果生成预签名 URL 失败，记录错误日志并返回错误信息
		msg := fmt.Sprintf("生成预签名 URL 失败: %v", err)
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	// 返回生成的预签名结果
	return result, nil
}

// Close 关闭所有存储相关连接，包括数据库和缓存
// - ctx 上下文对象，用于控制请求的截止时间、取消信号等
func (s *storage) Close(ctx context.Context) {
	// 关闭数据库连接
	if s.Db != nil {
		// 使用上下文获取数据库实例
		sqlDB, err := s.Db.WithContext(ctx).DB()
		// 如果获取实例失败，记录错误并返回
		if err != nil {
			logger.Error("获取数据库实例失败: " + err.Error())
		} else {
			// 关闭数据库连接
			err = sqlDB.Close()
			// 如果关闭连接失败，记录错误
			if err != nil {
				logger.Error("关闭数据库连接失败: " + err.Error())
			} else {
				// 记录数据库连接已关闭的信息
				logger.Info("MySQL 数据库连接已关闭")
			}
		}
	}

	// 关闭缓存连接
	if s.Cache != nil {
		err := s.Cache.Close()
		if err != nil {
			logger.Error("关闭缓存失败: " + err.Error())
		} else {
			logger.Info("缓存已安全关闭")
		}
	}
}
