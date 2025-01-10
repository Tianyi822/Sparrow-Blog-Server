package aliyun

import (
	"context"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
)

// ConnectOss 连接 OSS
func ConnectOss() *oss.Client {
	// 创建一个静态凭证提供者，使用配置文件中的 AccessKeyId 和 AccessKeySecret
	provider := credentials.NewStaticCredentialsProvider(config.OSSConfig.AccessKeyId, config.OSSConfig.AccessKeySecret)
	// 加载默认配置，并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(config.OSSConfig.Region)

	// 创建 OSS Client
	// 使用配置信息创建一个新的 OSS 客户端
	ossClient := oss.NewClient(cfg)

	// 获取 Bucket 信息
	// 创建一个获取 Bucket 信息的请求，指定要获取的 Bucket 名称
	request := &oss.GetBucketInfoRequest{
		Bucket: oss.Ptr(config.OSSConfig.Bucket),
	}

	// 这里传入一个空的 context，只是用于检查是否连接成功，后续操作还是要传入项目的 context
	// 使用 OSS 客户端发送请求获取 Bucket 信息
	result, err := ossClient.GetBucketInfo(context.Background(), request)
	if err != nil {
		// 如果获取 Bucket 信息失败，记录错误日志
		logger.Error("获取 bucket 信息失败: %v", err)
	}

	// 记录获取 Bucket 信息的响应状态码
	logger.Info("获取 bucket 信息: %v\n", result.StatusCode)

	// 返回创建的 OSS 客户端
	return ossClient
}
