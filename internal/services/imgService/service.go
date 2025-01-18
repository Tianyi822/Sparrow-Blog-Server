package imgService

import (
	"context"
	"fmt"
	"h2blog/internal/model/dto"
	"h2blog/internal/model/po"
	"h2blog/internal/model/vo"
	"h2blog/internal/repository/imgInfoRepo"
	"h2blog/pkg/logger"
	"h2blog/pkg/utils"
	"h2blog/pkg/webp"
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
func ConvertAndAddImg(ctx context.Context, imgDtos []dto.ImgDto) (vo.ImgInfosVo, error) {
	// 图片 vo 对象，包含压缩成功的和未成功的
	var imgInfosVo vo.ImgInfosVo

	if !webp.Converter.IsEmpty() {
		return imgInfosVo, fmt.Errorf("转换器中还有未完成的任务")
	}

	err := webp.Converter.AddBatchTasks(ctx, imgDtos)
	if err != nil {
		return imgInfosVo, err
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
			break
		case completedMsg, ok := <-outputCh:
			if ok {
				imgId := genImgId(completedMsg.ImgDto.ImgName)
				if completedMsg.Flag { // 转换成功，存入数据库
					imgPos = append(imgPos, po.ImgInfo{
						ImgId:   imgId,
						ImgName: completedMsg.ImgDto.ImgName,
					})
					successImgsVo = append(successImgsVo, vo.ImgInfoVo{
						ImgId:   imgId,
						ImgName: completedMsg.ImgDto.ImgName,
					})
				} else { // 转换失败，存入失败列表
					failImgsVo = append(failImgsVo, vo.ImgInfoVo{
						ImgName: completedMsg.ImgDto.ImgName,
					})
				}
			} else { // 通道关闭
				// 保存数据到数据库
				_, err := imgInfoRepo.AddImgInfoBatch(ctx, imgPos)
				if err != nil {
					return imgInfosVo, err
				}
				// 将成功和失败的数据返回
				imgInfosVo.Success = successImgsVo
				imgInfosVo.Fail = failImgsVo
				return imgInfosVo, nil
			}
		case <-webp.Converter.GetCompletionStatus():
			// 保存数据到数据库
			_, err := imgInfoRepo.AddImgInfoBatch(ctx, imgPos)
			if err != nil {
				return imgInfosVo, err
			}
			// 将成功和失败的数据返回
			imgInfosVo.Success = successImgsVo
			imgInfosVo.Fail = failImgsVo
			return imgInfosVo, nil
		}
	}
}
