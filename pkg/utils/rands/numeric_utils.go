// Package rands @Author larry
// @Date 2026/3/30 00:00
// @Desc 生成随机数字串
package rands

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

//┌────────┬────────┬──────────┐
//│ 总量 n │ 位数 d │ 碰撞概率 │
//├────────┼────────┼──────────┤
//│ 1 亿   │ 20     │ 5×10⁻⁵   │
//├────────┼────────┼──────────┤
//│ 1 亿   │ 24     │ 5×10⁻⁹   │
//├────────┼────────┼──────────┤
//│ 10 亿  │ 24     │ 5×10⁻⁷   │
//└────────┴────────┴──────────┘

// Numeric 生成 18 位随机数字串（默认长度）
func Numeric() string {
	return NumericN(18)
}

// NumericN 生成指定长度的随机数字串（n: 数字串长度，最小为 1，返回: 纯数字字符串）
func NumericN(n int) string {
	if n < 1 {
		n = 1
	}

	const digits = "0123456789"
	var builder strings.Builder
	builder.Grow(n)

	for i := 0; i < n; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			logrus.WithError(err).Errorf("rand.Int failed!")
			randomIndex = big.NewInt(0)
		}
		builder.WriteByte(digits[randomIndex.Int64()])
	}

	return builder.String()
}

// TimestampNumeric 生成毫秒时间戳 + 随机数字串（默认总长度 18 位）
//
// 格式: 13 位毫秒时间戳 + 5 位随机数 = 18 位数字串
//
// 冲突概率（同一毫秒内）:
//   - 随机空间: 10⁵ = 100,000
//   - ~400 个并发请求时冲突概率达 1%
//   - 适用场景: 中低并发 (<100 QPS)
//
// 返回值不会超过 int64 上限（19 位）
func TimestampNumeric() string {
	return TimestampNumericN(18)
}

// TimestampNumericN 生成毫秒时间戳 + 随机数字串（指定总长度）
//
// 参数:
//   - n: 总长度（最小 13，推荐 18），格式: 13 位毫秒时间戳 + (n-13) 位随机数
//   - 当 n < 13 时自动调整为 13；当 n > 19 时返回字符串（超出 int64 范围）
//
// 冲突概率（同一毫秒内）:
//   - n=18 (5 位随机): 10⁵ 空间，~400 并发达 1% 冲突
//   - n=19 (6 位随机): 10⁶ 空间，~1,250 并发达 1% 冲突
//   - n=20 (7 位随机): 10⁷ 空间，~4,000 并发达 1% 冲突
//   - 计算公式: P(冲突) ≈ n²/(2×10^随机位数)
func TimestampNumericN(n int) string {
	const timestampLen = 13
	if n < timestampLen {
		n = timestampLen
	}

	randomLen := n - timestampLen
	timestampMs := time.Now().UnixMilli()

	if randomLen == 0 {
		return fmt.Sprintf("%d", timestampMs)
	}

	return fmt.Sprintf("%d%s", timestampMs, NumericN(randomLen))
}
