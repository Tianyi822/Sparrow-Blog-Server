package adminservice

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/cache"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/repositories/blogrepo"
	"h2blog_server/internal/repositories/categoryrepo"
	"h2blog_server/internal/repositories/imgrepo"
	"h2blog_server/internal/repositories/tagrepo"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/webjwt"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
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
		msg := "用户邮箱不一致"
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

// GetBlogsToAdminPosts 获取所有博客及其关联的标签和分类信息。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递元数据。
//
// 返回值:
//   - []*dto.BlogDto: 包含博客及其关联标签和分类信息的 DTO 列表。
//   - error: 如果在查询博客、标签或分类时发生错误，则返回该错误。
func GetBlogsToAdminPosts(ctx context.Context) ([]*dto.BlogDto, error) {
	blogDtos, err := blogrepo.FindAllBlogs(ctx, false)
	if err != nil {
		return nil, err
	}

	// 遍历博客列表，为每个博客获取其关联的标签数据。
	for _, blogDto := range blogDtos {
		tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
		if err != nil {
			return nil, err
		}

		blogDto.Tags = tags
	}

	// 遍历博客列表，为每个博客获取其关联的分类数据。
	for _, blogDto := range blogDtos {
		category, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
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
	categories, err := categoryrepo.GetAllCategories(ctx)
	if err != nil {
		return nil, nil, err
	}

	tags, err := tagrepo.GetAllTags(ctx)
	if err != nil {
		return nil, nil, err
	}

	return categories, tags, nil
}

// GetBlogData 根据博客ID获取博客的详细信息，包括关联的标签和分类。
//
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要查询的博客的唯一标识符。
//
// 返回值:
//   - *dto.BlogDto: 包含博客详细信息的数据传输对象，包括标签和分类信息。
//   - string: 预签名 URL，用于读取 OSS 中的文章内容。
//   - error: 如果查询过程中发生错误，则返回错误信息；否则返回 nil。
func GetBlogData(ctx context.Context, id string) (*dto.BlogDto, string, error) {
	// 根据博客ID从仓库中获取博客的基础信息。
	blogDto, err := blogrepo.FindBlogById(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// 根据博客ID获取与该博客关联的所有标签。
	tags, err := tagrepo.FindTagsByBlogId(ctx, blogDto.BlogId)
	if err != nil {
		return nil, "", err
	}
	blogDto.Tags = tags

	// 根据博客的分类ID获取分类信息，并将其映射为 CategoryDto 对象。
	category, err := categoryrepo.FindCategoryById(ctx, blogDto.CategoryId)
	if err != nil {
		return nil, "", err
	}
	blogDto.Category = dto.CategoryDto{
		CategoryId:   category.CategoryId,
		CategoryName: category.CategoryName,
	}

	// 获取预签名 URL，用于读取 OSS 中文章内容
	presignUrl, err := storage.Storage.GenPreSignUrl(
		ctx,
		ossstore.GenOssSavePath(blogDto.BlogTitle, ossstore.MarkDown),
		ossstore.MarkDown,
		ossstore.Get,
		1*time.Minute,
	)
	if err != nil {
		return nil, "", err
	}

	// 返回包含完整信息的博客DTO对象。
	return blogDto, presignUrl.URL, nil
}

// GetAllImgs 获取所有图片的基本信息，并为每张图片生成预签名的访问链接。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//
// 返回值:
//   - []dto.ImgDto: 包含所有图片信息的切片，每张图片的URL字段已更新为预签名链接。
//   - error: 如果在获取图片信息、生成预签名链接或缓存操作中发生错误，则返回相应的错误信息。
func GetAllImgs(ctx context.Context) ([]dto.ImgDto, error) {
	// 从存储库中获取所有图片的基本信息。
	imgs, err := imgrepo.FindAllImgs(ctx)
	if err != nil {
		return nil, err
	}

	// 遍历每张图片，为其生成预签名的访问链接并更新图片的URL字段。
	for _, img := range imgs {
		// 检查缓存中是否已存在该图片的预签名链接，如果不存在则将其写入缓存。
		cacheKey := storage.BuildImgCacheKey(img.ImgId)
		_, err = storage.Storage.Cache.GetString(ctx, cacheKey)
		if errors.Is(err, cache.ErrNotFound) {
			// 根据图片名称和类型生成OSS存储路径。
			path := ossstore.GenOssSavePath(img.ImgName, img.ImgType)

			// 为图片生成预签名的访问链接，有效期为30分钟。
			presign, err := storage.Storage.GenPreSignUrl(
				ctx,
				path,
				img.ImgType,
				ossstore.Get,
				35*time.Minute,
			)
			if err != nil {
				// 如果生成预签名链接失败，记录错误日志并返回错误。
				msg := fmt.Sprintf("获取图片链接失败: %v", err)
				logger.Error(msg)
				return nil, err
			}

			err = storage.Storage.Cache.SetWithExpired(ctx, cacheKey, presign.URL, 30*time.Minute)
			if err != nil {
				// 如果缓存写入失败，记录错误日志并返回错误。
				msg := fmt.Sprintf("缓存图片链接失败: %v", err)
				logger.Error(msg)
				return nil, err
			}
		}
	}

	// 返回包含预签名链接的图片信息切片。
	return imgs, nil
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
		err := categoryrepo.AddCategory(tx, &categoryDto)
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
			newTags, err = tagrepo.AddTags(tx, tagsWithoutId)
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
		if err := blogrepo.AddBlog(tx, blogDto); err != nil {
			tx.Rollback()
			return err
		}

		// 建立标签与博客的关联关系
		if err := tagrepo.AddBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := blogrepo.UpdateBlog(tx, blogDto); err != nil {
			tx.Rollback()
			return err
		}

		// 更新标签与博客的关联关系
		if err := tagrepo.UpdateBlogTagAssociation(tx, blogDto.BlogId, blogDto.Tags); err != nil {
			tx.Rollback()
			return err
		}
	}
	logger.Info("完成博客的更新或创建操作")

	// 提交事务
	tx.Commit()

	return nil
}

// DeleteBlogById 删除指定ID的博客
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递元数据。
//   - id: 要删除的博客的唯一标识符。
//
// 返回值:
//   - error: 如果删除过程中发生错误，则返回错误信息；否则返回 nil。
func DeleteBlogById(ctx context.Context, id string) error {
	// 获取博客标题
	blogTitle, err := blogrepo.FindBlogTitleById(ctx, id)
	if err != nil {
		return err
	}

	// 开启删除博客事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除博客失败: %v", r)
			tx.Rollback()
		}
	}()

	// 调用仓库方法根据ID删除博客。
	err = blogrepo.DeleteBlogById(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除博客标签关联数据
	err = tagrepo.DeleteBlogTagAssociationByBlogId(tx, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 删除博客对应的 Markdown 文件
	err = storage.Storage.DeleteObject(ctx, ossstore.GenOssSavePath(blogTitle, ossstore.MarkDown))
	if err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	tx.Commit()
	logger.Info("删除博客数据成功")

	// 清理无用标签和分类
	cleanUpTx := storage.Storage.Db.WithContext(ctx).Begin()
	if err = tagrepo.CleanTagsWithoutBlog(cleanUpTx); err != nil {
		cleanUpTx.Rollback()
		return err
	}

	if err = categoryrepo.CleanCategoriesWithoutBlog(cleanUpTx); err != nil {
		cleanUpTx.Rollback()
		return err
	}
	cleanUpTx.Commit()

	return nil
}

func SetTop(ctx context.Context, id string) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("修改置顶状态失败: %v", r)
			tx.Rollback()
		}
	}()

	if err := blogrepo.SetTopById(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func ChangeBlogState(ctx context.Context, id string) error {
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("修改博客状态失败: %v", r)
			tx.Rollback()
		}
	}()

	if err := blogrepo.ChangeBlogStateById(tx, id); err != nil {
		return err
	}

	tx.Commit()

	return nil
}

