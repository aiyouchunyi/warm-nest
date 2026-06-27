// Package numbers @Author larry
// @Date 2024/4/22 09:01:00
// @Desc decimal工具类
package numbers

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/strings"
)

// Format 格式化decimal字符串
// 如果字符串为空，则返回0
func Format(str string) string {
	if IsZero(str) {
		return "0"
	}
	return str
}

// Valid 检查多个decimal字符串是否有效
func Valid(strs ...string) error {
	for _, str := range strs {
		_, err := Decimal(str)
		if err != nil {
			return err
		}
	}
	return nil
}

// MustAdd 计算多个decimal字符串的和
func MustAdd(strs ...string) string {
	out, err := Add(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).WithError(err).Warnf("MustAdd failed!")
	}
	return out
}

// Add 计算多个decimal字符串的和
func Add(strs ...string) (string, error) {
	if strings.BothBlank(strs...) {
		return "", nil
	}
	var sum decimal.Decimal
	for _, str := range strs {
		d, err := Decimal(str)
		if err != nil {
			return "", err
		}
		sum = sum.Add(d)
	}
	return sum.String(), nil
}

func MustAbsAdd(strs ...string) string {
	out, err := AbsAdd(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).WithError(err).Warnf("MustAbsAdd failed!")
	}
	return out
}

// AbsAdd 计算多个decimal字符串的绝对值和
func AbsAdd(strs ...string) (string, error) {
	if strings.BothBlank(strs...) {
		return "", nil
	}
	var sum decimal.Decimal
	for _, str := range strs {
		d, err := Decimal(str)
		if err != nil {
			return "", err
		}
		sum = sum.Add(d.Abs())
	}
	return sum.String(), nil
}

// MustSub 计算多个decimal字符串的差
func MustSub(strs ...string) string {
	out, err := Sub(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).WithError(err).Warnf("MustSub failed!")
	}
	return out
}

// Sub 计算多个decimal字符串的差
// 多个字符串转decimal相减 返回字符串与error
func Sub(strs ...string) (string, error) {
	if strings.BothBlank(strs...) {
		return "", nil
	}
	var out decimal.Decimal
	for i, str := range strs {
		d, err := Decimal(str)
		if err != nil {
			return "", err
		}
		if i == 0 {
			out = d
		} else {
			out = out.Sub(d)
		}
	}
	return out.String(), nil
}

// MustDiv 计算多个decimal字符串的商
// 多个字符串转decimal相除 返回字符串与error
func MustDiv(strs ...string) string {
	out, err := Div(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).Warnf("MustDiv failed!")
	}
	return out
}

// Div 计算多个decimal字符串的商
// 多个字符串转decimal相除 返回字符串与error
func Div(strs ...string) (string, error) {
	if strings.BothBlank(strs...) {
		return "", nil
	}
	var out decimal.Decimal
	for i, str := range strs {
		d, err := Decimal(str)
		if err != nil {
			return "", err
		}
		if i == 0 {
			out = d
		} else {
			if d.IsZero() {
				return "", fmt.Errorf("divide by zero %d", i)
			}
			out = out.Div(d)
		}
	}
	return out.String(), nil
}

// MustMulti 计算多个decimal字符串的积
func MustMulti(strs ...string) string {
	out, err := Multi(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).WithError(err).Warnf("MustMulti failed!")
	}
	return out
}

// Multi 计算多个decimal字符串的积
// 多个字符串转decimal相乘 返回字符串与error
func Multi(strs ...string) (string, error) {
	if strings.BothBlank(strs...) {
		return "", nil
	}
	var out decimal.Decimal
	for i, str := range strs {
		d, err := Decimal(str)
		if err != nil {
			return "", err
		}
		if i == 0 {
			out = d
		} else {
			out = out.Mul(d)
		}
	}
	return out.String(), nil
}

func MustCompare(str1, str2 string) int {
	out, err := Compare(str1, str2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str1": str1,
			"str2": str2,
		}).WithError(err).Warnf("MustCompare failed!")
	}
	return out
}

// Compare 两个decimal字符串比较大小
func Compare(str1, str2 string) (int, error) {
	d1, err := Decimal(str1)
	if err != nil {
		return 0, err
	}
	d2, err := Decimal(str2)
	if err != nil {
		return 0, err
	}
	return d1.Cmp(d2), nil
}

