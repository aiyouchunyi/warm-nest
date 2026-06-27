// Package traces @Author Larry
// @Date 2024/10/11 15:00
// @Desc

package traces

import (
	"warm-nest/pkg/utils/rands"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

// TraceId generates a trace id with a prefix
func TraceId(prefix string) string {
	return strings.MergeStr("", prefix, times.UnixMilliStr(), rands.Id4())
}

func TraceId18(prefix string) string {
	//整体是18位,prefix 超过原样输出,不够添加随机字符串
	if len(prefix) >= 18 {
		return prefix
	}
	return strings.MergeStr("", prefix, rands.RandomN(18-len(prefix)))
}
