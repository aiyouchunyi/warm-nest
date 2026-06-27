// Package context @Author larry
// @Date 2025/2/12 11:32
// @Desc

package variable

import (
	"warm-nest/pkg/kinds/variables"
	"warm-nest/pkg/tool/machines/model/constant"
)

type Option func(vars variables.Variables)

// Once 不等待
func Once() Option {
	return MaxWait("0")
}

// MaxWait 结束等待次数[-1重新执行到结束,空-未配置,采用默认配置,n重试N次]
func MaxWait(maxWait string) Option {
	return func(vars variables.Variables) {
		vars.Set(constant.MachineMaxWait, maxWait)
	}
}
