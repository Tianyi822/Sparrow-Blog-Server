package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"h2blog/internal/model/dto"
	"h2blog/pkg/logger"
	"h2blog/pkg/resp"

	"github.com/gin-gonic/gin"
)

// RowDataToMap 将请求的参数转换成 map[string]string
// 在没有明确指定 DTO 的情况下，可以调用这个方法
func RowDataToMap(ctx *gin.Context) (map[string]string, error) {
	reqData := make(map[string]string)

	if ctx.Request.Method != "GET" && ctx.Request.Method != "DELETE" {
		// 拿到 RawData 数据
		rowData, err := ctx.GetRawData()
		// 处理获取 rawData 数据失败的情况
		if err != nil {
			msg := fmt.Sprintf(
				"%s => %s 请求获取 RawData 失败: %s",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				err.Error(),
			)
			logger.Error(msg)
			resp.BadRequest(ctx, msg, -1)
		}

		err = json.Unmarshal(rowData, &reqData)
		if err != nil {
			msg := fmt.Sprintf(
				"%s => %s 请求解析失败，请检查数据格式是否正确(kv 都是字符串)，err: %s",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				err.Error(),
			)
			logger.Error(msg)
			resp.BadRequest(ctx, msg, -1)
			return nil, err
		}
	}

	return reqData, nil
}

// GetBlogDto 从请求中获取 BlogInfoDto 对象
func GetBlogDto(ctx *gin.Context) (*dto.BlogInfoDto, error) {
	// 初始化一个BlogInfoDto对象
	blogDto := &dto.BlogInfoDto{}

	// 解析失败就不再继续处理
	err := rowDataToDto(ctx, blogDto)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	return blogDto, nil
}

// GetImgsDto 从请求中获取 ImgsDto 对象
func GetImgsDto(ctx *gin.Context) (*dto.ImgsDto, error) {
	// 初始化一个ImgsDto对象
	imgsDto := &dto.ImgsDto{}

	err := rowDataToDto(ctx, imgsDto)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	return imgsDto, nil
}

// RowDataToDto 将请求的参数转换成 DTO
func rowDataToDto(ctx *gin.Context, dto dto.Dto) error {
	if ctx.Request.Method != "GET" {
		// 拿到 RawData 数据
		rowData, err := ctx.GetRawData()
		// 处理获取 rawData 数据失败的情况
		if err != nil {
			msg := fmt.Sprintf(
				"%s => %s 请求获取 RawData 失败: %s",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				err.Error(),
			)
			logger.Error(msg)
			return errors.New(msg)
		}

		err = json.Unmarshal(rowData, dto)
		if err != nil {
			msg := fmt.Sprintf("%s => %s 请求解析失败，请检查数据格式是否符合待转换值类型，err: %s", ctx.Request.Method, ctx.Request.URL.Path, err.Error())
			logger.Error(msg)
			return errors.New(msg)
		}
	}

	return nil
}
