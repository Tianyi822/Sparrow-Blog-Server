package utils

import (
	"encoding/hex"
	"fmt"
	"math/rand"

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

// HashWithLength 对输入字符串进行 BLAKE2b-512 哈希，并返回指定长度的十六进制字符串。
// - input: 原字符串
// - length: 要求输出的十六进制长度 (最大 128, 因为 512bit = 64byte = 128 hex chars)
func HashWithLength(input string, length int) (string, error) {
	// 1. 创建一个 BLAKE2b-512 哈希对象
	//    第二个参数可以传密钥(nil表示无密钥, 即纯哈希), 第三个参数可指定hash长度(此处先拿到全量64字节, 再手动截断)
	hasher, err := blake2b.New512(nil)
	if err != nil {
		return "", err
	}
	// 2. 写入原字符串
	_, err = hasher.Write([]byte(input))
	if err != nil {
		return "", err
	}
	// 3. 拿到 64 字节(512 bit)的哈希值
	fullSum := hasher.Sum(nil) // []byte, 长度 64

	// 4. Hex 编码 => 最长可得 128 个 hex 字符
	fullHex := hex.EncodeToString(fullSum) // string, 长度 128

	// 5. 截断输出长度, 避免越界; 如果用户要求太长, 就限制在 128
	if length < 1 {
		length = 1
	}
	if length > 128 {
		length = 128
	}
	return fullHex[:length], nil
}

// GenRandomString generates a random string of specified length
// The string contains random ASCII characters (32-126, printable characters)
func GenRandomString(length int) string {
	// Create byte slice to store random characters
	result := make([]byte, length)

	// Get random bytes
	for i := 0; i < length; i++ {
		// Generate random number between 32 and 126 (printable ASCII characters)
		// rand.Int63() generates a non-negative 63-bit integer
		result[i] = byte(32 + rand.Int63n(95)) // 95 = 126 - 32 + 1
	}

	return string(result)
}
