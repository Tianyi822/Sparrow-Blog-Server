package adminservices

import (
	"context"
	"errors"
	"fmt"
	"sparrow_blog_server/cache"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/pkg/webjwt"
	"sparrow_blog_server/storage"
	"time"
)

// Login 函数用于验证用户登录信息。
// 参数：
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//   - email: 用户提供的邮箱地址，用于验证用户身份。
//   - verificationCode: 用户提供的验证码，用于验证用户输入的正确性。
//
// 返回值：
//   - string: 登录成功后返回的 Token（当前开发阶段未实现，返回空字符串）。
//   - error: 如果验证失败或发生错误，返回相应的错误信息。
func Login(ctx context.Context, email, verificationCode string) (string, error) {
	// 检查用户邮箱是否与配置中的邮箱一致
	if email != config.User.UserEmail {
		msg := fmt.Sprintf("登录邮箱 %v 与配置邮箱不一致", email)
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	// 从缓存中获取存储的验证码
	verCodeInCache, err := storage.Storage.Cache.GetString(ctx, storage.VerificationCodeKey)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			// 验证码不存在或已过期
			logger.Warn("验证码过期")
			return "", errors.Join(err, errors.New("验证码过期"))
		}
		// 处理其他缓存获取错误
		msg := fmt.Sprintf("验证码缓存获取失败: %v", err.Error())
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	// 验证用户提供的验证码是否与缓存中的验证码一致
	if verCodeInCache != verificationCode {
		msg := "验证码错误"
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	// 尝试删除缓存中的验证码，避免重复使用
	// 删除失败不会影响系统功能，仅记录日志
	if err = storage.Storage.Cache.Delete(ctx, storage.VerificationCodeKey); err != nil {
		logger.Warn("删除验证码缓存失败: %v", err)
	}

	// 使用 JWT 工具生成并返回 Token
	token, err := webjwt.GenerateJWTToken()
	if err != nil {
		msg := fmt.Sprintf("生成 Token 失败: %v", err.Error())
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	return token, nil
}

// Logout 函数用于处理用户登出操作。
// 参数：
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//   - token: 用户当前的登录令牌，用于标识要登出的用户会话。
//
// 返回值：
//   - error: 如果登出过程中发生错误，返回相应的错误信息；否则返回 nil。
//
// 函数流程：
//  1. 从缓存中删除用户的登录令牌
//  2. 将令牌加入黑名单，防止令牌被重复使用
func Logout(ctx context.Context, token string) error {
	logger.Info("将当前 token 加入黑名单")
	revokedTokenKey := fmt.Sprintf("%v%v", storage.UserRevokedTokenKeyPre, token)
	err := storage.Storage.Cache.SetWithExpired(ctx, revokedTokenKey, token, time.Duration(config.Server.TokenExpireDuration)*24*time.Hour)
	if err != nil {
		msg := fmt.Sprintf("缓存 token 黑名单失败: %v", err.Error())
		logger.Warn(msg)
		return errors.New(msg)
	}

	return nil
}

// UpdateConfig 更新项目配置信息。
// 该函数从 config 包中获取当前项目的配置信息，包括用户、服务器、MySQL、OSS、缓存和日志等设置，
// 并将这些配置信息存储到一个 projConfig 结构体中。随后，调用 projConfig 的 Store 方法将配置信息持久化。
// 如果 Store 方法返回错误，UpdateConfig 函数会将此错误返回，否则返回 nil，表示配置更新成功。
func UpdateConfig() error {
	// 创建一个 projConfig 实例，填充当前项目的配置信息。
	projConfig := config.ProjectConfig{
		User:         config.User,
		Server:       config.Server,
		MySQL:        config.MySQL,
		Oss:          config.Oss,
		Cache:        config.Cache,
		Logger:       config.Logger,
		SearchEngine: config.SearchEngine,
	}

	// 调用 projConfig 的 Store 方法，尝试将配置信息持久化。
	// 如果 Store 方法返回错误，直接返回该错误。
	err := projConfig.Store()
	if err != nil {
		return err
	}

	// 如果配置信息持久化成功，返回 nil 表示操作成功。
	return nil
}