// IsExistImgByName 检查指定的图片是否存在于数据库和OSS存储中。
// 如果图片在数据库中不存在，或者在OSS中不存在且数据库中存在，则返回true。
// 如果在数据库中找到图片但OSS中不存在，则会尝试删除数据库中的记录，并返回true。
// 如果发生错误，则返回true和错误信息。
// 参数:
//   - ctx: 上下文，用于传递请求范围的信息。
//   - imgName: 图片的名称。
//
// 返回值:
//   - bool: 图片是否存在。
//   - error: 错误信息，如果有的话。
func IsExistImgByName(ctx context.Context, imgName string) (bool, error) {
	// 通过图片名称从数据库中查找图片信息。
	imgDto, err := imgrepo.FindImgByName(ctx, imgName)
	if err != nil {
		// 如果查找过程中出现错误，返回true和错误信息。
		return false, nil
	}

	// 尝试从OSS中获取图片内容。
	flag, err := storage.Storage.IsExist(ctx, ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType))
	// 如果OSS中图片不存在，但数据库中存在记录，则需要删除数据库中的记录。
	if err != nil {
		return false, err
	}

	if flag {
		return true, nil
	} else {
		return false, nil
	}
}

// RenameImgById 根据图片 ID 修改图片名称，包括 OSS 中的文件名和数据库中的记录。
// 参数:
//   - ctx: 上下文对象，用于控制请求生命周期和传递上下文信息。
//   - imgId: 图片的唯一标识符，用于查询和更新图片信息。
//   - newName: 新的图片名称，用于替换旧的图片名称。
//
// 返回值:
//   - error: 如果操作失败，返回错误信息；如果成功，返回 nil。
func RenameImgById(ctx context.Context, imgId string, newName string) error {
	// 根据图片 ID 查询图片信息，确保图片存在并获取其详细信息
	imgDto, err := imgrepo.FindImgById(ctx, imgId)
	if err != nil {
		return err
	}

	logger.Info("重命名 OSS 中的图片名称")
	// 生成 OSS 中的旧路径和新路径，并调用存储服务重命名 OSS 中的文件
	oldPath := ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType)
	newPath := ossstore.GenOssSavePath(newName, imgDto.ImgType)
	if renameErr := storage.Storage.RenameObject(ctx, oldPath, newPath); renameErr != nil {
		return renameErr
	}
	logger.Info("重命名 OSS 中的图片名称成功")

	logger.Info("更新数据库中的图片名称")
	// 开启数据库事务，更新数据库中图片名称，并根据更新结果提交或回滚事务
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("重命名图片失败: %v", r)
			tx.Rollback()
		}
	}()

	if err = imgrepo.UpdateImgNameById(tx, imgId, newName); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	logger.Info("更新数据库中的图片名称成功")

	logger.Info("删除缓存中保存的预签名 URL")
	if delErr := storage.Storage.Cache.Delete(ctx, storage.BuildImgCacheKey(imgId)); delErr != nil {
		return delErr
	}
	logger.Info("删除缓存中保存的预签名 URL 成功")

	// 生成新的预签名 URL
	presign, err := storage.Storage.GenPreSignUrl(
		ctx,
		ossstore.GenOssSavePath(newName, imgDto.ImgType),
		imgDto.ImgType,
		ossstore.Get,
		35*time.Minute,
	)
	if err != nil {
		return err
	}
	// 缓存新的预签名 URL
	err = storage.Storage.Cache.SetWithExpired(ctx, storage.BuildImgCacheKey(imgId), presign.URL, 30*time.Minute)
	if err != nil {
		logger.Warn("缓存新的预签名 URL 失败")
		return err
	}
	logger.Info("缓存新的预签名 URL 成功")
	logger.Info("完成图片的更新操作")

	return nil
}

