package adminService

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/cache"
	"h2blog_server/env"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repository/blogRepo"
	"h2blog_server/internal/repository/categoryRepo"
	"h2blog_server/internal/repository/tagRepo"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/webjwt"
	"h2blog_server/storage"
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
		msg := "用户邮箱不一致"
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	// 从缓存中获取存储的验证码
	verCodeInCache, err := storage.Storage.Cache.GetString(ctx, env.VerificationCodeKey)
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
	if err = storage.Storage.Cache.Delete(ctx, env.VerificationCodeKey); err != nil {
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

// GetBlogsToAdminPosts 获取所有博客及其关联的标签和分类信息。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客及其关联标签和分类信息的 DTO 列表。
//   - error: 如果在查询博客、标签或分类时发生错误，则返回该错误。
func GetBlogsToAdminPosts(ctx context.Context) ([]*dto.BlogDto, error) {
	blogDtos, err := blogRepo.FindAllBlogs(ctx, false)
	if err != nil {
		return nil, err
	}

	// 遍历博客列表，为每个博客获取其关联的标签数据。
	for _, blogDto := range blogDtos {
		tags, err := tagRepo.FindTagsByBlogId(ctx, blogDto.BlogId)
		if err != nil {
			return nil, err
		}

		blogDto.Tags = tags
	}

	// 遍历博客列表，为每个博客获取其关联的分类数据。
	for _, blogDto := range blogDtos {
		category, err := categoryRepo.FindCategoryById(ctx, blogDto.CategoryId)
		if err != nil {
			return nil, err
		}
		blogDto.Category = dto.CategoryDto{
			CategoryId:   category.CategoryId,
			CategoryName: category.CategoryName,
		}
	}

	return blogDtos, nil
}

// GetAllCategoriesAndTags 获取所有的分类和标签信息。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递上下文信息。
//
// 返回值:
//   - []*dto.CategoryDto: 包含所有分类信息的 DTO 列表。
//   - []*dto.TagDto: 包含所有标签信息的 DTO 列表。
//   - error: 如果在获取分类或标签时发生错误，则返回具体的错误信息；否则返回 nil。
func GetAllCategoriesAndTags(ctx context.Context) ([]*dto.CategoryDto, []*dto.TagDto, error) {
	categories, err := categoryRepo.GetAllCategories(ctx)
	if err != nil {
		return nil, nil, err
	}

	tags, err := tagRepo.GetAllTags(ctx)
	if err != nil {
		return nil, nil, err
	}

	return categories, tags, nil
}

// UpdateOrAddBlog 更新或添加博客信息，并处理相关的分类和标签逻辑。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//   - blogDto: 包含博客信息的数据传输对象，包括博客内容、分类和标签等信息。
//
// 返回值:
//   - error: 如果操作过程中发生错误，则返回具体的错误信息；否则返回 nil。
func UpdateOrAddBlog(ctx context.Context, blogDto *dto.BlogDto) error {
	// 开启事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("更新或创建博客数据失败: %v", r)
			tx.Rollback()
		}
	}()

	// 如果 blogDto 中没有 CategoryId，则表示该分类是新的，需要新建分类。
	if len(blogDto.CategoryId) == 0 {
		categoryDto := dto.CategoryDto{
			CategoryName: blogDto.Category.CategoryName,
		}
		err := categoryRepo.AddCategory(tx, &categoryDto)
		if err != nil {
			tx.Rollback()
			return err
		}
		blogDto.CategoryId = categoryDto.CategoryId
	}

	// 保存新建的标签
	var newTags []dto.TagDto

	// 检查 blogDto 中的标签 ID，如果标签没有 ID，则需要创建新的标签。
	if len(blogDto.Tags) != 0 {
		// 分离不携带 ID 的标签和携带 ID 的标签
		var tagsWithoutId []dto.TagDto
		var tagsWithId []dto.TagDto
		for _, tag := range blogDto.Tags {
			if tag.TagId != "" {
				tagsWithId = append(tagsWithId, tag)
			} else {
				tagsWithoutId = append(tagsWithoutId, tag)
			}
		}

		// 如果存在没有 ID 的标签，则调用仓库方法批量创建新标签。
		if len(tagsWithoutId) != 0 {
			var err error
			newTags, err = tagRepo.AddTags(tx, tagsWithoutId)
			if err != nil {
				tx.Rollback()
				return err
			}
			// 将有 ID 的标签和新创建的标签合并回 blogDto。
			blogDto.Tags = append(tagsWithId, newTags...)
		}
	}

	// 根据 blogDto 是否包含 BlogId 判断是新增博客还是更新博客。
	if len(blogDto.BlogId) == 0 {
		if err := blogRepo.AddBlog(tx, blogDto); err != nil {
			tx.Rollback()
			return err
		}

		// 建立标签与博客的关联关系
		if err := tagRepo.AddBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := blogRepo.UpdateBlog(tx, blogDto); err != nil {
			tx.Rollback()
			return err
		}

		// 更新标签与博客的关联关系
		if err := tagRepo.UpdateBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	tx.Commit()
	logger.Info("完成博客的更新或创建操作")

	return nil
}

