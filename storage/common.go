package storage

// VerificationCodeKey 验证码缓存 key
const VerificationCodeKey = "verification_code"

// TokenCodeKey token 缓存 key
const TokenCodeKey = "token_code_key"

// ImgCacheKeyPrefix 图片缓存 key 前缀
const ImgCacheKeyPrefix = "img_cache_"

// BlogCacheKeyPrefix 博客缓存 key 前缀
const BlogCacheKeyPrefix = "blog_cache_"

func BuildImgCacheKey(imgId string) string {
	return ImgCacheKeyPrefix + imgId
}

func BuildBlogCacheKey(blogId string) string {
	return BlogCacheKeyPrefix + blogId
}
