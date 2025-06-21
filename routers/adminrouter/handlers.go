package adminrouter

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/internal/model/vo"
	"sparrow_blog_server/internal/services/adminservices"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/email"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/routers/resp"
	"sparrow_blog_server/routers/tools"
	"sparrow_blog_server/searchengine"
	"sparrow_blog_server/storage"
	"sparrow_blog_server/storage/ossstore"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// getUserInfo 获取用户信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func getUserInfo(ctx *gin.Context) {
	resp.Ok(ctx, "获取用户信息成功", map[string]string{
		"user_name": config.User.Username,
	})
}

// sendLoginVerificationCode 发送登录验证码
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求数据和响应方法
//
// 功能描述:
//
//	该函数从请求中解析用户提交的数据，验证用户邮箱是否正确，
//	并调用邮件服务发送验证码。根据操作结果返回相应的 HTTP 响应。
func sendLoginVerificationCode(ctx *gin.Context) {
	// 从请求中解析原始数据为 map，并处理可能的解析错误
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "登录信息解析错误", err.Error())
		return
	}

	// 验证用户提交的邮箱是否与配置中的用户邮箱一致
	if rawData["user_email"].(string) != config.User.UserEmail {
		resp.BadRequest(ctx, "用户邮箱错误", "")
		return
	}

	// 调用邮件服务发送验证码邮件，并处理发送过程中可能出现的错误
	err = email.SendVerificationCodeBySys(ctx)
	if err != nil {
		resp.Err(ctx, "验证码发送失败", err.Error())
		return
	}

	// 如果验证码发送成功，返回成功的 HTTP 响应
	resp.Ok(ctx, "验证码发送成功", nil)
}

// login 处理用户登录请求
// 参数:
//   - ctx *gin.Context: Gin 上下文，用于处理 HTTP 请求和响应
func login(ctx *gin.Context) {
	// 从请求中解析原始数据为 Map 格式
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}

	userEmail, getErr := tools.GetStringFromRawData(rawData, "user_email")
	if getErr != nil {
		resp.BadRequest(ctx, "用户邮箱解析错误", getErr.Error())
		return
	}

	verificationCode, getErr := tools.GetStringFromRawData(rawData, "verification_code")
	if getErr != nil {
		resp.BadRequest(ctx, "验证码解析错误", getErr.Error())
		return
	}

	token, err := adminservices.Login(ctx, userEmail, verificationCode)
	if err != nil {
		resp.Err(ctx, "登录失败", err.Error())
		return
	}

	resp.Ok(ctx, "登录成功", map[string]string{
		"token": token,
	})
}

// logout 处理用户登出请求
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func logout(ctx *gin.Context) {
	token := ctx.GetString("token")
	err := adminservices.Logout(ctx, token)
	if err != nil {
		resp.Err(ctx, "退出登录失败", err.Error())
		return
	}
	resp.Ok(ctx, "退出登录成功", nil)
}

// genPresignPutUrl 生成预签名的文件上传URL
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求参数中获取文件名和文件类型
//  2. 根据文件类型生成对应的OSS存储路径
//  3. 生成预签名的上传URL
//  4. 返回预签名URL给客户端
func genPresignPutUrl(ctx *gin.Context) {
	// 从请求参数中获取文件名和文件类型
	fileName := ctx.Param("file_name")
	fileType := ctx.Param("file_type")

	// 根据文件类型生成存储路径
	var path string
	switch strings.ToLower(fileType) {
	case ossstore.MarkDown:
		// Markdown文件处理
		fileType = ossstore.MarkDown
		path = ossstore.GenOssSavePath(fileName, ossstore.MarkDown)
	case ossstore.Webp:
		// Webp图片处理
		fileType = ossstore.Webp
		path = ossstore.GenOssSavePath(fileName, ossstore.Webp)
	default:
		// 不支持的文件类型返回错误
		resp.BadRequest(ctx, "文件类型错误", nil)
		return
	}

	// 生成预签名URL，有效期2分钟
	presign, err := storage.Storage.GenPreSignUrl(ctx, path, fileType, ossstore.Put, 2*time.Minute)
	if err != nil {
		resp.Err(ctx, "获取预签名URL失败", err.Error())
		return
	}

	// 返回预签名URL给客户端
	resp.Ok(ctx, "获取成功", map[string]string{
		"pre_sign_put_url": presign.URL,
	})
}

// getAllBlogs 获取所有博客数据并转换为视图对象格式返回
// 参数:
//   - ctx *gin.Context: Gin 框架的上下文对象，用于处理 HTTP 请求和响应
func getAllBlogs(ctx *gin.Context) {
	// 调用 adminservices.GetBlogsToAdminPosts 获取博客数据的 DTO 列表
	// 如果发生错误，则返回错误响应
	blogDtos, err := adminservices.GetBlogsToAdminPosts(ctx)
	if err != nil {
		resp.Err(ctx, "获取博客失败", err.Error())
		return
	}

	// 将 DTO 列表转换为 VO 列表，以便前端使用
	blogVos := make([]vo.BlogVo, 0, len(blogDtos))
	for _, blogDto := range blogDtos {
		// 构造分类信息的 VO 对象
		category := vo.CategoryVo{
			CategoryId:   blogDto.Category.CategoryId,
			CategoryName: blogDto.Category.CategoryName,
		}

		// 构造标签信息的 VO 列表
		tags := make([]vo.TagVo, 0, len(blogDto.Tags))
		for _, tag := range blogDto.Tags {
			tags = append(tags, vo.TagVo{
				TagId:   tag.TagId,
				TagName: tag.TagName,
			})
		}

		// 构造博客信息的 VO 对象，并将其添加到结果列表中
		blogVo := vo.BlogVo{
			BlogId:       blogDto.BlogId,
			BlogTitle:    blogDto.BlogTitle,
			BlogImageId:  blogDto.BlogImageId,
			Category:     &category,
			Tags:         tags,
			BlogState:    blogDto.BlogState,
			BlogWordsNum: blogDto.BlogWordsNum,
			BlogIsTop:    blogDto.BlogIsTop,
			CreateTime:   blogDto.CreateTime,
			UpdateTime:   blogDto.UpdateTime,
		}
		blogVos = append(blogVos, blogVo)
	}

	// 返回成功响应，包含转换后的博客 VO 列表
	resp.Ok(ctx, "获取博客成功", blogVos)
}