// Round 字符串转decimal，保留 n 位小数
// str: 待转换的字符串(四舍五入)
func Round(str string, precision int) string {
	if strings.IsBlank(str) {
		return ""
	}
	d, err := Decimal(str)
	if err != nil {
		return str
	}
	return d.Round(int32(precision)).String()
}

// RoundCeil 字符串转decimal，保留 n 位小数
// str: 待转换的字符串(向上取整)
func RoundCeil(str string, precision int) string {
	if strings.IsBlank(str) {
		return ""
	}
	d, err := Decimal(str)
	if err != nil {
		return str
	}
	return d.RoundCeil(int32(precision)).String()
}

func Truncate(str string, precision int) string {
	if strings.IsBlank(str) {
		return ""
	}
	d, err := Decimal(str)
	if err != nil {
		return str
	}
	return d.Truncate(int32(precision)).String()
}

// FormatRound 字符串转decimal，保留 n 位小数
func FormatRound(str string, precision int) string {
	return Format(Round(str, precision))
}

func Floor(s string, precision int) string {
	words := strings.Split(s, ".")
	if len(words) == 2 {
		if len(words[1]) > precision {
			if len(words[1]) >= 8 && words[1][len(words[1])-6:len(words[1])-1] == "99999" {
				ds, _ := decimal.NewFromString(s)
				return ds.StringFixed(int32(precision))
			}

			if precision > 0 {
				s = strings.Join(".", words[0], words[1][:precision])
			} else {
				s = words[0]
			}
		}
	}
	ds, _ := decimal.NewFromString(s)
	if ds.Equal(decimal.Zero) {
		return "0"
	}
	return s
}

// MustNeg 取反
func MustNeg(str string) string {
	out, err := Neg(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustNeg failed!")
	}
	return out
}

// Neg 取反
func Neg(s string) (string, error) {
	d, err := Decimal(s)
	if err != nil {
		return "", err
	}
	return d.Neg().String(), nil
}

// MustAbs 取绝对值
func MustAbs(str string) string {
	out, err := Abs(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustAbs failed!")
	}
	return out
}

// Abs 取绝对值
func Abs(s string) (string, error) {
	d, err := Decimal(s)
	if err != nil {
		return "", err
	}
	return d.Abs().String(), nil
}

// MustBetween 判断字符串是否在两个数之间
func MustBetween(str, min, max string) bool {
	out, err := Between(str, min, max)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
			"min": min,
			"max": max,
		}).WithError(err).Warnf("MustBetween failed!")
	}
	return out
}

// Between 判断字符串是否在两个数之间
func Between(str, min, max string) (bool, error) {
	d, err := Decimal(str)
	if err != nil {
		return false, err
	}
	minD, err := Decimal(min)
	if err != nil {
		return false, err
	}
	maxD, err := Decimal(max)
	if err != nil {
		return false, err
	}
	return d.GreaterThanOrEqual(minD) && d.LessThanOrEqual(maxD), nil
}

func MustLt(str1, str2 string) bool {
	out, err := Lt(str1, str2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str1": str1,
			"str2": str2,
		}).WithError(err).Warnf("MustLt failed!")
	}
	return out
}

// Lt 判断字符串是否小于
func Lt(str1, str2 string) (bool, error) {
	d1, err := Decimal(str1)
	if err != nil {
		return false, err
	}
	d2, err := Decimal(str2)
	if err != nil {
		return false, err
	}
	return d1.LessThan(d2), nil
}

func MustGt(str1, str2 string) bool {
	out, err := Gt(str1, str2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str1": str1,
			"str2": str2,
		}).WithError(err).Warnf("MustGt failed!")
	}
	return out
}

// Gt 判断字符串是否大于
func Gt(str1, str2 string) (bool, error) {
	d1, err := Decimal(str1)
	if err != nil {
		return false, err
	}
	d2, err := Decimal(str2)
	if err != nil {
		return false, err
	}
	return d1.GreaterThan(d2), nil
}

func MustPrecision(s string) string {
	out, err := Precision(s)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"s": s,
		}).WithError(err).Warnf("MustPrecision failed!")
	}
	return out
}

