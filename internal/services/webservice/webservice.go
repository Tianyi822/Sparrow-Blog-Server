package webservice

import (
	"context"
	"fmt"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repositories/blogrepo"
	"h2blog_server/internal/repositories/categoryrepo"
	"h2blog_server/internal/repositories/tagrepo"
	"h2blog_server/pkg/config"
)

// GetHomeData 获取首页数据。
// 该函数通过查询配置和数据库来收集用户信息、博客、分类和标签等数据。
// 参数:
//   - ctx context.Context: 上下文对象，用于传递请求范围的 deadline、取消信号等。
//
// 返回值:
//   - map[string]any: 包含首页所需数据的映射，包括用户信息、博客、分类和标签等。
//   - error: 如果查询数据过程中发生错误，则返回该错误。
func GetHomeData(ctx context.Context) (map[string]any, error) {
	// 初始化结果映射，填充用户配置信息。
	result := map[string]any{
		"user_name":           config.User.Username,
		"user_email":          config.User.UserEmail,
		"user_github_address": config.User.UserGithubAddress,
		"user_hobbies":        config.User.UserHobbies,
		"type_writer_content": config.User.TypeWriterContent,
		"background_image":    config.User.BackgroundImage,
		"avatar_image":        config.User.AvatarImage,
		"web_logo":            config.User.WebLogo,
	}

	// 定义一个结构体用于存储查询结果和可能的错误。
	type resultData struct {
		Blogs      any
		Categories any
		Tags       any
		Err        error
	}

	// 创建一个带缓冲的通道，用于接收查询结果。
	ch := make(chan resultData, 3)

	// 启动三个协程，分别查询博客、分类和标签数据。
	go func() {
		blogDtos, err := blogrepo.FindAllBlogs(ctx, true)
		if err != nil {
			ch <- resultData{Err: fmt.Errorf("failed to find all blogs: %w", err)}
			return
		}

		var blogVos []dto.BlogDto
		for _, blogDto := range blogDtos {
			cat, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
			if err != nil {
				ch <- resultData{Err: fmt.Errorf("failed to find category by id: %w", err)}
				return
			}

			tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
			if err != nil {
				ch <- resultData{Err: fmt.Errorf("failed to find tags by blog id: %w", err)}
				return
			}

			blogVo := dto.BlogDto{
				BlogId:       blogDto.BlogId,
				BlogTitle:    blogDto.BlogTitle,
				BlogImageId:  blogDto.BlogImageId,
				BlogBrief:    blogDto.BlogBrief,
				BlogWordsNum: blogDto.BlogWordsNum,
				BlogIsTop:    blogDto.BlogIsTop,
				BlogState:    blogDto.BlogState,
				Category:     cat,
				Tags:         tags,
				CreateTime:   blogDto.CreateTime,
			}
			blogVos = append(blogVos, blogVo)
		}

		ch <- resultData{Blogs: blogVos}
	}()

	go func() {
		categories, err := categoryrepo.FindAllCategories(ctx)
		if err != nil {
			ch <- resultData{Err: fmt.Errorf("failed to get all categories: %w", err)}
			return
		}

		var cateVos []dto.CategoryDto
		for _, cate := range categories {
			cateVo := dto.CategoryDto{
				CategoryId:   cate.CategoryId,
				CategoryName: cate.CategoryName,
			}
			cateVos = append(cateVos, cateVo)
		}

		ch <- resultData{Categories: cateVos}
	}()

	go func() {
		tags, err := tagrepo.FindAllTags(ctx)
		if err != nil {
			ch <- resultData{Err: fmt.Errorf("failed to get all tags: %w", err)}
			return
		}

		cateVos := make([]dto.TagDto, 0, len(tags))
		for _, tag := range tags {
			cateVo := dto.TagDto{
				TagId:   tag.TagId,
				TagName: tag.TagName,
			}
			cateVos = append(cateVos, cateVo)
		}

		ch <- resultData{Tags: cateVos}
	}()

	for i := 0; i < 3; i++ {
		r := <-ch
		if r.Err != nil {
			return nil, r.Err
		}
		if r.Blogs != nil {
			result["blogs"] = r.Blogs
		} else if r.Categories != nil {
			result["categories"] = r.Categories
		} else if r.Tags != nil {
			result["tags"] = r.Tags
		}
	}

	// 返回包含所有首页数据的映射。
	return result, nil
}
