// Package pool @Author larry
// @Date 2025/2/17 17:10
// @Desc

package pools

import (
	"time"
)

// NewQueuePool 获取协程池
func NewQueuePool() *Pool {
	return NewWithConfig(Config{
		Name:             "QueuePool",
		MaxGoroutineNum:  1,
		MaxBlockingTasks: 100,
		ExpiryDuration:   60 * time.Second,
	})
}
