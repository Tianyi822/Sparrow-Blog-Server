package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sparrow_blog_server/internal/model/dto"
	"sparrow_blog_server/routers/resp"
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

// GetBlogDto 从 gin.Context 中提取数据并将其转换为 BlogDto 对象。
// 参数:
//   - ctx: *gin.Context，表示当前的 HTTP 请求上下文，包含请求数据和响应方法。
//
// 返回值:
//   - *dto.BlogDto: 转换后的 BlogDto 对象，如果转换失败则返回 nil。
//   - error: 如果在数据转换过程中发生错误，则返回该错误；否则返回 nil。
func GetBlogDto(ctx *gin.Context) (*dto.BlogDto, error) {
	// 初始化一个 BlogDto 对象，用于存储从请求中提取的数据。
	blogDto := &dto.BlogDto{}

	// 调用 rowDataToDto 函数将请求数据映射到 blogDto 对象中。
	// 如果映射过程中发生错误，返回错误信息并通过 resp.BadRequest 响应客户端。
	err := rowDataToDto(ctx, blogDto)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	// 如果数据转换成功，返回填充好的 BlogDto 对象和 nil 错误。
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

// GetImgDtos 从 HTTP 请求中解析 RawData 数据并将其转换为 ImgsDto 对象
func GetImgDtos(ctx *gin.Context) (*dto.ImgsDto, error) {
	imgDtos := &dto.ImgsDto{}

	err := rowDataToDto(ctx, imgDtos)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	return imgDtos, nil
}

// GetFriendLinkDto 从 gin.Context 中提取数据并将其转换为 FriendLinkDto 对象。
// 参数:
//   - ctx: *gin.Context，表示当前的 HTTP 请求上下文，包含请求数据和响应方法。
//
// 返回值:
//   - *dto.FriendLinkDto: 转换后的 FriendLinkDto 对象，如果转换失败则返回 nil。
//   - error: 如果在数据转换过程中发生错误，则返回该错误；否则返回 nil。
func GetFriendLinkDto(ctx *gin.Context) (*dto.FriendLinkDto, error) {
	// 初始化一个 FriendLinkDto 对象，用于存储从请求中提取的数据。
	friendLinkDto := &dto.FriendLinkDto{}

	// 调用 rowDataToDto 函数将请求数据映射到 friendLinkDto 对象中。
	// 如果映射过程中发生错误，返回错误信息并通过 resp.BadRequest 响应客户端。
	err := rowDataToDto(ctx, friendLinkDto)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	// 如果数据转换成功，返回填充好的 FriendLinkDto 对象和 nil 错误。
	return friendLinkDto, nil
}

// GetCommentDto 从 gin.Context 中提取数据并将其转换为 CommentDto 对象。
// 参数:
//   - ctx: *gin.Context，表示当前的 HTTP 请求上下文，包含请求数据和响应方法。
//
// 返回值:
//   - *dto.CommentDto: 转换后的 CommentDto 对象，如果转换失败则返回 nil。
//   - error: 如果在数据转换过程中发生错误，则返回该错误；否则返回 nil。
func GetCommentDto(ctx *gin.Context) (*dto.CommentDto, error) {
	// 初始化一个 CommentDto 对象，用于存储从请求中提取的数据。
	commentDto := &dto.CommentDto{}

	// 调用 rowDataToDto 函数将请求数据映射到 commentDto 对象中。
	// 如果映射过程中发生错误，返回错误信息并通过 resp.BadRequest 响应客户端。
	err := rowDataToDto(ctx, commentDto)
	if err != nil {
		resp.BadRequest(ctx, err.Error(), -1)
		return nil, err
	}

	// 如果数据转换成功，返回填充好的 CommentDto 对象和 nil 错误。
	return commentDto, nil
}

// rowDataToDto 将 HTTP 请求中的原始数据解析为指定的 DTO（数据传输对象）。
// 参数:
//   - ctx: *gin.Context，表示当前的 HTTP 请求上下文，包含请求方法、路径和原始数据等信息。
//   - dto: dto.Dto，表示目标数据传输对象，用于存储解析后的数据。
//
// 返回值:
//   - error: 如果解析过程中发生错误（如获取原始数据失败或 JSON 解析失败），返回相应的错误信息；否则返回 nil。
func rowDataToDto(ctx *gin.Context, dto dto.Dto) error {
	if ctx.Request.Method != "GET" {
		// 获取请求体中的原始数据 (RawData)。
		rowData, err := ctx.GetRawData()
		// 如果获取原始数据失败，记录详细的错误信息并返回。
		if err != nil {
			msg := fmt.Sprintf(
				"%s => %s 请求获取 RawData 失败: %s",
				ctx.Request.Method,
				ctx.Request.URL.Path,
				err.Error(),
			)
			return errors.New(msg)
		}

		// 将原始数据解析为指定的 DTO 对象。
		// 如果解析失败，记录详细的错误信息并返回。
		err = json.Unmarshal(rowData, dto)
		if err != nil {
			msg := fmt.Sprintf("%s => %s 请求解析失败，请检查数据格式是否符合待转换值类型，err: %s", ctx.Request.Method, ctx.Request.URL.Path, err.Error())
			return errors.New(msg)
		}
	}

	return nil
}

// GetStringFromRawData 从原始数据中提取指定键的值并将其转换为字符串类型。
// 参数:
//   - rawData: 包含键值对的 map，值可以是任意类型。
//   - key: 要提取的键名。
//
// 返回值:
//   - string: 提取并转换成功的字符串值。
//   - error: 如果键不存在或值类型不支持，则返回相应的错误信息。
func GetStringFromRawData(rawData map[string]any, key string) (string, error) {
	// 检查键是否存在，如果不存在则返回错误
	val, ok := rawData[key]
	if !ok {
		return "", fmt.Errorf("'%s' 字段不存在", key)
	}

	// 根据值的实际类型进行处理
	switch v := val.(type) {
	case string:
		// 如果值是字符串类型，直接返回
		return v, nil
	case float64:
		// 处理指数表示法的浮点数
		// 使用 FormatFloat 将浮点数转换为字符串，'f' 表示普通十进制格式，-1 表示保留所有有效数字
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		// 处理整数，将其转换为字符串
		return strconv.Itoa(v), nil
	default:
		// 如果值的类型不被支持，返回错误
		return "", fmt.Errorf("'%s' 类型不支持", key)
	}
}

// GetUInt16FromRawData 从原始数据中提取指定键的值并将其转换为 uint16 类型。
// 参数:
//   - reqData: 包含键值对的 map，值可以是任意类型。
//   - key: 要提取的键名。
//
// 返回值:
//   - uint16: 提取并转换成功的 uint16 值。
//   - error: 如果键不存在、值类型不支持或值超出 uint16 范围，则返回相应的错误信息。
func GetUInt16FromRawData(reqData map[string]any, key string) (uint16, error) {
	// 检查键是否存在，如果不存在则返回错误。
	val, ok := reqData[key]
	if !ok {
		return 0, fmt.Errorf("'%s' 字段不存在", key)
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

// GetUInt8FromRawData 从原始数据中提取指定键的值并将其转换为 uint8 类型。
// 参数:
//   - reqData: 包含键值对的 map，值可以是任意类型。
//   - key: 要提取的键名。
//
// 返回值:
//   - uint8: 提取并转换成功的 uint8 值。
//   - error: 如果键不存在、值类型不支持或值超出 uint8 范围，则返回相应的错误信息。
func GetUInt8FromRawData(reqData map[string]any, key string) (uint8, error) {
	// 检查键是否存在，如果不存在则返回错误
	val, ok := reqData[key]
	if !ok {
		return 0, fmt.Errorf("'%s' 字段不存在", key)
	}

	// 根据值的实际类型进行处理
	switch v := val.(type) {
	case float64:
		// 如果值是 float64 类型，检查是否在 uint8 范围内
		if v < 0 || v > math.MaxUint8 {
			return 0, fmt.Errorf("'%s' 超出 uint8 范围", key)
		}
		return uint8(v), nil
	case int:
		// 如果值是 int 类型，检查是否在 uint8 范围内
		if v < 0 || v > math.MaxUint8 {
			return 0, fmt.Errorf("'%s' 超出 uint8 范围", key)
		}
		return uint8(v), nil
	case string:
		// 如果值是字符串类型，尝试将其解析为无符号整数
		num, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return 0, fmt.Errorf("不可使用的无符号整型值 '%s': %w", key, err)
		}
		// 检查解析后的值是否在 uint8 范围内
		if num > math.MaxUint8 {
			return 0, fmt.Errorf("'%s' 超出 uint8 范围", key)
		}
		return uint8(num), nil
	default:
		// 如果值的类型不被支持，返回错误
		return 0, fmt.Errorf("'%s' 类型不支持", key)
	}
}

// GetFloatFromRawData 从原始数据中提取指定键的浮点数值。
// 参数:
//   - reqData: 包含键值对的原始数据映射，值可以是任意类型。
//   - key: 需要提取浮点数值的键。
//
// 返回值:
//   - float32: 提取并转换成功的浮点数值。
//   - error: 如果键不存在、值类型不支持或转换失败，则返回相应的错误信息。
func GetFloatFromRawData(reqData map[string]any, key string) (float32, error) {
	// 检查键是否存在，如果不存在则返回错误。
	val, ok := reqData[key]
	if !ok {
		return 0, fmt.Errorf("'%s' 字段不存在", key)
	}

	// 根据值的实际类型进行处理。
	switch v := val.(type) {
	case float64:
		// 如果值是 float64 类型，直接转换为 float32。
		return float32(v), nil
	case int:
		// 如果值是 int 类型，转换为 float32。
		return float32(v), nil
	case string:
		// 如果值是字符串类型，尝试将其解析为浮点数。
		num, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return 0, fmt.Errorf("不可使用的浮点值 '%s': %w", key, err)
		}
		return float32(num), nil
	default:
		// 如果值的类型不被支持，返回错误。
		return 0, fmt.Errorf("'%s' 类型不支持", key)
	}
}

// GetBoolFromRawData 从原始数据中提取指定键的布尔值。
// 参数:
//   - reqData: 包含键值对的原始数据映射，值可以是任意类型。
//   - key: 需要提取布尔值的键名。
//
// 返回值:
//   - bool: 提取并转换成功的布尔值。
//   - error: 如果键不存在、值类型不支持或转换失败，则返回相应的错误信息。
func GetBoolFromRawData(reqData map[string]any, key string) (bool, error) {
	val, ok := reqData[key]
	if !ok {
		return false, fmt.Errorf("'%s' 字段不存在", key)
	}
	switch v := val.(type) {
	case bool:
		return v, nil
	case string:
		// 尝试将字符串转换为布尔值
		num, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("不可使用的布尔值 '%s': %w", key, err)
		}
		return num, nil
	default:
		// 如果值的类型不被支持，返回错误
		return false, fmt.Errorf("'%s' 类型不支持", key)
	}
}

// GetStrListFromRawData 从原始数据中提取字符串列表
// 参数:
//   - reqData: 包含键值对的原始数据映射，值可以是任意类型
//   - key: 需要提取字符串列表的键名
//
// 返回值:
//   - []string: 提取并转换成功的字符串列表
//   - error: 如果键不存在、值类型不支持或转换失败，则返回相应的错误信息
func GetStrListFromRawData(reqData map[string]any, key string) ([]string, error) {
	// 检查键是否存在，如果不存在则返回错误
	val, ok := reqData[key]
	if !ok {
		return nil, fmt.Errorf("'%s' 字段不存在", key)
	}

	switch v := val.(type) {
	case []string:
		// 如果值已经是字符串切片类型，直接返回
		return v, nil
	case []any:
		// 如果值是任意类型的切片，尝试将每个元素转换为字符串
		var list []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				list = append(list, str)
			} else {
				// 如果有任何元素不能转换为字符串，返回错误
				return nil, fmt.Errorf("'%s' 类型不支持", key)
			}
		}
		return list, nil
	default:
		// 如果值的类型既不是字符串切片也不是任意类型切片，返回错误
		return nil, fmt.Errorf("'%s' 类型不支持", key)
	}
}
