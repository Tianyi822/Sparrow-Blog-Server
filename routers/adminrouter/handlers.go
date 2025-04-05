package adminrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/email"
	"h2blog_server/internal/model/vo"
	"h2blog_server/internal/services/adminservice"
	"h2blog_server/internal/services/imgservice"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"strings"
	"time"
)

// sendVerificationCode 处理发送验证码的请求。
// 参数:
//   - *gin.Context: HTTP 请求上下文，包含请求数据和响应方法。
//
// 功能描述:
//
//	该函数从请求中解析用户提交的数据，验证用户邮箱是否正确，
//	并调用邮件服务发送验证码。根据操作结果返回相应的 HTTP 响应。
func sendVerificationCode(ctx *gin.Context) {
	// 从请求中解析原始数据为 map，并处理可能的解析错误。
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "登录信息解析错误", err.Error())
		return
	}

	// 验证用户提交的邮箱是否与配置中的用户邮箱一致。
	if rawData["user_email"].(string) != config.User.UserEmail {
		resp.BadRequest(ctx, "用户邮箱错误", "")
		return
	}

	// 调用邮件服务发送验证码邮件，并处理发送过程中可能出现的错误。
	err = email.SendVerificationCodeEmail(ctx, config.User.UserEmail)
	if err != nil {
		resp.Err(ctx, "验证码发送失败", err.Error())
		return
	}

	// 如果验证码发送成功，返回成功的 HTTP 响应。
	resp.Ok(ctx, "验证码发送成功", nil)
}

// login 处理用户登录请求。
// 参数:
//   - Gin 上下文，用于处理 HTTP 请求和响应。
func login(ctx *gin.Context) {
	// 从请求中解析原始数据为 Map 格式
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "登录信息解析错误", err.Error())
		return
	}

	token, err := adminservice.Login(ctx, rawData["user_email"].(string), rawData["verification_code"].(string))
	if err != nil {
		resp.Err(ctx, "登录失败", err.Error())
		return
	}

	resp.Ok(ctx, "登录成功", map[string]string{
		"token": token,
	})
}

func genPresignPutUrl(ctx *gin.Context) {
	fileName := ctx.Param("file_name")
	fileType := ctx.Param("file_type")

	var path string
	switch strings.ToLower(fileType) {
	case ossstore.MarkDown:
		fileType = ossstore.MarkDown
		path = ossstore.GenOssSavePath(fileName, ossstore.MarkDown)
	case ossstore.Webp:
		fileType = ossstore.Webp
		path = ossstore.GenOssSavePath(fileName, ossstore.Webp)
	default:
		resp.BadRequest(ctx, "文件类型错误", nil)
		return
	}

	presign, err := storage.Storage.GenPreSignUrl(ctx, path, fileType, ossstore.Put, 2*time.Minute)
	if err != nil {
		resp.Err(ctx, "获取预签名URL失败", err.Error())
		return
	}

	resp.Ok(ctx, "获取成功", map[string]string{
		"pre_sign_put_url": presign.URL,
	})
}

// getAllBlogs 是一个处理函数，用于获取所有博客数据并将其转换为视图对象（VO）格式后返回。
// 参数:
//   - ctx: Gin 框架的上下文对象，用于处理 HTTP 请求和响应。
func getAllBlogs(ctx *gin.Context) {
	// 调用 adminservice.GetBlogsToAdminPosts 获取博客数据的 DTO 列表。
	// 如果发生错误，则返回错误响应。
	blogDtos, err := adminservice.GetBlogsToAdminPosts(ctx)
	if err != nil {
		resp.Err(ctx, "获取博客失败", err.Error())
		return
	}

	// 将 DTO 列表转换为 VO 列表，以便前端使用。
	blogVos := make([]vo.BlogVo, 0, len(blogDtos))
	for _, blogDto := range blogDtos {
		// 构造分类信息的 VO 对象。
		category := vo.CategoryVo{
			CategoryId:   blogDto.Category.CategoryId,
			CategoryName: blogDto.Category.CategoryName,
		}

		// 构造标签信息的 VO 列表。
		tags := make([]vo.TagVo, 0, len(blogDto.Tags))
		for _, tag := range blogDto.Tags {
			tags = append(tags, vo.TagVo{
				TagId:   tag.TagId,
				TagName: tag.TagName,
			})
		}

		// 构造博客信息的 VO 对象，并将其添加到结果列表中。
		blogVo := vo.BlogVo{
			BlogId:       blogDto.BlogId,
			BlogTitle:    blogDto.BlogTitle,
			Category:     category,
			Tags:         tags,
			BlogState:    blogDto.BlogState,
			BlogWordsNum: blogDto.BlogWordsNum,
			BlogIsTop:    blogDto.BlogIsTop,
			CreateTime:   blogDto.CreateTime,
			UpdateTime:   blogDto.UpdateTime,
		}
		blogVos = append(blogVos, blogVo)
	}

	// 返回成功响应，包含转换后的博客 VO 列表。
	resp.Ok(ctx, "获取博客成功", blogVos)
}