// DeleteBlog 删除指定ID的博客，并根据相关联的数据进行清理操作。
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlog(ctx context.Context, id string) error {
	// 获取与博客关联的分类ID。
	categoryId, err := blogRepo.GetCategoryIdByBlogId(ctx, id)
	if err != nil {
		return err
	}

	// 获取与博客关联的所有标签。
	tags, err := tagRepo.FindTagsByBlogId(ctx, id)
	if err != nil {
		return err
	}

	// 开启删除博客事务
	deleteBlogTx := storage.Storage.Db.WithContext(ctx).Begin()
	// 调用仓库方法根据ID删除博客。
	err = blogRepo.DeleteBlogById(deleteBlogTx, id)
	if err != nil {
		deleteBlogTx.Rollback()
		return err
	}

	// 删除博客标签关联数据
	err = tagRepo.DeleteBlogTagAssociationByBlogId(deleteBlogTx, id)
	if err != nil {
		deleteBlogTx.Rollback()
		return err
	}
	// 博客删除就提交，以便删除后续的标签和分类
	deleteBlogTx.Commit()

	// 开启维护分类和标签数据的事务
	catTagTx := storage.Storage.Db.WithContext(ctx).Begin()
	// 统计该分类下剩余的博客数量。
	num, err := blogRepo.CalBlogsCountByCategoryId(ctx, categoryId)
	if err != nil {
		return err
	}

	// 如果该分类下没有博客，则删除该分类。
	if num == 0 {
		err = categoryRepo.DeleteCategoryById(catTagTx, categoryId)
		if err != nil {
			catTagTx.Rollback()
			return err
		}
	}

	// 遍历所有与博客关联的标签，检查每个标签是否还有其他博客关联
	tagsWithoutBlog := make([]dto.TagDto, 0)
	for _, tag := range tags {
		num, err = tagRepo.CalBlogsCountByTagId(ctx, tag.TagId)
		if err != nil {
			return err
		}

		// 如果某个标签没有其他博客关联，则删除该标签。
		if num == 0 {
			tagsWithoutBlog = append(tagsWithoutBlog, tag)
		}
	}

	if len(tagsWithoutBlog) != 0 {
		err = tagRepo.DeleteTags(catTagTx, tagsWithoutBlog)
		if err != nil {
			catTagTx.Rollback()
			return err
		}
	}
	// 标签和分类数据维护完成，提交事务
	catTagTx.Commit()
	logger.Info("删除博客数据成功")

	return nil
}

func SetTop(ctx context.Context, id string) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()

	if err := blogRepo.SetTopById(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func ChangeBlogState(ctx context.Context, id string) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()

	if err := blogRepo.ChangeBlogStateById(tx, id); err != nil {
		return err
	}

	tx.Commit()

	return nil
}
