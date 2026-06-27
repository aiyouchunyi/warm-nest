// Package rands @Author larry
// File uuid_utils.go
// @Date 2024/9/25 16:34:00
// @Desc
package rands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	strings2 "strings"

	"github.com/google/uuid"

	"warm-nest/pkg/utils/strings"
)

var shortChars = []rune("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// PrefixId8 生成8位短ID，带前缀
func PrefixId8(prefix string) string {
	return prefix + Id8()
}

// PrefixId16 生成16位短ID，带前缀
func PrefixId16(prefix string) string {
	return prefix + Id16()
}

// Id4 生成4位短ID
func Id4() string {
	id32 := Id32()
	var shortBuffer strings2.Builder
	for i := 0; i < 4; i++ {
		str := id32[i*8 : i*8+8]
		x, _ := hex.DecodeString(str)
		shortBuffer.WriteRune(shortChars[int(x[0])%len(shortChars)])
	}
	return shortBuffer.String()
}

// Id8 生成8位短ID
func Id8() string {
	id32 := Id32()
	var shortBuffer strings2.Builder
	for i := 0; i < 8; i++ {
		str := id32[i*4 : i*4+4]
		x, _ := hex.DecodeString(str)
		shortBuffer.WriteRune(shortChars[int(x[0])%len(shortChars)])
	}
	return shortBuffer.String()
}

// Id16 生成16位短ID
func Id16() string {
	id32 := Id32()
	var shortBuffer strings2.Builder
	for i := 0; i < 16; i++ {
		var str string
		if i < 15 {
			str = id32[i*2 : i*2+4]
		} else {
			str = id32[i*2 : i*2+2]
		}
		x, _ := hex.DecodeString(str)
		shortBuffer.WriteRune(shortChars[int(x[0])%len(shortChars)])
	}
	return shortBuffer.String()
}

func Id32() string {
	return strings2.ReplaceAll(uuid.NewString(), "-", "")
}

// Sha256 生成sha256哈希值
func Sha256(input ...string) string {
	inputs := strings.Join("", input...)
	// 创建一个新的 SHA-256 哈希
	hash := sha256.New()
	// 写入字节数据
	hash.Write([]byte(inputs))
	// 计算并返回 16 进制编码的哈希值
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// UUID 判断字符串是否是合法的UUID
func UUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
