package storage

// VerificationCodeKey 验证码缓存 key
const VerificationCodeKey = "verification_code"

// UserRevokedTokenKeyPre 用户已撤销的 token key 前缀
const UserRevokedTokenKeyPre = "user_invoked_token_"

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
