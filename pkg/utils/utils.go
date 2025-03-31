package utils

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/blake2b"
)

// GenId 根据名称生成一个唯一的ID。
// 参数:
//
//	name - 用于生成ID的字符串。
//
// 返回值:
//
//	生成的ID字符串和一个错误对象，如果生成失败则返回错误。
func GenId(name string) (string, error) {
	// 使用envs包的HashWithLength函数生成一个长度为16的哈希字符串作为博客ID
	str, err := HashWithLength(name, 16)
	// 检查是否生成成功，如果失败则记录错误并尝试重新生成
	if err != nil {
		// 初始化计数器，用于限制重试次数
		count := 0
		name = fmt.Sprintf("%v%d", name, count)
		// 使用for循环尝试重新生成ID，最多重试3次
		for count <= 3 && err != nil {
			str, err = HashWithLength(name, 16)
			count++
			name = fmt.Sprintf("%v%d", name, count)
		}
	}

	// 如果仍然失败，则记录错误并返回空字符串
	if err != nil {
		return "", fmt.Errorf("生成 ID 失败: %v", err)
	}

	// 返回生成的ID
	return str, nil
}

// HashWithLength 使用 BLAKE2b-512 哈希算法对输入字符串进行哈希计算，并返回指定长度的十六进制编码结果。
// 参数:
//   - input: 待哈希的字符串。
//   - length: 返回的哈希值的十六进制字符长度，范围为 1 到 128。如果超出范围，会被限制在有效范围内。
//
// 返回值:
//   - string: 截断后的十六进制编码哈希值。
//   - error: 如果在创建哈希对象或写入数据时发生错误，则返回相应的错误信息。
func HashWithLength(input string, length int) (string, error) {
	// 创建一个 BLAKE2b-512 哈希对象，使用无密钥模式。
	hashEncoder, err := blake2b.New512(nil)
	if err != nil {
		return "", err
	}

	// 将输入字符串写入哈希对象。
	_, err = hashEncoder.Write([]byte(input))
	if err != nil {
		return "", err
	}

	// 计算完整的 64 字节哈希值。
	fullSum := hashEncoder.Sum(nil)

	// 将哈希值转换为十六进制编码字符串，最大长度为 128 个字符。
	fullHex := hex.EncodeToString(fullSum)

	// 根据用户指定的长度截断十六进制字符串，确保长度在有效范围内。
	if length < 1 {
		length = 1
	}
	if length > 128 {
		length = 128
	}
	return fullHex[:length], nil
}
