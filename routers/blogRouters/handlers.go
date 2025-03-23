package blogRouters

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/internal/services/blogService"
	"h2blog_server/pkg/resp"
)

// 获取指定博客的详细信息
func getBlogById(ctx *gin.Context) {
	blogId := ctx.Param("blog_id")
	blog, err := blogService.GetH2BlogInfoById(ctx, blogId)
	if err != nil {
		resp.Err(ctx, err.Error(), -1)
		return
	}

	resp.Ok(ctx, "获取博客成功", blog)
}
