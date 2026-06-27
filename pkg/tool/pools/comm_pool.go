// Package pool @Author larry
// @Date 2025/2/17 16:56
// @Desc

package pools

import (
	"sync"
)

var commPool *Pool
var commPoolOnce sync.Once

// GetCommPool 获取协程池
func GetCommPool() *Pool {
	commPoolOnce.Do(func() {
		commPool = NewPool("CommPool")
	})
	return commPool
}

func SubmitPool(fn func()) error {
	return GetCommPool().pool.Submit(fn)
}
