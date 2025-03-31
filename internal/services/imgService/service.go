package imgService

import (
	"context"
	"errors"
	"fmt"
	"h2blog_server/internal/model/dto"
	"h2blog_server/internal/model/po"
	"h2blog_server/internal/model/vo"
	"h2blog_server/internal/repository/imgInfoRepo"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/pkg/utils"
	"h2blog_server/pkg/webp"
	"h2blog_server/storage"
	"h2blog_server/storage/ossstore"
	"time"
)

// GetPreSignUrlOfImg 获取图片预签名 URL
func GetPreSignUrlOfImg(ctx context.Context) (string, error) {
	// 查看缓存中是否存在
	if url, err := storage.Storage.Cache.GetString(ctx, config.User.BackgroundImage); err == nil {
		return url, nil
	}

	imgPo, err := imgInfoRepo.GetBackgroundImg(ctx)
	if err != nil {
		return "", err
	}

	result, err := storage.Storage.PreSignUrl(ctx, ossstore.GenOssSavePath(imgPo.ImgName, imgPo.ImgType), ossstore.Get, 10*time.Minute)
	if err != nil {
		return "", err
	}

	err = storage.Storage.Cache.SetWithExpired(ctx, config.User.BackgroundImage, result.URL, result.Expiration.Sub(time.Now()))
	if err != nil {
		msg := fmt.Sprintf("缓存图片预签名 URL 失败: %v", err)
		logger.Warn(msg)
		return "", errors.New(msg)
	}

	return result.URL, nil
}

// FindImgById 根据图片ID查询单条图片信息
// - ctx: 上下文对象
// - imgId: 图片ID
//
// 返回值
// - vo.ImgInfoVo 是包含查询结果的 VO 对象
// - error 是可能出现的错误信息
func FindImgById(ctx context.Context, imgId string) (*vo.ImgInfoVo, error) {
	// TODO: 这里应该返回的是一个可访问图片的 Url，该 Url 有访问时限，但在前端开发到这个功能之前，暂时都是返回一个 vo

	imgPo, err := imgInfoRepo.FindImgById(ctx, imgId)
	if err != nil {
		return nil, err
	}

	return &vo.ImgInfoVo{
		ImgId:   imgPo.ImgId,
		ImgName: imgPo.ImgName,
		ImgType: imgPo.ImgType,
	}, nil
}

// FindImgsByNameLike 根据图片名模糊查询图片信息
// - ctx 是上下文对象，用于控制请求的生命周期
// - name 是图片名
//
// 返回值
// - vo.ImgInfosVo 是包含查询结果的 VO 对象
// - error 是可能出现的错误信息
func FindImgsByNameLike(ctx context.Context, name string) (*vo.ImgInfosVo, error) {
	if len(name) == 0 {
		return nil, errors.New("图片名不能为空")
	}

	imgInfos, err := imgInfoRepo.FindImgsByNameLike(ctx, name)
	if err != nil {
		return nil, err
	}

	// 将 po 对象转换为 vo 对象
	var imgInfosVo []vo.ImgInfoVo
	for _, imgInfo := range imgInfos {
		imgInfosVo = append(imgInfosVo, vo.ImgInfoVo{
			ImgId:   imgInfo.ImgId,
			ImgName: imgInfo.ImgName,
			ImgType: imgInfo.ImgType,
		})
	}

	return &vo.ImgInfosVo{
		Success: imgInfosVo,
	}, nil
}

