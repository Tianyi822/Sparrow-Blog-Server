package blogService

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/internal/model/vo"
	"h2blog_server/internal/repository/blogInfoRepo"
)

// GetH2BlogInfoById 用于获取指定博客信息
//   - ctx 是 Gin 框架的上下文对象，用于处理 HTTP 请求和响应
//   - blogId 是要获取的博客的唯一标识符
//
// 返回值
//   - *dto.BlogInfoDto 表示获取到的博客信息
//   - error 表示获取过程中可能发生的错误
func GetH2BlogInfoById(ctx *gin.Context, blogId string) (*vo.BlogInfoVo, error) {
	// 先查询是否有该数据
	blogInfoPo, err := blogInfoRepo.FindBlogById(ctx, blogId)
	// 检查是否有错误发生
	if err != nil {
		// 如果有错误，直接返回当前num值和错误信息
		return nil, err
	} else {
		return &vo.BlogInfoVo{
			BlogId: blogInfoPo.BId,
			Title:  blogInfoPo.Title,
			Brief:  blogInfoPo.Brief,
		}, nil
	}
}
