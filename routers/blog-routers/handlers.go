package blog_routers

import (
	"github.com/gin-gonic/gin"
	"h2blog/internal/model/dto"
	"h2blog/internal/services/blogService"
	"h2blog/pkg/resp"
	"h2blog/routers/tools"
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

// modifyBlog 用于修改博客信息
//   - ctx 是 Gin 框架的上下文对象，用于处理 HTTP 请求和响应
func modifyBlog(ctx *gin.Context) {
	blogDto, err := tools.GetBlogDto(ctx)
	if err != nil {
		return
	}

	num, err := blogService.ModifyH2BlogInfo(ctx, blogDto)
	if err != nil {
		resp.Err(ctx, err.Error(), num)
		return
	}
	resp.Ok(ctx, "修改博客成功", num)
}

func deleteBlogById(ctx *gin.Context) {
	blogDto := &dto.BlogInfoDto{
		BlogId: ctx.Param("blog_id"),
	}

	num, err := blogService.DeleteH2BlogInfo(ctx, blogDto)
	if err != nil {
		resp.Err(ctx, err.Error(), num)
		return
	}
	resp.Ok(ctx, "删除博客成功", num)
}

func addBlogInfo(ctx *gin.Context) {
	blogDto, err := tools.GetBlogDto(ctx)
	if err != nil {
		return
	}

	num, err := blogService.AddH2BlogInfo(ctx, blogDto)
	if err != nil {
		resp.Err(ctx, err.Error(), -1)
		return
	}

	resp.Ok(ctx, "创建博客成功", num)
}