// deleteBlog 删除指定的博客
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func deleteBlog(ctx *gin.Context) {
	if err := adminservices.DeleteBlogById(ctx, ctx.Param("blog_id")); err != nil {
		resp.Err(ctx, "删除博客失败", err.Error())
		return
	}

	resp.Ok(ctx, "删除成功", nil)
}

// changeBlogState 修改博客状态
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func changeBlogState(ctx *gin.Context) {
	if err := adminservices.ChangeBlogState(ctx, ctx.Param("blog_id")); err != nil {
		resp.Err(ctx, "修改博客状态失败", err.Error())
		return
	}

	resp.Ok(ctx, "修改博客状态成功", nil)
}

// setTop 设置博客置顶状态
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func setTop(ctx *gin.Context) {
	if err := adminservices.SetTop(ctx, ctx.Param("blog_id")); err != nil {
		resp.Err(ctx, "修改置顶失败", err.Error())
		return
	}

	resp.Ok(ctx, "已修改是否置顶", nil)
}

// getAllTagsCategories 获取所有的分类和标签信息，并以结构化的方式返回给客户端
// 参数:
//   - ctx *gin.Context: Gin框架的上下文对象，用于处理HTTP请求和响应
func getAllTagsCategories(ctx *gin.Context) {
	// 调用adminService的GetAllCategoriesAndTags方法获取分类和标签数据
	// 如果发生错误，则返回错误信息
	categories, tags, err := adminservices.GetAllCategoriesAndTags(ctx)
	if err != nil {
		resp.Err(ctx, "获取失败", err.Error())
		return
	}

	// 将分类数据转换为CategoryVo视图对象列表
	categoryVos := make([]vo.Vo, 0, len(categories))
	for _, category := range categories {
		categoryVos = append(categoryVos, &vo.CategoryVo{
			CategoryId:   category.CategoryId,
			CategoryName: category.CategoryName,
		})
	}

	// 将标签数据转换为TagVo视图对象列表
	tagVos := make([]vo.Vo, 0, len(tags))
	for _, tag := range tags {
		tagVos = append(tagVos, &vo.TagVo{
			TagId:   tag.TagId,
			TagName: tag.TagName,
		})
	}

	// 将分类和标签的视图对象列表封装为响应数据，并返回成功信息
	resp.Ok(ctx, "获取成功", map[string][]vo.Vo{
		"categories": categoryVos,
		"tags":       tagVos,
	})
}

// updateOrAddBlog 处理博客更新或添加请求
// 参数:
//   - ctx *gin.Context: 框架的上下文对象
func updateOrAddBlog(ctx *gin.Context) {
	// 从请求中解析博客数据传输对象 (DTO)，如果解析失败则返回错误响应
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

	if blogDto.BlogImageId == "" {
		resp.BadRequest(ctx, "博客封面不能为空", "")
		return
	}

	if blogDto.BlogWordsNum == 0 {
		resp.BadRequest(ctx, "博客不能为空", "")
		return
	}

	// 调用服务层方法更新或添加博客，如果操作失败则返回错误响应
	err = adminservices.UpdateOrAddBlog(ctx, blogDto)
	if err != nil {
		resp.Err(ctx, "添加或更新失败", err.Error())
		return
	}

	// 如果操作成功，返回成功的HTTP响应
	resp.Ok(ctx, "操作成功", map[string]string{
		"blog_id": blogDto.BlogId,
	})
}

// getBlogData 获取指定博客的详细数据
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func getBlogData(ctx *gin.Context) {
	blogId := ctx.Param("blog_id")
	blogDto, url, err := adminservices.GetBlogData(ctx, blogId)
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
		BlogId:      blogDto.BlogId,
		BlogTitle:   blogDto.BlogTitle,
		BlogImageId: blogDto.BlogImageId,
		BlogBrief:   blogDto.BlogBrief,
		Category: &vo.CategoryVo{
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

// addImgs 添加图片
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func addImgs(ctx *gin.Context) {
	imgsDto, err := tools.GetImgDtos(ctx)
	if err != nil {
		return
	}

	if err := adminservices.AddImgs(ctx, imgsDto.Imgs); err != nil {
		resp.Err(ctx, "添加失败", err.Error())
		return
	}

	resp.Ok(ctx, "添加成功", nil)
}

// getAllImgs 获取所有图片
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func getAllImgs(ctx *gin.Context) {
	imgDtos, err := adminservices.GetAllImgs(ctx)
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

// deleteImg 删除指定图片
// 参数:
//   - ctx *gin.Context: HTTP请求上下文
func deleteImg(ctx *gin.Context) {
	if err := adminservices.DeleteImg(ctx, ctx.Param("img_id")); err != nil {
		resp.Err(ctx, "删除失败", err.Error())
		return
	}

	resp.Ok(ctx, "删除成功", nil)
}

// renameImg 修改指定图片的名称
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求上下文中解析图片数据传输对象 (ImgDto)
//  2. 验证请求路径中的图片ID与解析出的图片ID是否一致
//  3. 调用 adminservices.RenameImgById 方法修改图片名称
//  4. 根据操作结果返回成功或失败的响应
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
	if err := adminservices.RenameImgById(ctx, imgDto.ImgId, imgDto.ImgName); err != nil {
		resp.Err(ctx, "修改失败", err.Error())
		return
	}

	// 返回操作成功的响应
	resp.Ok(ctx, "修改成功", nil)
}

// isExist 检查指定名称的图片是否存在
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求参数中获取图片名称
//  2. 调用服务层方法检查图片是否存在
//  3. 根据检查结果返回相应的响应
//     - 如果图片存在，返回成功响应并设置flag为true
//     - 如果图片不存在，返回成功响应并设置flag为false
//     - 如果查询过程发生错误，返回错误响应
func isExist(ctx *gin.Context) {
	flag, err := adminservices.IsExistImgByName(ctx, ctx.Param("img_name"))
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

// getUserConfig 获取用户基本信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从系统配置中获取用户信息，包括用户名、邮箱、SMTP配置等
//  2. 将获取到的用户信息封装为map结构返回给客户端
//  3. 返回成功的HTTP响应
func getUserConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]any{
		"user_name":           config.User.Username,
		"user_email":          config.User.UserEmail,
		"user_github_address": config.User.UserGithubAddress,
		"user_hobbies":        config.User.UserHobbies,
		"type_writer_content": config.User.TypeWriterContent,
		"background_image":    config.User.BackgroundImage,
		"avatar_image":        config.User.AvatarImage,
		"web_logo":            config.User.WebLogo,
		"icp_filing_number":   config.User.ICPFilingNumber,
	})
}

