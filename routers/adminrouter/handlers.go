package adminrouter

import (
	"github.com/gin-gonic/gin"
	"h2blog_server/email"
	"h2blog_server/internal/model/vo"
	"h2blog_server/internal/services/adminservice"
	"h2blog_server/internal/services/imgservice"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/resp"
	"h2blog_server/routers/tools"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"strconv"
	"strings"
	"time"
)

// sendLoginVerificationCode 处理发送验证码的请求。
// 参数:
//   - *gin.Context: HTTP 请求上下文，包含请求数据和响应方法。
//
// 功能描述:
//
//	该函数从请求中解析用户提交的数据，验证用户邮箱是否正确，
//	并调用邮件服务发送验证码。根据操作结果返回相应的 HTTP 响应。
func sendLoginVerificationCode(ctx *gin.Context) {
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
	err = email.SendVerificationCodeBySys(ctx)
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

// getUserConfig 获取用户基本信息
// 参数:
//   - ctx *gin.Context: HTTP请求上下文，包含请求参数和响应方法
//
// 功能描述:
//  1. 从系统配置中获取用户信息，包括用户名、邮箱、SMTP配置等
//  2. 将获取到的用户信息封装为map结构返回给客户端
//  3. 返回成功的HTTP响应
func getUserConfig(ctx *gin.Context) {
	resp.Ok(ctx, "获取成功", map[string]string{
		"user_name":        config.User.Username,
		"user_email":       config.User.UserEmail,
		"smtp_account":     config.User.SmtpAccount,
		"smtp_address":     config.User.SmtpAddress,
		"smtp_port":        strconv.Itoa(int(config.User.SmtpPort)),
		"background_image": config.User.BackgroundImage,
		"avatar_image":     config.User.AvatarImage,
		"web_logo":         config.User.WebLogo,
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

	// 验证并获取新的邮箱地址
	newEmail := strings.TrimSpace(rawData["user_email"].(string))
	if anaErr := tools.AnalyzeEmail(newEmail); anaErr != nil {
		resp.BadRequest(ctx, "邮箱格式有误，请检查错误", anaErr.Error())
		return
	}

	// 验证并获取SMTP账号
	smtpAccount := strings.TrimSpace(rawData["smtp_account"].(string))
	if len(smtpAccount) == 0 {
		resp.BadRequest(ctx, "SMTP账号不能为空", "")
		return
	}

	// 验证并获取SMTP服务器地址
	smtpAddress := strings.TrimSpace(rawData["smtp_address"].(string))
	if len(smtpAddress) == 0 {
		resp.BadRequest(ctx, "SMTP地址不能为空", "")
		return
	}

	// 验证并获取SMTP授权码
	smtpAuthCode := strings.TrimSpace(rawData["smtp_auth_code"].(string))
	if len(smtpAuthCode) == 0 {
		resp.BadRequest(ctx, "SMTP授权码不能为空", "")
		return
	}

	// 验证并获取SMTP端口号
	smtpPort, err := tools.GetUInt16FromRawData(rawData, "smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "SMTP端口号有误，请检查错误", err.Error())
		return
	}

	// 使用新的配置信息发送验证邮件
	if err := email.SendVerificationCodeByArgs(
		ctx,
		newEmail,
		smtpAccount,
		smtpAddress,
		smtpAuthCode,
		smtpPort,
	); err != nil {
		resp.Err(ctx, "发送失败", err.Error())
		return
	}

	// 发送成功，返回原始配置数据
	resp.Ok(ctx, "发送成功", rawData)
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
		resp.BadRequest(ctx, "请求数据有误，请检查错误", err.Error())
		return
	}

	// 从缓存中获取验证码
	verifiedCode, err := storage.Storage.Cache.GetString(ctx, storage.VerificationCodeKey)
	if err != nil {
		resp.BadRequest(ctx, "验证码过期", err.Error())
		return
	}
	defer func() {
		if err := storage.Storage.Cache.Delete(ctx, storage.VerificationCodeKey); err != nil {
			logger.Warn("删除验证码缓存失败: ", err.Error())
		}
	}()

	// 验证用户提交的验证码是否正确
	if verifiedCode != rawData["verified_code"].(string) {
		resp.BadRequest(ctx, "验证码错误", "")
		return
	}

	// 获取并验证用户名
	userName := strings.TrimSpace(rawData["user_name"].(string))
	if len(userName) == 0 {
		resp.BadRequest(ctx, "用户名不能为空", "")
		return
	}

	// 获取并验证用户邮箱格式
	userEmail := strings.TrimSpace(rawData["user_email"].(string))
	if anaErr := tools.AnalyzeEmail(userEmail); anaErr != nil {
		resp.BadRequest(ctx, "用户邮箱配置错误", anaErr.Error())
		return
	}

	// 获取并验证SMTP账号邮箱格式
	smtpAccount := strings.TrimSpace(rawData["smtp_account"].(string))
	if anaErr := tools.AnalyzeEmail(smtpAccount); anaErr != nil {
		resp.BadRequest(ctx, "系统邮箱配置错误", anaErr.Error())
		return
	}

	// 获取SMTP服务器地址
	smtpAddress := strings.TrimSpace(rawData["smtp_address"].(string))

	// 获取并验证SMTP端口号
	smtpPort, err := tools.GetUInt16FromRawData(rawData, "smtp_port")
	if err != nil {
		resp.BadRequest(ctx, "系统邮箱端口配置错误", err.Error())
		return
	}

	// 获取SMTP授权码
	smtpAuthCode := strings.TrimSpace(rawData["smtp_auth_code"].(string))

	// 获取用户界面相关配置
	backgroundImage := strings.TrimSpace(rawData["background_image"].(string))
	avatarImage := strings.TrimSpace(rawData["avatar_image"].(string))
	webLogo := strings.TrimSpace(rawData["web_logo"].(string))

	// 构造新的用户配置对象
	userConfig := config.UserConfigData{
		Username:        userName,
		UserEmail:       userEmail,
		SmtpAccount:     smtpAccount,
		SmtpAddress:     smtpAddress,
		SmtpPort:        smtpPort,
		SmtpAuthCode:    smtpAuthCode,
		BackgroundImage: backgroundImage,
		AvatarImage:     avatarImage,
		WebLogo:         webLogo,
	}
	config.User = userConfig

	// 更新配置到存储系统
	if upErr := adminservice.UpdateConfig(); upErr != nil {
		resp.Err(ctx, "更新失败", upErr.Error())
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
		"token_expire_duration": config.Server.TokenExpireDuration,
		"cors_origins":          config.Server.Cors.Origins,
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
		resp.BadRequest(ctx, "请求数据有误，请检查错误", err.Error())
		return
	}

	// 验证Token密钥
	tokenKey := strings.TrimSpace(rawData["token_key"].(string))
	if anaErr := tools.AnalyzeTokenKey(tokenKey); anaErr != nil {
		resp.BadRequest(ctx, "Token 密钥配置错误", anaErr.Error())
		return
	}

	// 验证Token过期时间
	tokenExpireDur, getErr := tools.GetUInt8FromRawData(rawData, "token_expire_duration")
	if getErr != nil {
		resp.BadRequest(ctx, "Token 过期时间配置错误", getErr.Error())
		return
	}

	// 获取并验证跨域源配置
	origins, getErr := tools.GetStrListFromRawData(rawData, "cors_origins")
	if getErr != nil {
		resp.BadRequest(ctx, "跨域源配置错误", getErr.Error())
		return
	}
	anaErr := tools.AnalyzeCorsOrigins(origins)
	if anaErr != nil {
		resp.BadRequest(ctx, "跨域源配置错误", anaErr.Error())
		return
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
	}

	// 更新配置到存储系统
	if upErr := adminservice.UpdateConfig(); upErr != nil {
		resp.Err(ctx, "更新失败", upErr.Error())
		return
	}

	// 返回更新成功的响应
	resp.Ok(ctx, "更新成功", nil)
}
