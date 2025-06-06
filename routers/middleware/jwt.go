package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/webjwt"
	"sparrow_blog_server/routers/resp"
	"sparrow_blog_server/storage"
	"strings"
)

// AnalyzeJWT 用于分析和验证请求中的 JWT 令牌
// 确保请求的用户身份合法性，以及令牌的信息是否与配置中的用户信息相匹配
func AnalyzeJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取请求头中的Authorization字段值
		authorization := ctx.GetHeader("Authorization")
		// 检查Authorization值是否为空
		if len(strings.TrimSpace(authorization)) == 0 {
			// 如果为空，则返回未授权的错误响应，并中断请求处理
			resp.TokenIsUnauthorized(ctx, "请先登录", nil)
			ctx.Abort()
			return
		}

		// 判断该 token 是否在黑名单中
		userRevokedTokenKey := fmt.Sprintf("%v%v", storage.UserRevokedTokenKeyPre, authorization)
		getString, cacheErr := storage.Storage.Cache.GetString(ctx, userRevokedTokenKey)
		// 如果在黑名单中，则返回错误信息，并中断请求处理
		if cacheErr == nil && getString != "" {
			resp.TokenIsUnauthorized(ctx, "token 已被禁用", nil)
			ctx.Abort()
			return
		}

		// 解析JWT令牌，获取claims信息
		claims, err := webjwt.ParseJWTToken(authorization)
		// 如果解析失败，则返回错误信息，并中断请求处理
		if err != nil {
			resp.TokenIsUnauthorized(ctx, "token 解析失败", err.Error())
			ctx.Abort()
			return
		}

		// 验证claims中的用户邮箱是否与配置中的用户邮箱相匹配
		if claims.UserEmail != config.User.UserEmail {
			// 如果不匹配，则返回错误信息，并中断请求处理
			resp.TokenIsUnauthorized(ctx, "token 数据错误：邮箱不匹配", nil)
			ctx.Abort()
			return
		}

		// 验证claims中的用户名是否与配置中的用户名相匹配
		if claims.UserName != config.User.Username {
			// 如果不匹配，则返回错误信息，并中断请求处理
			resp.TokenIsUnauthorized(ctx, "token 数据错误：用户名不匹配", nil)
			ctx.Abort()
			return
		}

		// 将 token 值放置到全局中，方便后续使用
		ctx.Set("token", authorization)

		// 如果所有验证都通过，则继续执行下一个中间件或处理函数
		ctx.Next()
	}
}
