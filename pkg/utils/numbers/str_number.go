// Package numbers @Author larry
// File str_number.go
// @Date 2024/8/20 17:21:00
// @Desc 字符串转换为数字
package numbers

import (
	"fmt"
	strings2 "strings"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// NewT 任意类型转字符串
func NewT(data interface{}) string {
	switch v := data.(type) {
	case float32:
		return New(float64(v))
	case float64:
		return New(v)
	case uint:
		return NewFromUint64(uint64(v))
	case uint8:
		return NewFromUint64(uint64(v))
	case uint16:
		return NewFromUint64(uint64(v))
	case uint32:
		return NewFromUint64(uint64(v))
	case uint64:
		return NewFromUint64(v)
	case int:
		return NewFromInt64(int64(v))
	case int8:
		return NewFromInt64(int64(v))
	case int16:
		return NewFromInt64(int64(v))
	case int32:
		return NewFromInt64(int64(v))
	case int64:
		return NewFromInt64(v)
	case decimal.Decimal:
		return v.String()
	case string:
		if d, err := Decimal(v); err == nil {
			return d.String()
		} else {
			logrus.WithField("data", data).WithError(err).Warn("NewT failed!")
		}
	}
	return ""
}

// New float64转字符串
func New(data float64) string {
	return decimal.NewFromFloat(data).String()
}

// NewFromUint64  uint64转字符串
func NewFromUint64(data uint64) string {
	return decimal.NewFromInt(int64(data)).String()
}

// NewFromInt64 int64转字符串
func NewFromInt64(data int64) string {
	return decimal.NewFromInt(data).String()
}

// MustFloat64 字符串转换为float64
func MustFloat64(str string) float64 {
	d, err := Float64(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustFloat64 failed!")
		return 0
	}
	return d
}

// Float64 字符串转换为float64
func Float64(str string) (float64, error) {
	d, err := Decimal(str)
	if err != nil {
		return 0, err
	}
	return d.InexactFloat64(), nil
}

func MustDecimal(str string) decimal.Decimal {
	d, err := Decimal(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustDecimal failed!")
	}
	return d
}

// Decimal 字符串转换为decimal
func Decimal(str string) (decimal.Decimal, error) {
	str = strings2.TrimSpace(str)
	if str == "" {
		return decimal.Zero, nil
	}
	d, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Zero, fmt.Errorf("decimal.NewFromString(%s) failed", str)
	}
	return d, nil
}

func MustInt64(str string) int64 {
	d, err := Int64(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustInt64 failed!")
	}
	return d
}

func MustInt64T(v any) int64 {
	return MustInt64(NewT(v))
}

func Int64(str string) (int64, error) {
	d, err := Decimal(str)
	if err != nil {
		return 0, err
	}
	return d.IntPart(), nil
}

// MustUint64 字符串转换为uint64
func MustUint64(str string) uint64 {
	d, err := Uint64(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustUint64 failed!")
	}
	return d
}

func Int(str string) (int, error) {
	d, err := Decimal(str)
	if err != nil {
		return 0, err
	}
	return int(d.IntPart()), nil
}

func MustInt(str string) int {
	d, err := Int(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustInt failed!")
	}
	return d
}

func MustIntT(v any) int {
	return MustInt(NewT(v))
}

func Int32(str string) (int32, error) {
	d, err := Int(str)
	if err != nil {
		return 0, err
	}
	return int32(d), nil
}

func MustInt32(str string) int32 {
	d, err := Int32(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustInt32 failed!")
	}
	return d
}

func MustInt32T(v any) int32 {
	return MustInt32(NewT(v))
}

func IsInt(str string) bool {
	_, err := Int(str)
	return err == nil
}

// Uint64 字符串转换为uint64
func Uint64(str string) (uint64, error) {
	d, err := Decimal(str)
	if err != nil {
		return 0, err
	}
	return uint64(d.IntPart()), nil
}

// IsZeroT 判断decimal字符串是否为0
func IsZeroT(data interface{}) bool {
	return IsZero(NewT(data))
}
