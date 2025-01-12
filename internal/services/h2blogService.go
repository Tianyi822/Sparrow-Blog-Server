package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"h2blog/internal/model/dto"
	"h2blog/internal/model/po"
	"h2blog/internal/model/vo"
	"h2blog/pkg/logger"
	"h2blog/pkg/markdown"
	"h2blog/pkg/utils"
	"h2blog/storage"
)

// genBlogId 用于生成博客的唯一标识符
//   - title 是博客的标题
//
// 返回值
//   - string 表示生成的博客ID
func genBlogId(title string) string {
	// 使用envs包的HashWithLength函数生成一个长度为16的哈希字符串作为博客ID
	str, err := utils.HashWithLength(title, 16)
	// 检查是否生成成功，如果失败则记录错误并尝试重新生成
	if err != nil {
		// 使用logger包记录错误信息，包括错误详情
		logger.Error("生成博客 ID 失败: %v，准备重新生成", err)
		// 初始化计数器，用于限制重试次数
		count := 0
		title = title + fmt.Sprintf("%d", count)
		// 使用for循环尝试重新生成博客ID，最多重试3次
		for count <= 2 && err != nil {
			str, err = utils.HashWithLength(title, 16)
			count++
		}
	}
	logger.Info("生成博客 ID 成功: %s", str)
	// 返回生成的博客ID
	return str
}

// GetH2BlogInfoById 用于获取指定博客信息
//   - ctx 是 Gin 框架的上下文对象，用于处理 HTTP 请求和响应
//   - blogId 是要获取的博客的唯一标识符
//
// 返回值
//   - *dto.BlogInfoDto 表示获取到的博客信息
//   - error 表示获取过程中可能发生的错误
func GetH2BlogInfoById(ctx *gin.Context, blogId string) (*vo.BlogInfoVo, error) {
	// 创建一个 BlogInfo 类型的实例 blogInfoPo，用于存储博客信息的数据模型
	blogInfoPo := po.BlogInfo{
		BlogId: blogId,
	}

	// 先查询是否有该数据
	num, err := blogInfoPo.FindOneById(ctx)
	// 检查是否有错误发生
	if err != nil {
		// 如果有错误，直接返回当前num值和错误信息
		return nil, err
	} else {
		// 如果没有错误，进一步处理
		if num > 0 {
			// 如果num大于0，表示找到了该数据，可以直接返回
			return &vo.BlogInfoVo{
				BlogId: blogInfoPo.BlogId,
				Title:  blogInfoPo.Title,
				Brief:  blogInfoPo.Brief,
			}, nil
		} else {
			// 如果num不大于0，表示没有找到要获取的记录
			return nil, fmt.Errorf("没有找到要获取的博客")
		}
	}
}

// AddH2BlogInfo 用于添加新的博客信息
//   - ctx 是 Gin 框架的上下文对象，用于处理 HTTP 请求和响应
//   - blogInfoDto 是包含博客信息的 BlogInfoDto 对象
//
// 返回值
//   - int64 表示添加操作影响的行数
//   - error 表示添加过程中可能发生的错误
func AddH2BlogInfo(ctx *gin.Context, blogInfoDto *dto.BlogInfoDto) (int64, error) {
	// 1. 从 dto 中获取到博客信息
	// 2. 加载博客的原始 markdown 文件成字符串
	mdStr, err := storage.Storage.GetContentFromOss(ctx, utils.GenOssSavePath(blogInfoDto.Name(), utils.MarkDown))
	if err != nil {
		return 0, err
	}

	// 3. 将 markdown 文件内容解析成 html
	htmlStr, err := markdown.Parse(mdStr)
	if err != nil {
		return 0, err
	}

	// 4. 将 html 内容保存到 oss 中，保存路径与 md 文件同目录
	savePath := utils.GenOssSavePath(blogInfoDto.Name(), utils.HTML)
	err = storage.Storage.PutContentToOss(ctx, htmlStr, savePath)
	if err != nil {
		return 0, err
	}

	// 5. 将博客信息保存到数据库中
	blogInfoPo := po.BlogInfo{
		BlogId: genBlogId(blogInfoDto.Title),
		Title:  blogInfoDto.Title,
		Brief:  blogInfoDto.Brief,
	}

	logger.Info("准备保存博客信息")
	// 先查询是否有该数据
	num, err := blogInfoPo.FindOneById(ctx)
	// 根据错误情况决定是添加一条新的博客信息还是更新已有的博客信息
	if err != nil {
		num, err = blogInfoPo.AddOne(ctx)
	} else {
		num, err = blogInfoPo.UpdateOneById(ctx)
	}
	logger.Info("保存博客信息成功: %v", num)

	return num, err
}

