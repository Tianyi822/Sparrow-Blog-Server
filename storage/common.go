package storage

// ImgCacheKeyPrefix 图片缓存 key 前缀
const ImgCacheKeyPrefix = "img_cache_"

func BuildImgCacheKey(imgId string) string {
	return ImgCacheKeyPrefix + imgId
}
