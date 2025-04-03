package imgservice

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/cache"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/internal/model/vo"
	"h2blog_server/internal/repositories/imgrepo"
	"h2blog_server/pkg/utils"
	"h2blog_server/pkg/webp"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"time"
)

// GetPresignUrlById 根据图片ID获取预签名URL。
// 参数:
//   - ctx: 上下文对象，用于控制请求的生命周期和传递上下文信息。
//   - imgId: 图片的唯一标识符，用于定位具体的图片资源。
//
// 返回值:
//   - string: 预签名URL，用于访问图片资源。
//   - error: 如果在获取缓存、查询图片信息或生成预签名URL过程中发生错误，则返回相应的错误信息。
func GetPresignUrlById(ctx context.Context, imgId string) (string, error) {
	// 尝试从缓存中获取图片的预签名URL。
	url, err := storage.Storage.Cache.GetString(ctx, storage.BuildImgCacheKey(imgId))
	if errors.Is(err, cache.ErrNotFound) {
		// 如果缓存中未找到对应的URL，则从数据库中查询图片信息。
		imgDto, err := imgrepo.FindImgById(ctx, imgId)
		if err != nil {
			return "", err
		}

		// 根据图片名称和类型生成OSS存储路径。
		ossPath := ossstore.GenOssSavePath(imgDto.ImgName, imgDto.ImgType)

		// 为生成的OSS路径创建预签名URL，有效期为35分钟。
		presign, err := storage.Storage.PreSignUrl(ctx, ossPath, ossstore.Get, 35*time.Minute)
		if err != nil {
			return "", err
		}

		// 缓存预签名URL到缓存中。
		err = storage.Storage.Cache.SetWithExpired(ctx, storage.BuildImgCacheKey(imgId), presign.URL, 30*time.Minute)
		if err != nil {
			return "", err
		}

		url = presign.URL
	}

	// 返回获取到的预签名URL。
	return url, nil
}

// ConvertAndAddImg 添加图片并转换
// - ctx 是上下文对象，用于控制请求的生命周期
// - imgsDto 是包含图片信息的 DTO 对象
//
// 返回值
// - vo.ImgInfosVo 是包含转换成功和失败的图片信息的 VO 对象
// - error 是可能出现的错误信息
func ConvertAndAddImg(ctx context.Context, imgDtos []dto.ImgDto) (*vo.ImgInfosVo, error) {
	if !webp.Converter.IsEmpty() {
		return nil, fmt.Errorf("转换器中还有未完成的任务")
	}

	err := webp.Converter.AddBatchTasks(ctx, imgDtos)
	if err != nil {
		return nil, err
	}

	// 获取转换器中的输出通道
	outputCh := webp.Converter.GetOutputCh()

	// 用于保存以下几种数据
	// - 转换成功
	// - 转换失败
	// - 上传失败
	var imgPos []po.H2Img

	var successImgsVo []vo.ImgVo
	var failImgsVo []vo.ImgVo

	for {
		select {
		case <-ctx.Done():
			return handleConvertedImgsData(ctx, imgPos, successImgsVo, failImgsVo)
		case data, ok := <-outputCh:
			if ok { // 通道未关闭

				// 生成 ID
				imgId, err := utils.GenId(data.ImgDto.ImgName)
				if err != nil {
					continue
				}

				imgPo := po.H2Img{
					ImgId:   imgId,
					ImgName: data.ImgDto.ImgName,
				}

				if data.Flag { // 转换成功，存入数据库
					// 构建 po 对象
					imgPo.ImgType = ossstore.Webp
					// 将转换成功的数据暂存到 imgPos 中
					imgPos = append(imgPos, imgPo)
					// 将转换成功的图片信息存入 successImgsVo 中
					successImgsVo = append(successImgsVo, vo.ImgVo{
						ImgId:   imgId,
						ImgName: data.ImgDto.ImgName,
					})
				} else { // 转换失败，存入失败列表
					// 将 err 对象转换为 webp 自定义的 Err 对象
					var webpErr *webp.Err
					errors.As(data.Err, &webpErr)
					// 转换失败则需要根据不同的 Err Code 进行操作
					switch webpErr.Code {
					case webp.ConvertError | webp.UploadError:
						// 转换失败和上传失败，数据还是存在于 Oss 中，数据库则保存原来格式
						imgPo.ImgType = data.ImgDto.ImgType
						imgPos = append(imgPos, imgPo)
						// 但还是要把转换失败的放入失败列表
						failImgsVo = append(failImgsVo, vo.ImgVo{
							ImgId:   imgId,
							ImgName: data.ImgDto.ImgName,
							Err:     data.Err.Error(),
						})
					default:
						// 其他错误，包括下载失败和删除失败，直接返回
						// TODO: 这里有个问题，删除失败的话，Oss 中会同时存在原格式以及 webp 格式的图片，以后再添加功能自动扫描
						failImgsVo = append(failImgsVo, vo.ImgVo{
							ImgName: data.ImgDto.ImgName,
							Err:     data.Err.Error(),
						})
					}
				}
			} else { // 通道关闭
				return handleConvertedImgsData(ctx, imgPos, successImgsVo, failImgsVo)
			}
		case <-webp.Converter.GetCompletionStatus():
			return handleConvertedImgsData(ctx, imgPos, successImgsVo, failImgsVo)
		}
	}
}

// handleConvertedImgsData 处理转换后的图片数据
func handleConvertedImgsData(ctx context.Context, imgPos []po.H2Img, successImgsVo, failImgsVo []vo.ImgVo) (*vo.ImgInfosVo, error) {
	// 图片 vo 对象，包含压缩成功的和未成功的
	imgInfosVo := &vo.ImgInfosVo{}
	// 保存数据到数据库
	_, err := imgrepo.AddImgInfoBatch(ctx, imgPos)
	if err != nil {
		return imgInfosVo, err
	}
	// 将成功和失败的数据返回
	imgInfosVo.Success = successImgsVo
	imgInfosVo.Failure = failImgsVo
	return imgInfosVo, nil
}