// Precision 获取小数位数
func Precision(s string) (string, error) {
	d, err := Decimal(s)
	if err != nil {
		return "", err
	}
	scale := d.Exponent()
	if scale < 0 {
		return NewT(-1 * scale), nil
	}
	return "0", nil
}

func MustLte(str1, str2 string) bool {
	out, err := Lte(str1, str2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str1": str1,
			"str2": str2,
		}).WithError(err).Warnf("MustLte failed!")
	}
	return out
}

// Lte 判断字符串是否小于等于
func Lte(str1, str2 string) (bool, error) {
	d1, err := Decimal(str1)
	if err != nil {
		return false, err
	}
	d2, err := Decimal(str2)
	if err != nil {
		return false, err
	}
	return d1.LessThanOrEqual(d2), nil
}

func MustGte(str1, str2 string) bool {
	out, err := Gte(str1, str2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str1": str1,
			"str2": str2,
		}).WithError(err).Warnf("MustGte failed!")
	}
	return out
}

// Gte 判断字符串是否大于等于
func Gte(str1, str2 string) (bool, error) {
	d1, err := Decimal(str1)
	if err != nil {
		return false, err
	}
	d2, err := Decimal(str2)
	if err != nil {
		return false, err
	}
	return d1.GreaterThanOrEqual(d2), nil
}

// MustEqual 判断字符串是否相等
func MustEqual(str1, str2 string) bool {
	out, err := Equal(str1, str2)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str1": str1,
			"str2": str2,
		}).WithError(err).Warnf("MustEqual failed!")
	}
	return out
}

// Equal 判断字符串是否相等
func Equal(str1, str2 string) (bool, error) {
	d1, err := Decimal(str1)
	if err != nil {
		return false, err
	}
	d2, err := Decimal(str2)
	if err != nil {
		return false, err
	}
	return d1.Equal(d2), nil
}

// MustLtZero 判断字符串是否小于0
func MustLtZero(str string) bool {
	out, err := LtZero(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustLtZero failed!")
	}
	return out
}

// LtZero 判断字符串是否小于0
func LtZero(str string) (bool, error) {
	d, err := Decimal(str)
	if err != nil {
		return false, err
	}
	return d.LessThan(decimal.Zero), nil
}

// MustGtZero 判断字符串是否大于0
func MustGtZero(str string) bool {
	out, err := GtZero(str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str": str,
		}).WithError(err).Warnf("MustGtZero failed!")
	}
	return out
}

// GtZero 判断字符串是否大于0
func GtZero(str string) (bool, error) {
	d, err := Decimal(str)
	if err != nil {
		return false, err
	}
	return d.GreaterThan(decimal.Zero), nil
}

// IsZero 判断decimal字符串是否为0
func IsZero(str string) bool {
	d, err := Decimal(str)
	if err != nil {
		return false
	}
	return d.IsZero()
}

func NgZero(strs ...string) bool {
	for _, str := range strs {
		if IsZero(str) {
			return false
		}
	}
	return true
}

// HasZero 判断decimal字符串是否为0
func HasZero(strs ...string) bool {
	for _, str := range strs {
		if IsZero(str) {
			return true
		}
	}
	return false
}

func MustMax(strs ...string) string {
	out, err := Max(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).WithError(err).Warnf("MustMax failed!")
	}
	return out
}

func Max(strs ...string) (string, error) {
	if len(strs) == 0 {
		return "", nil
	}
	maxD, err := Decimal(strs[0])
	if err != nil {
		return "", err
	}
	for _, str := range strs[1:] {
		d, err2 := Decimal(str)
		if err2 != nil {
			return "", err2
		}
		if d.GreaterThan(maxD) {
			maxD = d
		}
	}
	return maxD.String(), nil
}

func MustMin(strs ...string) string {
	out, err := Min(strs...)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"strs": strs,
		}).WithError(err).Warnf("MustMin failed!")
	}
	return out
}

func Min(strs ...string) (string, error) {
	if len(strs) == 0 {
		return "", nil
	}
	minD, err := Decimal(strs[0])
	if err != nil {
		return "", err
	}
	for _, str := range strs[1:] {
		d, err2 := Decimal(str)
		if err2 != nil {
			return "", err2
		}
		if d.LessThan(minD) {
			minD = d
		}
	}
	return minD.String(), nil
}
