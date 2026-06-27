// Package times @Author larry
// File period_utils.go
// @Date 2024/7/17 10:47:00
// @Desc 时间周期工具类
package times

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	YPeriod          = "Y"
	YMPeriod         = "YM"
	YMDPeriod        = "YMD"
	YPMPDPeriod      = "Y-M-D"
	YMDHPeriod       = "YMDH"
	YMDHMPeriod      = "YMDHM"
	YMDHMSPeriod     = "YMDHMS"
	YPMPDHCMCSPeriod = "Y-M-D H:M:S"

	YPMPDHCMCSUZPeriod = "Y-M-D H:M:S.uZ"
)

// ToMilli 字符串转换为毫秒时间戳
func ToMilli(period string) int64 {
	if period == "" {
		return 0
	}
	layout := ParseLayout(JudgeType(period))
	t, err := time.ParseInLocation(layout, period, time.Now().Location())
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

// ToUnix 字符串转换为毫秒时间戳
func ToUnix(period string) int64 {
	layout := ParseLayout(JudgeType(period))
	t, err := time.ParseInLocation(layout, period, time.Now().Location())
	if err != nil {
		return 0
	}
	return t.Unix()
}

// AddDuration 周期字符串加上指定时间
func AddDuration(period string, duration time.Duration) string {
	periodType := JudgeType(period)
	layout := ParseLayout(periodType)
	t, err := time.ParseInLocation(layout, period, time.Now().Location())
	if err != nil {
		return ""
	}
	t = t.Add(duration)
	return ToPeriod(t, periodType)
}

func Period(periodType string) string {
	return time.Now().Format(ParseLayout(periodType))
}

func PeriodLocal(periodType string) string {
	return ToLocalPeriod(time.Now(), periodType)
}

// MilliToPeriod 毫秒时间戳转换为周期字符串
func MilliToPeriod(milli int64, periodType string) string {
	t := time.UnixMilli(milli)
	return ToPeriod(t, periodType)
}

func MilliToPeriodLocal(milli int64, periodType string) string {
	t := time.UnixMilli(milli)
	return ToLocalPeriod(t, periodType)
}

// ToPeriod 时间转换为周期字符串
func ToPeriod(t time.Time, periodType string) string {
	return t.Format(ParseLayout(periodType))
}

func ToLocalPeriod(t time.Time, periodType string) string {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return t.In(location).Format(ParseLayout(periodType))
}

// ParseLayout 根据周期长度解析layout
func ParseLayout(periodType string) string {
	switch periodType {
	case YPeriod:
		return "2006"
	case YMPeriod:
		return "200601"
	case YMDPeriod:
		return "20060102"
	case YPMPDPeriod:
		return "2006-01-02"
	case YMDHPeriod:
		return "2006010215"
	case YMDHMPeriod:
		return "200601021504"
	case YMDHMSPeriod:
		return "20060102150405"
	case YPMPDHCMCSPeriod:
		return "2006-01-02 15:04:05"
	//case MPDHCMCSPeriod:
	//	return "01-02 15:04:05"
	case YPMPDHCMCSUZPeriod:
		return "2006-01-02 15:04:05.999999-07"
	}
	logrus.WithFields(logrus.Fields{
		"periodType": periodType,
	}).Error("ParseLayout failed!")
	return ""
}

func JudgeType(period string) string {

	switch len(period) {
	case 4:
		return YPeriod
	case 6:
		return YMPeriod
	case 8:
		return YMDPeriod
	case 10:
		if strings.Contains(period, "-") {
			return YPMPDPeriod
		}
		return YMDHPeriod
	case 12:
		return YMDHMPeriod
	case 14:
		return YMDHMSPeriod
	case 29:
		return YPMPDHCMCSUZPeriod
	}
	return ""
}
