package imgService

import (
	"context"
	"errors"
	"fmt"
	"h2blog/internal/model/dto"
	"h2blog/internal/model/po"
	"h2blog/internal/model/vo"
	"h2blog/internal/repository/imgInfoRepo"
	"h2blog/pkg/logger"
	"h2blog/pkg/utils"
	"h2blog/pkg/webp"
	"h2blog/storage"
	"h2blog/storage/oss"
)

// genImgId 用于生成图片的唯一标识符
//   - title 是图片的标题
//
// 返回值
//   - string 表示生成的图片ID
func genImgId(title string) string {
	// 使用envs包的HashWithLength函数生成一个长度为16的哈希字符串作为图片ID
	str, err := utils.HashWithLength(title, 16)
	// 检查是否生成成功，如果失败则记录错误并尝试重新生成
	if err != nil {
		// 使用logger包记录错误信息，包括错误详情
		logger.Error("生成图片 ID 失败: %v，准备重新生成", err)
		// 初始化计数器，用于限制重试次数
		count := 0
		title = title + fmt.Sprintf("%d", count)
		// 使用for循环尝试重新生成图片ID，最多重试3次
		for count <= 2 && err != nil {
			str, err = utils.HashWithLength(title, 16)
			count++
		}
	}
	logger.Info("生成图片 ID 成功: %s", str)
	// 返回生成的图片ID
	return str
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

	// 将转换成功的图片信息暂存在这里
	var imgPos []po.ImgInfo

	var successImgsVo []vo.ImgInfoVo
	var failImgsVo []vo.ImgInfoVo

	for {
		select {
		case <-ctx.Done():
			return handleConvertedImgsData(ctx, imgPos, successImgsVo, failImgsVo)
		case data, ok := <-outputCh:
			if ok {
				if data.Flag { // 转换成功，存入数据库
					// 生成 ID
					imgId := genImgId(data.ImgDto.ImgName)
					// 构建 po 对象
					imgPos = append(imgPos, po.ImgInfo{
						ImgId:   imgId,
						ImgName: data.ImgDto.ImgName,
						ImgType: oss.Webp,
					})
					successImgsVo = append(successImgsVo, vo.ImgInfoVo{
						ImgId:   imgId,
						ImgName: data.ImgDto.ImgName,
					})
				} else { // 转换失败，存入失败列表
					failImgsVo = append(failImgsVo, vo.ImgInfoVo{
						ImgName: data.ImgDto.ImgName,
						Err:     data.Err,
					})
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

	// 遍历 imgIds，获取到图片名称，将 OSS 中的图片删除掉
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
			// 从 OSS 中删除该图片
			ossPath := oss.GenOssSavePath(imgPo.ImgName, imgPo.ImgType)
			err = storage.Storage.DeleteObject(ctx, ossPath)

			if err != nil { // OSS 删除失败
				failImgsVo = append(failImgsVo, vo.ImgInfoVo{
					ImgId: imgId,
				})
			} else { // OSS 删除成功，则将 id 添加到待删除的数组中
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
func RenameImgs(ctx context.Context, imgId string, newName string) (vo.ImgInfoVo, error) {
	var imgInfoVo vo.ImgInfoVo

	// 根据 id 查询图片信息
	imgPo, err := imgInfoRepo.FindImgById(ctx, imgId)
	if err != nil {
		return imgInfoVo, err
	}

	// 更新 OSS 中的图片名称
	if imgPo != nil {
		// 生成新的 OSS 路径
		newOssPath := oss.GenOssSavePath(newName, imgPo.ImgType)
		// 生成旧的 OSS 路径
		oldOssPath := oss.GenOssSavePath(imgPo.ImgName, imgPo.ImgType)
		// 重命名 OSS 中的图片
		err = storage.Storage.RenameObject(ctx, oldOssPath, newOssPath)
	} else {
		return imgInfoVo, errors.New("图片不存在")
	}

	// OSS 重命名失败
	if err != nil {
		return imgInfoVo, err
	}

	// 更新数据库中的图片信息
	imgPo.ImgName = newName
	_, err = imgInfoRepo.UpdateImgNameById(ctx, imgId, newName)
	if err != nil {
		return imgInfoVo, err
	}

	// 返回图片信息
	imgInfoVo.ImgId = imgId
	imgInfoVo.ImgName = newName

	return imgInfoVo, nil
}
