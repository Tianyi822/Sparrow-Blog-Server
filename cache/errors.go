package cache

import (
	"errors"
	"fmt"
)

// 缓存操作的错误类型
var (
	// ErrTypeMismatch 在类型转换失败时返回
	ErrTypeMismatch = errors.New("type mismatch")

	// ErrOutOfRange 在数值超出目标类型范围时返回
	ErrOutOfRange = errors.New("value out of range")

	// ErrPointerNotAllowed 在尝试存储指针值时返回
	ErrPointerNotAllowed = errors.New("pointer values are not allowed")

	// ErrNotFound 在条目不存在或已过期时返回
	ErrNotFound = errors.New("entry not found")

	// ErrEmptyKey 在键为空或仅包含空格时返回
	ErrEmptyKey = errors.New("key is empty")
)

// ErrorCode 表示错误类型的枚举值
type ErrorCode int

// 定义所有可能的错误代码
const (
	CodeUnknown ErrorCode = iota
	CodeTypeMismatch
	CodeOutOfRange
	CodePointerNotAllowed
	CodeNotFound
	CodeEmptyKey
)

// CustomCacheError 是自定义错误类型，包含错误代码和错误信息
type CustomCacheError struct {
	Code    ErrorCode // 错误代码
	Message string    // 错误信息
	Err     error     // 原始错误（可选）
}

// Error 实现了error接口
func (e *CustomCacheError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[错误码:%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[错误码:%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误，支持errors.Is和errors.As
func (e *CustomCacheError) Unwrap() error {
	return e.Err
}

// NewTypeMismatchError 创建各种错误类型的工厂函数
func NewTypeMismatchError(msg string) error {
	return &CustomCacheError{
		Code:    CodeTypeMismatch,
		Message: msg,
		Err:     ErrTypeMismatch,
	}
}

func NewOutOfRangeError(msg string) error {
	return &CustomCacheError{
		Code:    CodeOutOfRange,
		Message: msg,
		Err:     ErrOutOfRange,
	}
}

func NewPointerNotAllowedError(msg string) error {
	return &CustomCacheError{
		Code:    CodePointerNotAllowed,
		Message: msg,
		Err:     ErrPointerNotAllowed,
	}
}

func NewNotFoundError(msg string) error {
	return &CustomCacheError{
		Code:    CodeNotFound,
		Message: msg,
		Err:     ErrNotFound,
	}
}

func NewEmptyKeyError(msg string) error {
	return &CustomCacheError{
		Code:    CodeEmptyKey,
		Message: msg,
		Err:     ErrEmptyKey,
	}
}

// WrapError 可以包装任何错误类型
func WrapError(code ErrorCode, msg string, err error) error {
	return &CustomCacheError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}