// ConvertAndAddImg 添加图片并转换
// - ctx 是上下文对象，用于控制请求的生命周期
// - imgsDto 是包含图片信息的 DTO 对象
//
// 返回值
// - vo.ImgInfosVo 是包含转换成功和失败的图片信息的 VO 对象
// - error 是可能出现的错误信息
func ConvertAndAddImg(ctx context.Context, imgsDto *dto.ImgsDto) (*vo.ImgInfosVo, error) {
	if !webp.Converter.IsEmpty() {
		return nil, fmt.Errorf("转换器中还有未完成的任务")
	}

	err := webp.Converter.AddBatchTasks(ctx, imgsDto.Imgs)
	if err != nil {
		return nil, err
	}

	// 获取转换器中的输出通道
	outputCh := webp.Converter.GetOutputCh()

	// 用于保存以下几种数据
	// - 转换成功
	// - 转换失败
	// - 上传失败
	var imgPos []po.ImgInfo

	var successImgsVo []vo.ImgInfoVo
	var failImgsVo []vo.ImgInfoVo

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

				imgPo := po.ImgInfo{
					ImgId:   imgId,
					ImgName: data.ImgDto.ImgName,
				}

				if data.Flag { // 转换成功，存入数据库
					// 构建 po 对象
					imgPo.ImgType = ossstore.Webp
					// 将转换成功的数据暂存到 imgPos 中
					imgPos = append(imgPos, imgPo)
					// 将转换成功的图片信息存入 successImgsVo 中
					successImgsVo = append(successImgsVo, vo.ImgInfoVo{
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
						failImgsVo = append(failImgsVo, vo.ImgInfoVo{
							ImgId:   imgId,
							ImgName: data.ImgDto.ImgName,
							Err:     data.Err.Error(),
						})
					default:
						// 其他错误，包括下载失败和删除失败，直接返回
						// TODO: 这里有个问题，删除失败的话，Oss 中会同时存在原格式以及 webp 格式的图片，以后再添加功能自动扫描
						failImgsVo = append(failImgsVo, vo.ImgInfoVo{
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
func handleConvertedImgsData(ctx context.Context, imgPos []po.ImgInfo, successImgsVo, failImgsVo []vo.ImgInfoVo) (*vo.ImgInfosVo, error) {
	// 图片 vo 对象，包含压缩成功的和未成功的
	imgInfosVo := &vo.ImgInfosVo{}
	// 保存数据到数据库
	_, err := imgInfoRepo.AddImgInfoBatch(ctx, imgPos)
	if err != nil {
		return imgInfosVo, err
	}
	// 将成功和失败的数据返回
	imgInfosVo.Success = successImgsVo
	imgInfosVo.Failure = failImgsVo
	return imgInfosVo, nil
}

// DeleteImgs 批量删除图片
// - ctx: 上下文
// - imgIds: 图片 ID 数组
//
// - 返回值
// - vo.ImgInfosVo: 包含压缩成功的和未成功的图片信息
// - error: 错误信息
func DeleteImgs(ctx context.Context, imgIds []string) (*vo.ImgInfosVo, error) {
	// 用于保存删除成功和失败的数据
	var successImgsVo []vo.ImgInfoVo
	var failImgsVo []vo.ImgInfoVo
	// 等待批量删除的 id
	var deleteIds []string

	// 遍历 imgIds，获取到图片名称，将 Oss 中的图片删除掉
	for _, imgId := range imgIds {
		// 根据 id 查询图片信息
		imgPo, err := imgInfoRepo.FindImgById(ctx, imgId)

		// 找不到只能返回 id
		if err != nil {
			failImgsVo = append(failImgsVo, vo.ImgInfoVo{
				ImgId: imgId,
			})
			continue
		}

		// 找到了，则执行删除操作
		if imgPo != nil {
			// 从 Oss 中删除该图片
			ossPath := ossstore.GenOssSavePath(imgPo.ImgName, imgPo.ImgType)
			err = storage.Storage.DeleteObject(ctx, ossPath)

			if err != nil { // Oss 删除失败
				failImgsVo = append(failImgsVo, vo.ImgInfoVo{
					ImgId: imgId,
				})
			} else { // Oss 删除成功，则将 id 添加到待删除的数组中
				successImgsVo = append(successImgsVo, vo.ImgInfoVo{
					ImgId:   imgId,
					ImgName: imgPo.ImgName,
				})
				deleteIds = append(deleteIds, imgId)
			}
		}
	}

	// 执行批量删除操作
	_, err := imgInfoRepo.DeleteImgInfoBatch(ctx, deleteIds)
	if err != nil {
		return nil, err
	}

	return &vo.ImgInfosVo{
		Success: successImgsVo,
		Failure: failImgsVo,
	}, nil
}

// RenameImgs 重命名图片
// - ctx: 上下文
// - imgId: 图片 ID
// - newName: 新的图片名称
//
// 返回值:
// - vo.ImgInfosVo: 成功重命名的图片信息
// - error: 错误信息
func RenameImgs(ctx context.Context, imgId string, newName string) (*vo.ImgInfoVo, error) {
	if len(newName) == 0 {
		return nil, errors.New("新名称不能为空")
	}

	// 根据 id 查询图片信息
	imgPo, err := imgInfoRepo.FindImgById(ctx, imgId)
	if err != nil {
		return nil, err
	}

	// 更新 Oss 中的图片名称
	if imgPo != nil {
		// 生成新的 Oss 路径
		newOssPath := ossstore.GenOssSavePath(newName, imgPo.ImgType)
		// 生成旧的 Oss 路径
		oldOssPath := ossstore.GenOssSavePath(imgPo.ImgName, imgPo.ImgType)
		// 重命名 Oss 中的图片
		err = storage.Storage.RenameObject(ctx, oldOssPath, newOssPath)
	} else {
		return nil, errors.New("图片不存在")
	}

	// Oss 重命名失败
	if err != nil {
		return nil, err
	}

	// 更新数据库中的图片信息
	imgPo.ImgName = newName
	_, err = imgInfoRepo.UpdateImgNameById(ctx, imgId, newName)
	if err != nil {
		return nil, err
	}

	// 返回图片信息
	return &vo.ImgInfoVo{
		ImgId:   imgId,
		ImgName: newName,
	}, nil
}