// verifyNewSmtpConfig 验证新的 SMTP 配置，发送验证邮件
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中获取并验证新的SMTP配置信息，包括邮箱、SMTP账号、地址、授权码和端口
//  2. 使用新的配置信息发送验证邮件，测试配置是否正确
//  3. 根据验证结果返回相应的响应
func verifyNewSmtpConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "请求数据有误，请检查错误", err.Error())
		return
	}

	// 验证并获取SMTP账号
	smtpAccount, getErr := tools.GetStringFromRawData(rawData, "server.smtp_account")
	if getErr != nil {
		resp.BadRequest(ctx, "SMTP账号配置错误", getErr.Error())
		return
	}
	if anaErr := tools.AnalyzeEmail(smtpAccount); anaErr != nil {
		resp.BadRequest(ctx, "SMTP账号配置错误", anaErr.Error())
		return
	}

	// 验证并获取SMTP服务器地址
	smtpAddress, getErr := tools.GetStringFromRawData(rawData, "server.smtp_address")
	if getErr != nil {
		resp.BadRequest(ctx, "SMTP服务器地址配置错误", getErr.Error())
		return
	}

	// 验证并获取SMTP授权码
	smtpAuthCode, getErr := tools.GetStringFromRawData(rawData, "server.smtp_auth_code")
	if getErr != nil {
		resp.BadRequest(ctx, "SMTP授权码配置错误", getErr.Error())
		return
	}

	// 验证并获取SMTP端口号
	smtpPort, err := tools.GetUInt16FromRawData(rawData, "server.smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "SMTP端口号有误，请检查错误", err.Error())
		return
	}

	// 使用新的配置信息发送验证邮件
	if err := email.SendVerificationCodeByArgs(
		ctx,
		config.User.UserEmail,
		smtpAccount,
		smtpAddress,
		smtpAuthCode,
		smtpPort,
	); err != nil {
		resp.Err(ctx, "发送失败", err.Error())
		return
	}

	// 发送成功，返回原始配置数据
	resp.Ok(ctx, "发送成功", nil)
}

// updateUserConfig 处理更新用户信息的请求
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 验证用户提交的验证码
//  2. 验证并获取用户基本信息(用户名、邮箱等)
//  3. 验证并获取SMTP邮件服务器配置信息
//  4. 更新用户界面相关配置(背景图、头像、网站logo等)
//  5. 将新的配置信息保存到系统中
func updateUserConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		// 解析失败时直接返回
		return
	}

	// 获取并验证用户名
	// 用户名为必填项,不能为空
	username, getErr := tools.GetStringFromRawData(rawData, "user.user_name")
	if getErr != nil {
		resp.BadRequest(ctx, "用户名配置错误", getErr.Error())
		return
	}

	// 获取并验证GitHub地址
	// GitHub地址为可选项，会提供默认值
	userGithubAddress, getErr := tools.GetStringFromRawData(rawData, "user.user_github_address")
	if getErr != nil || userGithubAddress == "" {
		userGithubAddress = "https://github.com/"
	}

	// 获取并验证用户爱好列表
	// 爱好列表为可选项,但需要是字符串数组格式
	userHobbies, getErr := tools.GetStrListFromRawData(rawData, "user.user_hobbies")
	if getErr != nil {
		resp.BadRequest(ctx, "爱好配置错误", getErr.Error())
		return
	}
	if len(userHobbies) > 10 {
		userHobbies = userHobbies[:10]
	}

	// 获取并验证打字机内容列表
	// 打字机内容为可选项,但需要是字符串数组格式
	typeWriterContent, getErr := tools.GetStrListFromRawData(rawData, "user.type_writer_content")
	if getErr != nil {
		resp.BadRequest(ctx, "打字机内容配置错误", getErr.Error())
		return
	}

	// 根据是否传递验证码判断用户是否有更新邮箱
	// 若传递，则需要从缓存中获取验证码进行验证
	// 若没有传递则只将其余的配置进行更新，不处理邮箱
	var userEmail string

	// 获取用户提交的验证码
	vefCode, getErr := tools.GetStringFromRawData(rawData, "user.verification_code")
	if getErr != nil || vefCode == "" {
		// 如果没有提供验证码,保持原有邮箱不变
		userEmail = config.User.UserEmail
	} else {
		// 从缓存中获取验证码
		vefCodeInCache, cacheErr := storage.Storage.Cache.GetString(ctx, storage.VerificationCodeKey)
		if cacheErr != nil {
			msg := fmt.Sprintf("验证码失效: %v", cacheErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}
		// 确保验证码使用后从缓存中删除
		defer func() {
			if delErr := storage.Storage.Cache.Delete(ctx, storage.VerificationCodeKey); delErr != nil {
				logger.Warn("删除验证码缓存失败: ", delErr.Error())
			}
		}()

		// 验证用户提交的验证码是否正确
		if vefCodeInCache != vefCode {
			resp.BadRequest(ctx, "验证码错误", nil)
			return
		}

		// 获取新的邮箱地址
		newEmail, err := tools.GetStringFromRawData(rawData, "user.user_email")
		if err != nil {
			resp.BadRequest(ctx, "用户邮箱配置错误", err.Error())
			return
		}
		userEmail = newEmail
	}

	icpFilingNumber, getErr := tools.GetStringFromRawData(rawData, "user.icp_filing_number")
	if getErr != nil {
		resp.BadRequest(ctx, "ICP备案号配置错误", getErr.Error())
		return
	}

	// 构造新的用户配置对象
	// 保持原有的背景图、头像和网站logo不变
	userConfig := config.UserConfigData{
		Username:          username,
		UserEmail:         userEmail,
		UserGithubAddress: userGithubAddress,
		UserHobbies:       userHobbies,
		TypeWriterContent: typeWriterContent,
		BackgroundImage:   config.User.BackgroundImage,
		AvatarImage:       config.User.AvatarImage,
		WebLogo:           config.User.WebLogo,
		ICPFilingNumber:   icpFilingNumber,
	}
	// 更新全局用户配置
	config.User = userConfig

	// 更新配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		msg := fmt.Sprintf("更新配置失败: %v", upErr.Error())
		resp.Err(ctx, msg, nil)
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// verifyNewEmail 验证新的邮箱地址并发送验证码
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证新的邮箱地址
//  2. 使用系统SMTP配置向新邮箱发送验证码
//  3. 根据发送结果返回相应的HTTP响应
func verifyNewEmail(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		// 解析失败时直接返回
		return
	}

	// 获取并验证新邮箱地址
	newEmail, getErr := tools.GetStringFromRawData(rawData, "user.user_email")
	if getErr != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", getErr.Error())
		return
	}

	// 发送验证码
	if err := email.SendVerificationCodeByArgs(
		ctx,
		newEmail,
		config.Server.SmtpAccount,
		config.Server.SmtpAddress,
		config.Server.SmtpAuthCode,
		config.Server.SmtpPort,
	); err != nil {
		resp.Err(ctx, "发送失败", err.Error())
		return
	}

	resp.Ok(ctx, "发送成功", nil)
}

