package storage

import (
	"fmt"
	"time"
)

// VerificationCodeKey 验证码缓存 key
const VerificationCodeKey = "verification_code"

// UserRevokedTokenKeyPre 用户已撤销的 token key 前缀
const UserRevokedTokenKeyPre = "user_invoked_token_"

// ImgCacheKeyPrefix 图片缓存 key 前缀
const ImgCacheKeyPrefix = "img_cache_"

// BlogCacheKeyPrefix 博客缓存 key 前缀
const BlogCacheKeyPrefix = "blog_cache_"

const BlogReadCountKeyPrefix = "blog_read_count_"

func BuildImgCacheKey(imgId string) string {
	return ImgCacheKeyPrefix + imgId
}

func BuildBlogCacheKey(blogId string) string {
	return BlogCacheKeyPrefix + blogId
}

// BuildBlogReadCountKey 构建博客阅读数缓存 key，缓存 key 格式：blog_read_count_<blogId>_<date>
func BuildBlogReadCountKey(blogId string) string {
	// 获取当前日期
	year, month, day := time.Now().Date()
	date := fmt.Sprintf("%d%02d%02d", year, month, day)
	return fmt.Sprintf("%s%s-%s", BlogReadCountKeyPrefix, blogId, date)
}