// ModifyH2BlogInfo 用于修改指定博客信息
//   - ctx 是 Gin 框架的上下文对象，用于处理 HTTP 请求和响应
//   - dto 是包含修改信息的 BlogInfoDto 对象
//
// 返回值
//   - int64 表示修改操作影响的行数
//   - error 表示修改过程中可能发生的错误
func ModifyH2BlogInfo(ctx *gin.Context, dto *dto.BlogInfoDto) (int64, error) {
	blogInfoPo := po.BlogInfo{
		BlogId: dto.BlogId,
	}

	// TODO: 后续会添加标签、分类等信息的修改

	logger.Info("准备修改博客信息")

	// 先查询是否有该数据
	num, err := blogInfoPo.FindOneById(ctx)
	// 根据错误情况决定是添加一条新的博客信息还是更新已有的博客信息
	if err != nil {
		return 0, err
	} else {
		if num > 0 {
			// 1. 重新解析 Markdown 文件内容并保存到 Oss 中
			// 从OSS存储中获取新Markdown文件内容
			mdStr, err := storage.Storage.GetContentFromOss(ctx, utils.GenOssSavePath(dto.Name(), utils.MarkDown))
			if err != nil {
				return 0, err
			}
			// 将Markdown内容解析为新HTML
			htmlStr, err := markdown.Parse(mdStr)
			if err != nil {
				return 0, err
			}
			// 将生成的HTML内容上传到新的HTML文件路径
			err = storage.Storage.PutContentToOss(ctx, htmlStr, utils.GenOssSavePath(dto.Name(), utils.HTML))
			if err != nil {
				return 0, err
			}

			// 2. 删除旧的 HTML 文件
			err = storage.Storage.DeleteObject(ctx, utils.GenOssSavePath(blogInfoPo.Title, utils.HTML))
			if err != nil {
				return 0, err
			}

			// 3. 修改数据库中的数据
			blogInfoPo.Title = dto.Title
			blogInfoPo.Brief = dto.Brief

			num, err = blogInfoPo.UpdateOneById(ctx)
		} else {
			return num, fmt.Errorf("没有找到要修改的博客")
		}
	}
	logger.Info("修改博客信息成功: %v", num)

	return num, err
}

// DeleteH2BlogInfo 用于删除指定博客信息
//   - ctx 是 Gin 框架的上下文对象，用于处理 HTTP 请求和响应
//   - blogId 是要删除的博客的唯一标识符
//
// 返回值
//   - int64 表示删除操作影响的行数
//   - error 表示删除过程中可能发生的错误
func DeleteH2BlogInfo(ctx *gin.Context, dto *dto.BlogInfoDto) (int64, error) {
	// 创建一个 BlogInfo 类型的实例 blogInfoPo，用于存储博客信息的数据模型
	blogInfoPo := po.BlogInfo{
		BlogId: dto.BlogId,
	}

	logger.Info("准备删除博客信息")
	// 先查询是否有该数据
	num, err := blogInfoPo.FindOneById(ctx)
	// 检查是否有错误发生
	if err != nil {
		// 如果有错误，直接返回当前num值和错误信息
		return num, err
	} else {
		// 如果没有错误，进一步处理
		if num > 0 {
			// 如果num大于0，表示需要删除一条记录
			// 1. 先删除 Oss 中的 HTML 文件和 Markdown 文件
			err = storage.Storage.DeleteObject(ctx, utils.GenOssSavePath(blogInfoPo.Title, utils.HTML))
			if err != nil {
				return 0, err
			}
			err = storage.Storage.DeleteObject(ctx, utils.GenOssSavePath(blogInfoPo.Title, utils.MarkDown))
			if err != nil {
				return 0, err
			}
			// 2. 删除数据库中的记录
			num, err = blogInfoPo.DeleteOneById(ctx)
		} else {
			// 如果num不大于0，表示没有找到要删除的记录
			return num, fmt.Errorf("没有找到要删除的博客")
		}
	}
	logger.Info("删除博客信息成功: %v", num)

	return num, err
}
