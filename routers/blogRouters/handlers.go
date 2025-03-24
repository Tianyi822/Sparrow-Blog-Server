package blogRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/pkg/resp"
)

// 获取指定博客的详细信息
func getBlogById(ctx *gin.Context) {

	resp.Ok(ctx, "获取博客成功", nil)
}
