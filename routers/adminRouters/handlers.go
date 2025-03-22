package adminRouters

import (
	"errors"
	"github.com/gin-gonic/gin"
	"h2blog_server/cache"
	"h2blog_server/email"
	"h2blog_server/env"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"
	"h2blog_server/storage"
)

// sendVerificationCode 处理发送验证码的请求。
// 参数:
//   - *gin.Context: HTTP 请求上下文，包含请求数据和响应方法。
//
// 功能描述:
//
//	该函数从请求中解析用户提交的数据，验证用户邮箱是否正确，
//	并调用邮件服务发送验证码。根据操作结果返回相应的 HTTP 响应。
func sendVerificationCode(ctx *gin.Context) {
	// 从请求中解析原始数据为 map，并处理可能的解析错误。
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "登录信息解析错误", err.Error())
		return
	}

	// 验证用户提交的邮箱是否与配置中的用户邮箱一致。
	if rawData["user_email"].(string) != config.User.UserEmail {
		resp.BadRequest(ctx, "用户邮箱错误", "")
		return
	}

	// 调用邮件服务发送验证码邮件，并处理发送过程中可能出现的错误。
	err = email.SendVerificationCodeEmail(ctx, config.User.UserEmail)
	if err != nil {
		resp.Err(ctx, "验证码发送失败", err.Error())
		return
	}

	// 如果验证码发送成功，返回成功的 HTTP 响应。
	resp.Ok(ctx, "验证码发送成功", nil)
}

// login 处理用户登录请求。
// 参数:
//   - Gin 上下文，用于处理 HTTP 请求和响应。
//
// 功能描述:
//  1. 从请求中解析原始数据，并验证其格式。
//  2. 检查用户邮箱是否正确。
//  3. 从缓存中获取验证码并验证其有效性。
//  4. 验证用户提交的验证码是否匹配。
//  5. 如果所有验证通过，返回登录成功的响应。
func login(ctx *gin.Context) {
	// 从请求中解析原始数据为 Map 格式
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "登录信息解析错误", err.Error())
		return
	}

	// 验证用户邮箱是否与配置中的邮箱一致
	if rawData["user_email"].(string) != config.User.UserEmail {
		resp.BadRequest(ctx, "用户邮箱错误", "")
		return
	}

	// 从缓存中获取验证码
	verCode, err := storage.Storage.Cache.GetString(ctx, env.VerificationCodeKey)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			// 如果验证码未找到，说明验证码已过期
			resp.BadRequest(ctx, "验证码过期", err.Error())
			return
		}
		// 其他缓存获取错误
		resp.Err(ctx, "验证码缓存获取失败", err.Error())
	}

	// 验证用户提交的验证码是否与缓存中的验证码匹配
	if rawData["verified_code"].(string) != verCode {
		resp.BadRequest(ctx, "验证码错误", "")
		return
	}

	if err = storage.Storage.Cache.Delete(ctx, env.VerificationCodeKey); err != nil {
		resp.Err(ctx, "验证码缓存删除失败", err.Error())
		return
	}

	// TODO: 这里应该返回一个 Token，但现在是开发状态，暂时不实现
	resp.Ok(ctx, "登录成功", nil)
}
