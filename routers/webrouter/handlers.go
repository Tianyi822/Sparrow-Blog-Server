package webrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
)

// userBasicInfo 返回用户的基本信息。
func userBasicInfo(ctx *gin.Context) {
	// 构造并返回成功的 JSON 响应，包含用户名和用户邮箱信息。
	resp.Ok(ctx, "获取成功", map[string]any{
		"user_name":           config.User.Username,
		"user_email":          config.User.UserEmail,
		"user_github_address": config.User.UserGithubAddress,
		"user_hobbies":        config.User.UserHobbies,
		"type_writer_content": config.User.TypeWriterContent,
		"background_image":    config.User.BackgroundImage,
		"avatar_image":        config.User.AvatarImage,
		"web_logo":            config.User.WebLogo,
	})
}
