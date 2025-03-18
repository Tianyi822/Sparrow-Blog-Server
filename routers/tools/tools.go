package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"h2blog_server/internal/model/dto"
	"h2blog_server/pkg/resp"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMapFromRawData 将请求的参数转换成 map[string]any
// 在没有明确指定 DTO 的情况下，可以调用这个方法
// 注意！！！！
// 获取其中的数据时一定要知道到底是什么数据类型，因为没有 dto 作为解析对象，没法明确知道是什么数据
func GetMapFromRawData(ctx *gin.Context) (map[string]any, error) {
	reqData := make(map[string]any)

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
			resp.BadRequest(ctx, msg, -1)
		}

		err = json.Unmarshal(rowData, &reqData)
		if err != nil {
			msg := fmt.Sprintf(
				"%s => %s 请求解析失败，请检查数据格式是否正确，err: %s",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				err.Error(),
			)
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

// GetImgDto 从请求中获取 ImgDto 对象
func GetImgDto(ctx *gin.Context) (*dto.ImgDto, error) {
	// 初始化一个ImgDto对象
	imgDto := &dto.ImgDto{}

	err := rowDataToDto(ctx, imgDto)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	return imgDto, nil
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
			return errors.New(msg)
		}

		err = json.Unmarshal(rowData, dto)
		if err != nil {
			msg := fmt.Sprintf("%s => %s 请求解析失败，请检查数据格式是否符合待转换值类型，err: %s", ctx.Request.Method, ctx.Request.URL.Path, err.Error())
			return errors.New(msg)
		}
	}

	return nil
}

func GetIntFromPostForm(c *gin.Context, key string) (int, error) {
	value := c.PostForm(key)
	if value == "" {
		return 0, fmt.Errorf("'%s' 为空", key)
	}
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("不可使用的整型值 '%s': %w", key, err)
	}
	return num, nil
}

// GetIntFromRawData 从给定的 map 中提取指定键的值，并尝试将其转换为整型。
// 参数:
//
//	reqData - 包含键值对的 map，值可以是任意类型。
//	key     - 要提取和转换的键。
//
// 返回值:
//
//	int  - 如果成功提取并转换，则返回对应的整数值。
//	error - 如果键不存在、值类型不支持或转换失败，则返回相应的错误信息。
func GetIntFromRawData(reqData map[string]any, key string) (int, error) {
	val, ok := reqData[key]
	if !ok {
		// 如果键不存在，返回错误提示键为空
		return 0, fmt.Errorf("'%s' 为空", key)
	}

	switch v := val.(type) {
	case float64:
		// 如果值是 float64 类型，检查是否在 int32 范围内
		if v < math.MinInt32 || v > math.MaxInt32 {
			return 0, fmt.Errorf("'%s' 超出 int32 范围", key)
		}
		return int(v), nil
	case int:
		// 如果值已经是 int 类型，直接返回
		return v, nil
	case string:
		// 如果值是字符串类型，尝试将其转换为整数
		num, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("不可使用的整型值 '%s': %w", key, err)
		}
		return num, nil
	default:
		// 如果值的类型不被支持，返回错误提示
		return 0, fmt.Errorf("'%s' 类型不支持", key)
	}
}

func GetUInt16FromPostForm(c *gin.Context, key string) (uint16, error) {
	value := c.PostForm(key)
	if value == "" {
		return 0, fmt.Errorf("'%s' 为空", key)
	}
	num, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("不可使用的无符号整型值 '%s': %w", key, err)
	}
	return uint16(num), nil
}

// GetUInt16FromRawData 从原始数据中提取指定键的值并将其转换为 uint16 类型。
// 参数:
// - reqData: 包含键值对的 map，值可以是任意类型。
// - key: 要提取的键名。
// 返回值:
// - uint16: 提取并转换成功的 uint16 值。
// - error: 如果键不存在、值类型不支持或值超出 uint16 范围，则返回相应的错误信息。
func GetUInt16FromRawData(reqData map[string]any, key string) (uint16, error) {
	// 检查键是否存在，如果不存在则返回错误。
	val, ok := reqData[key]
	if !ok {
		return 0, fmt.Errorf("'%s' 为空", key)
	}

	// 根据值的实际类型进行处理。
	switch v := val.(type) {
	case float64:
		// 如果值是 float64 类型，检查是否在 uint16 范围内。
		if v < 0 || v > math.MaxUint16 {
			return 0, fmt.Errorf("'%s' 超出 uint16 范围", key)
		}
		return uint16(v), nil
	case int:
		// 如果值是 int 类型，检查是否在 uint16 范围内。
		if v < 0 || v > math.MaxUint16 {
			return 0, fmt.Errorf("'%s' 超出 uint16 范围", key)
		}
		return uint16(v), nil
	case string:
		// 如果值是字符串类型，尝试将其解析为无符号整数。
		num, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return 0, fmt.Errorf("不可使用的无符号整型值 '%s': %w", key, err)
		}
		// 检查解析后的值是否在 uint16 范围内。
		if num > math.MaxUint16 {
			return 0, fmt.Errorf("'%s' 超出 uint16 范围", key)
		}
		return uint16(num), nil
	default:
		// 如果值的类型不被支持，返回错误。
		return 0, fmt.Errorf("'%s' 类型不支持", key)
	}
}

func GetFloatFromPostForm(c *gin.Context, key string) (float32, error) {
	value := c.PostForm(key)
	if value == "" {
		return 0, fmt.Errorf("'%s' 为空", key)
	}
	num, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, fmt.Errorf("不可使用的浮点值 '%s': %w", key, err)
	}
	return float32(num), nil
}