func deleteBlog(ctx *gin.Context) {
	if err := adminservice.DeleteBlogById(ctx, ctx.Param("blog_id")); err != nil {
		resp.Err(ctx, "删除博客失败", err.Error())
		return
	}

	resp.Ok(ctx, "删除成功", nil)
}

func changeBlogState(ctx *gin.Context) {
	if err := adminservice.ChangeBlogState(ctx, ctx.Param("blog_id")); err != nil {
		resp.Err(ctx, "修改博客状态失败", err.Error())
		return
	}

	resp.Ok(ctx, "修改博客状态成功", nil)
}

func setTop(ctx *gin.Context) {
	if err := adminservice.SetTop(ctx, ctx.Param("blog_id")); err != nil {
		resp.Err(ctx, "修改置顶失败", err.Error())
		return
	}

	resp.Ok(ctx, "已修改是否置顶", nil)
}

// getAllTagsCategories 获取所有的分类和标签信息，并以结构化的方式返回给客户端。
// 参数:
//   - ctx *gin.Context: Gin框架的上下文对象，用于处理HTTP请求和响应。
func getAllTagsCategories(ctx *gin.Context) {
	// 调用adminService的GetAllCategoriesAndTags方法获取分类和标签数据。
	// 如果发生错误，则返回错误信息。
	categories, tags, err := adminservice.GetAllCategoriesAndTags(ctx)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	// 将分类数据转换为CategoryVo视图对象列表。
	categoryVos := make([]vo.Vo, 0, len(categories))
	for _, category := range categories {
		categoryVos = append(categoryVos, &vo.CategoryVo{
			CategoryId:   category.CategoryId,
			CategoryName: category.CategoryName,
		})
	}

	// 将标签数据转换为TagVo视图对象列表。
	tagVos := make([]vo.Vo, 0, len(tags))
	for _, tag := range tags {
		tagVos = append(tagVos, &vo.TagVo{
			TagId:   tag.TagId,
			TagName: tag.TagName,
		})
	}

	// 将分类和标签的视图对象列表封装为响应数据，并返回成功信息。
	resp.Ok(ctx, "获取成功", map[string][]vo.Vo{
		"categories": categoryVos,
		"tags":       tagVos,
	})
}

// updateOrAddBlog 处理博客更新或添加请求
// 参数:
//   - ctx *gin.Context: 框架的上下文对象
func updateOrAddBlog(ctx *gin.Context) {
	// 从请求中解析博客数据传输对象 (DTO)，如果解析失败则返回错误响应。
	blogDto, err := tools.GetBlogDto(ctx)
	if err != nil {
		resp.BadRequest(ctx, "请求数据有误，请检查错误", err.Error())
		return
	}

	if blogDto.BlogTitle == "" {
		resp.BadRequest(ctx, "博客标题不能为空", "")
		return
	}

	if blogDto.Category.CategoryName == "" {
		resp.BadRequest(ctx, "博客分类不能为空", "")
		return
	}

	if blogDto.BlogWordsNum == 0 {
		resp.BadRequest(ctx, "博客不能为空", "")
		return
	}

	// 调用服务层方法更新或添加博客，如果操作失败则返回错误响应。
	err = adminservice.UpdateOrAddBlog(ctx, blogDto)
	if err != nil {
		resp.Err(ctx, "添加或更新失败", err.Error())
		return
	}

	// 如果操作成功，返回成功的HTTP响应。
	resp.Ok(ctx, "操作成功", map[string]string{
		"blog_id": blogDto.BlogId,
	})
}

