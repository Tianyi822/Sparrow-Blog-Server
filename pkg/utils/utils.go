package utils

import (
	"encoding/hex"
	"github.com/google/uuid"
	"golang.org/x/crypto/blake2b"
)

// GenUUID 生成一个随机的 UUID
func GenUUID() string {
	return uuid.New().String()
}

// GenShortUUID 生成一个短 UUID (去掉横线)
func GenShortUUID() string {
	return uuid.New().String()[:8]
}

// IsValidUUID 验证UUID是否有效
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
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