// DeleteImg 删除指定 ID 的图片信息及其相关资源。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - id: 图片的唯一标识符，用于定位需要删除的图片。
//
// 返回值:
//   - error: 如果在查找图片信息、删除 OSS 中的图片数据或删除数据库记录时发生错误，则返回相应的错误信息；否则返回 nil。
func DeleteImg(ctx context.Context, id string) error {
	// 查找图片信息，确保图片存在并获取其详细信息
	imgDto, err := imgrepo.FindImgById(ctx, id)
	if err != nil {
		return err
	}

	logger.Info("删除 OSS 中存储的图片文件")
	// 使用图片名称和类型生成存储路径
	if err := storage.Storage.DeleteObject(ctx, ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType)); err != nil {
		return err
	}
	logger.Info("删除 OSS 中存储的图片文件成功")

	logger.Info("删除数据库中与图片相关的记录")
	// 开启数据库事务，删除数据库中与图片相关的记录
	tx := storage.Storage.Db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("删除图片失败: %v", r)
			tx.Rollback()
		}
	}()

	if err := imgrepo.DeleteImgById(tx, id); err != nil {
		return err
	}
	tx.Commit()
	logger.Info("删除数据库中与图片相关的记录成功")

	logger.Info("删除缓存中保存的预签名 URL")
	// 删除缓存中保存的预签名 URL
	if err := storage.Storage.Cache.Delete(ctx, storage.BuildImgCacheKey(id)); err != nil {
		return err
	}
	logger.Info("删除缓存中保存的预签名 URL 成功")

	return nil
}

func UpdateConfig() error {
	projConfig := config.ProjectConfig{
		User:   config.User,
		Server: config.Server,
		MySQL:  config.MySQL,
		Oss:    config.Oss,
		Cache:  config.Cache,
		Logger: config.Logger,
	}

	err := projConfig.Store()
	if err != nil {
		return err
	}

	return nil
}
