// Package rands @Author larry
// File rand_utils.go
// @Date 2024/5/6 20:15:00
// @Desc
package rands

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"math/big"
	"strconv"

	"github.com/sirupsen/logrus"
)

func Random16() string {
	return RandomN(16)
}

func Random32() string {
	return RandomN(32)
}

// RandomN 字母（区分大小写）与数字的组合，可以是纯字母、纯数字且长度要在N位。
func RandomN(num int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, num)
	for i := range bytes {

		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			logrus.WithError(err).Errorf("rand.Int failed!")
			randomIndex = big.NewInt(0)
		}
		bytes[i] = letters[randomIndex.Int64()]
	}
	return string(bytes)
}

// RandomNStr returns an unique id
// num为偶数则有num位，num为奇数则为num-1位
func RandomNStr(num int) string {
	for {
		id := make([]byte, num/2)
		if _, err := io.ReadFull(rand.Reader, id); err != nil {
			panic(err) // This shouldn't happen
		}
		value := hex.EncodeToString(id)
		if _, err := strconv.ParseInt(TruncateID(value), 10, 64); err == nil {
			continue
		}
		return value
	}
}

func TruncateID(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return id[:shortLen]
}
