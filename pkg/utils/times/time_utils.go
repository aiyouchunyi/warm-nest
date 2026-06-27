// Package times @Author larry
// File time_utils.go
// @Date 2024/6/28 09:14:00
// @Desc
package times

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// UnixMilliStr 获取当前时间戳字符串
func UnixMilliStr() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

// UnixMilli 获取当前时间戳
func UnixMilli() int64 {
	return time.Now().UnixMilli()
}

func ToUnixMilli(unixMilliStr string) int64 {
	if unixMilliStr == "" {
		return 0
	}
	unixMilli, err := strconv.ParseInt(unixMilliStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Errorf("Error converting unixMilli to int64")
		return 0
	}
	return unixMilli
}

func BeforeMilli(gap int64) int64 {
	return time.Now().UnixMilli() - gap
}

func AfterMilli(gap int64) int64 {
	return time.Now().UnixMilli() + gap
}

func UnixStr() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func Unix() int64 {
	return time.Now().Unix()
}

// Gap 获取时间差
func Gap(start int64) int64 {
	return time.Now().UnixMilli() - start
}

// GapMS 获取时间差
func GapMS(start int64) string {
	return fmt.Sprintf("%dms", time.Now().UnixMilli()-start)
}