// updateUserVisuals 更新用户背景图片配置
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证背景图片ID
//  2. 检查背景图片是否存在
//  3. 更新用户配置中的背景图片设置
//  4. 保存更新后的配置
func updateUserVisuals(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}

	// 获取并清理背景图片 ID、头像图片 ID 和网站 logo 图片 ID
	backgroundImgId := strings.TrimSpace(rawData["user.background_image"].(string))
	avatarImgId := strings.TrimSpace(rawData["user.avatar_image"].(string))
	logoImgId := strings.TrimSpace(rawData["user.web_logo"].(string))

	// 验证背景图片是否存在
	flag, bkgErr := adminservices.IsExistImgById(ctx, backgroundImgId)
	if bkgErr != nil {
		// 查询过程发生错误时返回错误信息
		msg := fmt.Sprintf("查找图片报错: %v", bkgErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	if !flag {
		// 背景图片不存在时返回错误信息
		resp.BadRequest(ctx, "背景图片不存在", nil)
		return
	}

	// 验证头像图片是否存在
	flag, avatarErr := adminservices.IsExistImgById(ctx, avatarImgId)
	if avatarErr != nil {
		// 查询过程发生错误时返回错误信息
		msg := fmt.Sprintf("查找图片报错: %v", avatarErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	if !flag {
		// 头像图片不存在时返回错误信息
		resp.BadRequest(ctx, "头像图片不存在", nil)
		return
	}

	// 验证网站logo图片是否存在
	flag, logoErr := adminservices.IsExistImgById(ctx, logoImgId)
	if logoErr != nil {
		// 查询过程发生错误时返回错误信息
		msg := fmt.Sprintf("查找图片报错: %v", logoErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	if !flag {
		// 网站logo图片不存在时返回错误信息
		resp.BadRequest(ctx, "网站logo图片不存在", nil)
		return
	}

	// 更新用户配置中的图片设置
	config.User.BackgroundImage = backgroundImgId
	config.User.AvatarImage = avatarImgId
	config.User.WebLogo = logoImgId

	// 保存更新后的配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		msg := fmt.Sprintf("更新配置失败: %v", upErr.Error())
		resp.Err(ctx, msg, nil)
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// getServerConfig 获取服务器配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从系统配置中获取服务器配置信息，包括token过期时间和跨域源配置
//  2. 将配置信息封装为map结构返回给客户端
func getServerConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]any{
		"port":                  config.Server.Port,
		"token_expire_duration": config.Server.TokenExpireDuration,
		"cors_origins":          config.Server.Cors.Origins,
		"smtp_account":          config.Server.SmtpAccount,
		"smtp_address":          config.Server.SmtpAddress,
		"smtp_port":             config.Server.SmtpPort,
	})
}

// updateServerConfig 更新服务器配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证新的服务器配置信息
//  2. 验证Token密钥的合法性
//  3. 验证Token过期时间的设置
//  4. 验证跨域源配置的合法性
//  5. 更新系统配置并保存
func updateServerConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		msg := fmt.Sprintf("请求数据有误，请检查错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 验证Token密钥
	tokenKey, getErr := tools.GetStringFromRawData(rawData, "server.token_key")
	if getErr != nil {
		msg := fmt.Sprintf("Token 密钥配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	if anaErr := tools.AnalyzeTokenKey(tokenKey); anaErr != nil {
		msg := fmt.Sprintf("Token 密钥配置错误: %s", anaErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 验证Token过期时间
	tokenExpireDur, getErr := tools.GetUInt8FromRawData(rawData, "server.token_expire_duration")
	if getErr != nil {
		msg := fmt.Sprintf("Token 过期时间配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 获取并验证跨域源配置
	origins, getErr := tools.GetStrListFromRawData(rawData, "server.cors_origins")
	if getErr != nil {
		msg := fmt.Sprintf("跨域源配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	anaErr := tools.AnalyzeCorsOrigins(origins)
	if anaErr != nil {
		msg := fmt.Sprintf("跨域源配置错误: %s", anaErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	smtpAccount := config.Server.SmtpAccount
	smtpAddress := config.Server.SmtpAddress
	smtpPort := config.Server.SmtpPort
	smtpAuthCode := config.Server.SmtpAuthCode

	vefCode, getErr := tools.GetStringFromRawData(rawData, "server.verification_code")
	if getErr == nil {
		// 从缓存中获取验证码
		vefCodeInCache, cacheErr := storage.Storage.Cache.GetString(ctx, storage.VerificationCodeKey)
		if cacheErr != nil {
			msg := fmt.Sprintf("验证码失效: %v", cacheErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}
		// 确保验证码使用后从缓存中删除
		defer func() {
			if delErr := storage.Storage.Cache.Delete(ctx, storage.VerificationCodeKey); delErr != nil {
				logger.Warn("删除验证码缓存失败: ", delErr.Error())
			}
		}()

		if vefCodeInCache != vefCode {
			resp.BadRequest(ctx, "验证码错误", nil)
			return
		}

		smtpAccount, getErr = tools.GetStringFromRawData(rawData, "server.smtp_account")
		if getErr != nil {
			msg := fmt.Sprintf("SMTP账号配置错误: %s", getErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}
		if anaErr := tools.AnalyzeEmail(smtpAccount); anaErr != nil {
			msg := fmt.Sprintf("SMTP账号配置错误: %s", anaErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}

		smtpAddress, getErr = tools.GetStringFromRawData(rawData, "server.smtp_address")
		if getErr != nil {
			msg := fmt.Sprintf("SMTP地址配置错误: %s", getErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}

		smtpPort, getErr = tools.GetUInt16FromRawData(rawData, "server.smtp_port")
		if getErr != nil {
			msg := fmt.Sprintf("SMTP端口配置错误: %s", getErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}

		smtpAuthCode, getErr = tools.GetStringFromRawData(rawData, "server.smtp_auth_code")
		if getErr != nil {
			msg := fmt.Sprintf("SMTP认证码配置错误: %s", getErr.Error())
			resp.BadRequest(ctx, msg, nil)
			return
		}
	}

	// 构造新的服务器配置
	config.Server = config.ServerConfigData{
		Port:                config.Server.Port,
		TokenKey:            tokenKey,
		TokenExpireDuration: tokenExpireDur,
		Cors: config.CorsConfigData{
			Origins: origins,
			Methods: config.Server.Cors.Methods,
			Headers: config.Server.Cors.Headers,
		},
		SmtpAccount:  smtpAccount,
		SmtpAddress:  smtpAddress,
		SmtpPort:     smtpPort,
		SmtpAuthCode: smtpAuthCode,
	}

	// 更新配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		resp.Err(ctx, "更新失败", upErr.Error())
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// getLoggerConfig 获取日志配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//
//	从系统配置中获取日志相关配置信息，包括:
//	- 日志级别 (level)
//	- 日志文件路径 (path)
//	- 日志文件保留时间 (max_age)
//	- 单个日志文件大小限制 (max_size)
//	- 保留的日志文件备份数量 (max_backups)
//	- 是否压缩日志文件 (compress)
func getLoggerConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]any{
		"level":       config.Logger.Level,
		"dir_path":    filepath.Dir(config.Logger.Path),
		"max_age":     config.Logger.MaxAge,
		"max_size":    config.Logger.MaxSize,
		"max_backups": config.Logger.MaxBackups,
		"compress":    config.Logger.Compress,
	})
}

// updateLoggerConfig 更新日志配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证新的日志配置信息
//  2. 验证日志级别、文件路径等参数的合法性
//  3. 更新系统日志配置并保存
//  4. 返回操作结果
func updateLoggerConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		msg := fmt.Sprintf("请求数据有误，请检查错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 获取并验证日志级别
	level := strings.TrimSpace(rawData["logger.level"].(string))
	if anaErr := tools.AnalyzeLoggerLevel(level); anaErr != nil {
		msg := fmt.Sprintf("日志级别配置错误: %s", anaErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 获取并验证日志文件路径
	dirPath, anaErr := tools.AnalyzeAbsolutePath(rawData["logger.path"].(string))
	if anaErr != nil {
		msg := fmt.Sprintf("日志文件路径配置错误: %s", anaErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	dirPath = filepath.Join(dirPath, "sparrow_blog.log")

	// 获取并验证日志文件保留时间
	maxAge, getErr := tools.GetUInt16FromRawData(rawData, "logger.max_age")
	if getErr != nil {
		msg := fmt.Sprintf("日志文件保留时间配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 获取并验证单个日志文件大小限制
	maxSize, getErr := tools.GetUInt16FromRawData(rawData, "logger.max_size")
	if getErr != nil {
		msg := fmt.Sprintf("单个日志文件大小限制配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 获取并验证保留的日志文件备份数量
	maxBackups, getErr := tools.GetUInt16FromRawData(rawData, "logger.max_backups")
	if getErr != nil {
		msg := fmt.Sprintf("保留的日志文件备份数量配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 获取并验证是否压缩日志文件
	compress, getErr := tools.GetBoolFromRawData(rawData, "logger.compress")
	if getErr != nil {
		msg := fmt.Sprintf("是否压缩日志文件配置错误: %s", getErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 构造新的日志配置
	config.Logger = config.LoggerConfigData{
		Level:      level,      // 日志级别
		Path:       dirPath,    // 日志目录路径
		MaxAge:     maxAge,     // 日志文件保留时间(天)
		MaxSize:    maxSize,    // 单个日志文件大小限制(MB)
		MaxBackups: maxBackups, // 保留的日志文件备份数量
		Compress:   compress,   // 是否压缩日志文件
	}

	// 更新配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		resp.Err(ctx, "更新失败", upErr.Error())
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// getMysqlConfig 获取MySQL数据库配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从系统配置中获取MySQL相关配置信息，包括:
//     - 主机地址 (host)
//     - 端口号 (port)
//     - 数据库名称 (database)
//     - 用户名 (user)
//     - 最大连接数 (max_open)
//     - 最大空闲连接数 (max_idle)
//  2. 将配置信息封装为map结构返回给客户端
func getMysqlConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]any{
		"user":     config.MySQL.User,
		"host":     config.MySQL.Host,
		"port":     config.MySQL.Port,
		"database": config.MySQL.DB,
		"max_open": config.MySQL.MaxOpen,
		"max_idle": config.MySQL.MaxIdle,
	})
}

// updateMysqlConfig 更新MySQL数据库配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证MySQL配置参数
//  2. 验证各项参数的合法性，包括:
//     - 用户名和密码
//     - 主机地址和端口
//     - 数据库名称
//     - 连接池配置(最大连接数和空闲连接数)
//  3. 测试数据库连接配置的有效性
//  4. 更新系统配置并保存
func updateMysqlConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		msg := fmt.Sprintf("请求数据有误，请检查错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 初始化MySQL配置结构体。
	mysqlConfig := config.MySQLConfigData{}

	// 获取并验证数据库用户名
	user := strings.TrimSpace(rawData["mysql.user"].(string))
	if len(user) == 0 {
		resp.BadRequest(ctx, "数据库用户名不能为空", nil)
		return
	}
	mysqlConfig.User = user

	// 获取并验证数据库密码
	mysqlConfig.Password = strings.TrimSpace(rawData["mysql.password"].(string))

	// 获取并验证数据库主机地址
	host := strings.TrimSpace(rawData["mysql.host"].(string))
	if anaErr := tools.AnalyzeHostAddress(host); anaErr != nil {
		msg := fmt.Sprintf("数据库主机地址配置错误: %s", anaErr.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	mysqlConfig.Host = host

	// 获取并验证数据库端口号
	port, err := tools.GetUInt16FromRawData(rawData, "mysql.port")
	if err != nil {
		msg := fmt.Sprintf("数据库端口号配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	mysqlConfig.Port = port

	// 获取并验证数据库名称
	db := strings.TrimSpace(rawData["mysql.database"].(string))
	if len(db) == 0 {
		resp.BadRequest(ctx, "数据库名称不能为空", nil)
		return
	}
	mysqlConfig.DB = db

	// 获取并验证最大连接数
	maxOpen, err := tools.GetUInt16FromRawData(rawData, "mysql.max_open")
	if err != nil {
		msg := fmt.Sprintf("最大连接数配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	// 获取并验证最大空闲连接数
	maxIdle, err := tools.GetUInt16FromRawData(rawData, "mysql.max_idle")
	if err != nil {
		msg := fmt.Sprintf("最大空闲连接数配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 检查最大空闲连接数是否大于最大打开连接数。
	if maxIdle > maxOpen {
		resp.BadRequest(ctx, "最大空闲连接数不能大于最大打开连接数", nil)
		return
	}

	// 将获取的最大连接数和最大空闲连接数赋值给MySQL配置结构体。
	mysqlConfig.MaxOpen = maxOpen
	mysqlConfig.MaxIdle = maxIdle

	// 验证MySQL连接配置。
	if err = tools.AnalyzeMySqlConnect(&mysqlConfig); err != nil {
		// 如果连接配置验证失败，返回400错误响应。
		msg := fmt.Sprintf("数据库连接配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 将MySQL配置赋值给全局变量。
	config.MySQL = mysqlConfig

	// 更新配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		msg := fmt.Sprintf("更新失败: %s", upErr.Error())
		resp.Err(ctx, msg, nil)
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// getOssConfig 获取对象存储(OSS)配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从系统配置中获取OSS相关配置信息，包括:
//     - 访问端点 (endpoint)
//     - 地域信息 (region)
//     - 存储桶名称 (bucket)
//     - 图片存储路径 (image_oss_path)
//     - 博客内容存储路径 (blog_oss_path)
//  2. 将配置信息封装为map结构返回给客户端
func getOssConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]any{
		"endpoint":       config.Oss.Endpoint,
		"region":         config.Oss.Region,
		"bucket":         config.Oss.Bucket,
		"image_oss_path": config.Oss.ImageOssPath,
		"blog_oss_path":  config.Oss.BlogOssPath,
	})
}

// updateOssConfig 更新对象存储(OSS)配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证OSS配置参数
//  2. 验证各项参数的合法性，包括:
//     - 访问端点(endpoint)
//     - 地域信息(region)
//     - 访问密钥(access key)
//     - 存储桶名称(bucket)
//     - 图片存储路径(image path)
//  3. 测试OSS连接配置的有效性
//  4. 更新系统配置并保存
func updateOssConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		return
	}
	// 初始化OSS配置结构体。
	ossConfig := config.OssConfig{}

	// 从原始数据中提取并清理OSS配置参数
	ossConfig.Endpoint = strings.TrimSpace(rawData["oss.endpoint"].(string))                 // OSS访问端点
	ossConfig.Region = strings.TrimSpace(rawData["oss.region"].(string))                     // OSS地域信息
	ossConfig.AccessKeyId = strings.TrimSpace(rawData["oss.access_key_id"].(string))         // 访问密钥ID
	ossConfig.AccessKeySecret = strings.TrimSpace(rawData["oss.access_key_secret"].(string)) // 访问密钥密文
	ossConfig.Bucket = strings.TrimSpace(rawData["oss.bucket"].(string))                     // 存储桶名称
	ossConfig.ImageOssPath = strings.TrimSpace(rawData["oss.image_oss_path"].(string))       // 图片存储路径

	// 验证OSS配置。
	if err = tools.AnalyzeOssConfig(&ossConfig); err != nil {
		// 如果连接配置验证失败，返回400错误响应。
		msg := fmt.Sprintf("OSS连接配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 验证图片存储路径
	imageOssPath := strings.TrimSpace(rawData["oss.image_oss_path"].(string))
	if err = tools.AnalyzeOssPath(imageOssPath); err != nil {
		// 如果连接配置验证失败，返回400错误响应。
		msg := fmt.Sprintf("图片 OSS 路径配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	ossConfig.ImageOssPath = imageOssPath

	// 验证博客内容存储路径
	blogOssPath := strings.TrimSpace(rawData["oss.blog_oss_path"].(string))
	if err = tools.AnalyzeOssPath(blogOssPath); err != nil {
		// 如果连接配置验证失败，返回400错误响应。
		msg := fmt.Sprintf("博客 OSS 路径配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	ossConfig.BlogOssPath = blogOssPath

	// 将OSS配置赋值给全局变量
	config.Oss = ossConfig

	// 更新配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		msg := fmt.Sprintf("更新失败: %s", upErr.Error())
		resp.Err(ctx, msg, nil)
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// getCacheAndIndexConfig 获取缓存配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从系统配置中获取缓存相关配置信息，包括:
//     - AOF持久化是否启用 (enable_aof)
//     - AOF文件存储目录路径 (aof_dir_path)
//     - AOF文件大小限制 (aof_mix_size)
//     - AOF文件是否压缩 (aof_compress)
//  2. 将配置信息封装为map结构返回给客户端
func getCacheAndIndexConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]any{
		"enable_aof":   config.Cache.Aof.Enable,
		"aof_dir_path": filepath.Dir(config.Cache.Aof.Path),
		"aof_mix_size": config.Cache.Aof.MaxSize,
		"aof_compress": config.Cache.Aof.Compress,
		"index_path":   config.SearchEngine.IndexPath,
	})
}

// updateCacheAndIndexConfig 更新缓存配置信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析并验证缓存配置参数
//  2. 验证各项参数的合法性，包括:
//     - AOF持久化开关
//     - AOF文件存储路径
//     - AOF文件大小限制
//     - AOF文件压缩选项
//  3. 更新系统配置并保存
//  4. 返回操作结果
func updateCacheAndIndexConfig(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		msg := fmt.Sprintf("解析请求数据失败: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 初始化缓存配置结构体。
	cacheConfig := config.CacheConfig{}

	// 从原始数据中提取并清理缓存配置参数
	cacheConfig.Aof.Enable, err = tools.GetBoolFromRawData(rawData, "cache.aof.enable") // AOF持久化是否启用
	if err != nil {
		msg := fmt.Sprintf("AOF持久化开关配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	aofDirPath, err := tools.AnalyzeAbsolutePath(rawData["cache.aof.path"].(string))
	if err != nil {
		msg := fmt.Sprintf("AOF路径配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	if strings.HasSuffix(aofDirPath, "/aof") {
		cacheConfig.Aof.Path = filepath.Join(aofDirPath, "sparrow_blog.aof")
	} else {
		cacheConfig.Aof.Path = filepath.Join(aofDirPath, "aof", "sparrow_blog.aof")
	}

	cacheConfig.Aof.MaxSize, err = tools.GetUInt16FromRawData(rawData, "cache.aof.max_size")
	if err != nil {
		msg := fmt.Sprintf("AOF文件大小限制配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	cacheConfig.Aof.Compress, err = tools.GetBoolFromRawData(rawData, "cache.aof.compress")
	if err != nil {
		msg := fmt.Sprintf("AOF文件压缩选项配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 从请求中获取文本搜索引擎索引文件路径
	indexPath, err := tools.GetStringFromRawData(rawData, "search_engine.index_path")
	if err != nil {
		msg := fmt.Sprintf("索引文件路径解析错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}
	indexPath, err = tools.AnalyzeAbsolutePath(indexPath)
	if err != nil {
		msg := fmt.Sprintf("索引文件路径配置错误: %s", err.Error())
		resp.BadRequest(ctx, msg, nil)
		return
	}

	// 将缓存配置赋值给全局变量。
	config.Cache = cacheConfig

	// 更新索引文件路径
	config.SearchEngine.IndexPath = indexPath

	// 更新配置到存储系统
	if upErr := adminservices.UpdateConfig(); upErr != nil {
		msg := fmt.Sprintf("更新失败: %s", upErr.Error())
		resp.Err(ctx, msg, nil)
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}

// rebuildIndex 重建搜索索引接口
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 接收管理员的重建索引请求
//  2. 调用搜索引擎的重建索引功能
//  3. 支持超时控制，防止长时间阻塞
//  4. 返回重建结果给客户端
//
// HTTP方法: PUT
// 路径: /admin/setting/cache-index/rebuild-index
// 权限: 需要管理员JWT认证
//
// 响应格式:
//   - 成功: {"code": 200, "message": "重建索引成功", "data": {"duration_ms": 12345}}
//   - 失败: {"code": 500, "message": "重建索引失败", "data": "错误详情"}
func rebuildIndex(ctx *gin.Context) {
	logger.Info("管理员请求重建搜索索引")

	// 创建带超时的上下文，防止重建过程过长
	// 设置10分钟超时，对于大量文档的重建应该足够
	rebuildCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 记录开始时间，用于计算重建耗时
	startTime := time.Now()

	// 调用搜索引擎的重建索引功能
	err := searchengine.RebuildIndex(rebuildCtx)
	if err != nil {
		// 重建失败，记录错误日志并返回错误响应
		logger.Error("重建搜索索引失败: " + err.Error())

		// 根据错误类型返回不同的响应
		if errors.Is(err, context.DeadlineExceeded) {
			resp.Err(ctx, "重建索引超时", "重建过程超过10分钟限制，请检查系统资源或联系管理员")
		} else if errors.Is(err, context.Canceled) {
			resp.Err(ctx, "重建索引被取消", "重建过程被用户或系统取消")
		} else {
			resp.Err(ctx, "重建索引失败", err.Error())
		}
		return
	}

	// 计算重建耗时
	duration := time.Since(startTime)
	durationMs := float64(duration) / float64(time.Millisecond)

	// 记录成功日志
	logger.Info(fmt.Sprintf("重建搜索索引成功，耗时: %.2f毫秒", durationMs))

	// 返回成功响应，包含重建耗时信息
	resp.Ok(ctx, "重建索引成功", map[string]any{
		"duration_ms":     durationMs,
		"duration_string": duration.String(),
		"message":         "搜索索引已成功重建，所有文档已重新索引",
	})
}

// ================== 友链管理相关 Handler ==================

// getAllFriendLinks 获取所有友链信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 调用服务层获取所有友链数据
//  2. 将DTO转换为VO格式返回给前端
//  3. 返回友链列表给客户端
func getAllFriendLinks(ctx *gin.Context) {
	// 调用服务层获取所有友链数据
	friendLinkDtos, err := adminservices.GetAllFriendLinks(ctx)
	if err != nil {
		resp.Err(ctx, "获取友链失败", err.Error())
		return
	}

	// 将DTO列表转换为VO列表，以便前端使用
	friendLinkVos := make([]vo.FriendLinkVo, 0, len(friendLinkDtos))
	for _, friendLinkDto := range friendLinkDtos {
		friendLinkVo := vo.FriendLinkVo{
			FriendLinkId:    friendLinkDto.FriendLinkId,
			FriendLinkName:  friendLinkDto.FriendLinkName,
			FriendLinkUrl:   friendLinkDto.FriendLinkUrl,
			FriendAvatarUrl: friendLinkDto.FriendAvatarUrl,
			FriendDescribe:  friendLinkDto.FriendDescribe,
			Display:         friendLinkDto.Display,
		}
		friendLinkVos = append(friendLinkVos, friendLinkVo)
	}

	// 返回成功响应
	resp.Ok(ctx, "获取友链成功", friendLinkVos)
}

// updateFriendLink 更新友链信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从请求中解析友链更新信息
//  2. 验证友链ID和必要字段
//  3. 调用服务层更新友链
//  4. 返回更新结果给客户端
func updateFriendLink(ctx *gin.Context) {
	// 从请求中解析原始数据
	rawData, err := tools.GetMapFromRawData(ctx)
	if err != nil {
		resp.BadRequest(ctx, "解析请求数据失败", err.Error())
		return
	}

	// 解析友链ID
	friendLinkId, err := tools.GetStringFromRawData(rawData, "friend_link_id")
	if err != nil {
		resp.BadRequest(ctx, "友链ID解析错误", err.Error())
		return
	}
	if strings.TrimSpace(friendLinkId) == "" {
		resp.BadRequest(ctx, "友链ID不能为空", "")
		return
	}

	// 解析友链名称
	friendLinkName, err := tools.GetStringFromRawData(rawData, "friend_link_name")
	if err != nil {
		resp.BadRequest(ctx, "友链名称解析错误", err.Error())
		return
	}
	if strings.TrimSpace(friendLinkName) == "" {
		resp.BadRequest(ctx, "友链名称不能为空", "")
		return
	}

	// 解析友链URL
	friendLinkUrl, err := tools.GetStringFromRawData(rawData, "friend_link_url")
	if err != nil {
		resp.BadRequest(ctx, "友链URL解析错误", err.Error())
		return
	}
	if strings.TrimSpace(friendLinkUrl) == "" {
		resp.BadRequest(ctx, "友链URL不能为空", "")
		return
	}

	// 解析友链头像URL（可选）
	friendAvatarUrl, err := tools.GetStringFromRawData(rawData, "friend_avatar_url")
	if err != nil {
		friendAvatarUrl = ""
	}

	// 解析友链描述（可选）
	friendDescribe, err := tools.GetStringFromRawData(rawData, "friend_describe")
	if err != nil {
		friendDescribe = ""
	}

	// 解析显示状态（可选，默认为 false）
	display, err := tools.GetBoolFromRawData(rawData, "display")
	if err != nil {
		display = false // 默认不展示
	}

	// 构造友链DTO
	friendLinkDto := &dto.FriendLinkDto{
		FriendLinkId:    strings.TrimSpace(friendLinkId),
		FriendLinkName:  strings.TrimSpace(friendLinkName),
		FriendLinkUrl:   strings.TrimSpace(friendLinkUrl),
		FriendAvatarUrl: friendAvatarUrl,
		FriendDescribe:  friendDescribe,
		Display:         display,
	}

	// 调用服务层更新友链
	err = adminservices.UpdateFriendLink(ctx, friendLinkDto)
	if err != nil {
		resp.Err(ctx, "更新友链失败", err.Error())
		return
	}

	// 返回成功响应
	resp.Ok(ctx, "更新友链成功", nil)
}

// deleteFriendLink 删除友链
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从路径参数中获取友链ID
//  2. 验证友链ID的有效性
//  3. 调用服务层删除友链
//  4. 返回删除结果给客户端
func deleteFriendLink(ctx *gin.Context) {
	// 从路径参数中获取友链ID
	friendLinkId := ctx.Param("friend_link_id")
	if strings.TrimSpace(friendLinkId) == "" {
		resp.BadRequest(ctx, "友链ID不能为空", "")
		return
	}

	// 调用服务层删除友链
	err := adminservices.DeleteFriendLinkById(ctx, strings.TrimSpace(friendLinkId))
	if err != nil {
		resp.Err(ctx, "删除友链失败", err.Error())
		return
	}

	// 返回成功响应
	resp.Ok(ctx, "删除友链成功", nil)
}

// updateFriendLinkDisplay 切换友链显示状态
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从路径参数中获取友链ID
//  2. 验证友链ID的有效性
//  3. 调用服务层切换友链显示状态
//  4. 返回切换结果给客户端
func updateFriendLinkDisplay(ctx *gin.Context) {
	// 从路径参数中获取友链ID
	friendLinkId := ctx.Param("friend_link_id")
	if strings.TrimSpace(friendLinkId) == "" {
		resp.BadRequest(ctx, "友链ID不能为空", "")
		return
	}

	// 调用服务层切换友链显示状态
	newDisplay, err := adminservices.UpdateFriendLinkDisplay(ctx, strings.TrimSpace(friendLinkId))
	if err != nil {
		resp.Err(ctx, "切换友链显示状态失败", err.Error())
		return
	}

	// 返回成功响应，包含切换后的状态
	resp.Ok(ctx, "切换友链显示状态成功", map[string]bool{
		"display": newDisplay,
	})
}