func getBlogData(ctx *gin.Context) {
	blogId := ctx.Param("blog_id")
	blogDto, url, err := adminservice.GetBlogData(ctx, blogId)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	tagVos := make([]vo.TagVo, 0, len(blogDto.Tags))
	for _, tag := range blogDto.Tags {
		tagVos = append(tagVos, vo.TagVo{
			TagId:   tag.TagId,
			TagName: tag.TagName,
		})
	}

	blogVo := vo.BlogVo{
		BlogId:    blogDto.BlogId,
		BlogTitle: blogDto.BlogTitle,
		BlogBrief: blogDto.BlogBrief,
		Category: vo.CategoryVo{
			CategoryId:   blogDto.Category.CategoryId,
			CategoryName: blogDto.Category.CategoryName,
		},
		Tags:         tagVos,
		BlogState:    blogDto.BlogState,
		BlogWordsNum: blogDto.BlogWordsNum,
	}

	resp.Ok(ctx, "获取成功", map[string]any{
		"blog_data":   blogVo,
		"content_url": url,
	})
}

func addImgs(ctx *gin.Context) {
	imgsDto, err := tools.GetImgDtos(ctx)
	if err != nil {
		return
	}

	if err := imgservice.AddImgs(ctx, imgsDto.Imgs); err != nil {
		resp.Err(ctx, "添加失败", err.Error())
		return
	}

	resp.Ok(ctx, "添加成功", nil)
}

func getAllImgs(ctx *gin.Context) {
	imgDtos, err := adminservice.GetAllImgs(ctx)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	imgVos := make([]vo.ImgVo, 0, len(imgDtos))
	for _, imgDto := range imgDtos {
		imgVos = append(imgVos, vo.ImgVo{
			ImgId:      imgDto.ImgId,
			ImgName:    imgDto.ImgName,
			ImgType:    imgDto.ImgType,
			CreateTime: imgDto.CreateTime,
		})
	}

	resp.Ok(ctx, "获取成功", imgVos)
}

func deleteImg(ctx *gin.Context) {
	if err := adminservice.DeleteImg(ctx, ctx.Param("img_id")); err != nil {
		resp.Err(ctx, "删除失败", err.Error())
		return
	}

	resp.Ok(ctx, "删除成功", nil)
}

// renameImg 修改指定图片的名称。
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法。
//
// 功能描述:
//  1. 从请求上下文中解析图片数据传输对象 (ImgDto)。
//  2. 验证请求路径中的图片ID与解析出的图片ID是否一致。
//  3. 调用 adminservice.RenameImgById 方法修改图片名称。
//  4. 根据操作结果返回成功或失败的响应。
func renameImg(ctx *gin.Context) {
	// 从请求上下文中获取图片数据传输对象 (ImgDto)
	imgDto, err := tools.GetImgDto(ctx)
	if err != nil {
		return
	}

	if imgDto.ImgName == "" {
		resp.BadRequest(ctx, "图片名称不能为空", nil)
		return
	}

	// 验证请求路径中的图片ID与 ImgDto 中的图片ID是否匹配
	imgId := ctx.Param("img_id")
	if imgId != imgDto.ImgId {
		resp.BadRequest(ctx, "图片ID不匹配", nil)
		return
	}

	// 调用服务层方法修改图片名称，并处理可能的错误
	if err := adminservice.RenameImgById(ctx, imgDto.ImgId, imgDto.ImgName); err != nil {
		resp.Err(ctx, "修改失败", err.Error())
		return
	}

	// 返回操作成功的响应
	resp.Ok(ctx, "修改成功", nil)
}

func isExist(ctx *gin.Context) {
	flag, err := adminservice.IsExistImg(ctx, ctx.Param("img_name"))
	if err != nil {
		resp.Err(ctx, "查询失败", err.Error())
		return
	}

	if flag {
		resp.Ok(ctx, "图片存在", flag)
	} else {
		resp.Ok(ctx, "图片不存在", flag)
	}
}
