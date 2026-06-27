// Package pools @Author larry
// @Date 2024/4/19 16:27:00
// @Desc 协程池
package pools

import (
	"fmt"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
)

type Config struct {
	Name             string        `gorm:"comment:协程池名称"`
	MaxGoroutineNum  int           `gorm:"comment:最大协程数"`
	MaxBlockingTasks int           `gorm:"comment:最大阻塞任务数"`
	ExpiryDuration   time.Duration `gorm:"comment:过期时间"`
}

// Pool 可解决Panic,未捕获问题
type Pool struct {
	pool *ants.Pool
}

func NewPool(name string) *Pool {
	return NewWithConfig(Config{
		Name:             name,
		MaxGoroutineNum:  4,
		MaxBlockingTasks: 100,
		ExpiryDuration:   60 * time.Second,
	})
}

// Submit 提交任务
func (p *Pool) Submit(task func()) error {
	return p.pool.Submit(task)
}

// NewWithConfig 创建协程池
func NewWithConfig(config Config) *Pool {
	pool, err := ants.NewPool(
		config.MaxGoroutineNum,
		ants.WithExpiryDuration(config.ExpiryDuration),
		ants.WithPanicHandler(func(v interface{}) {
			logrus.Errorf("[Pool-%s] recovered from panic! error= %+v stack=%+v", v, config.Name, errors.Callers())
		}),
		ants.WithMaxBlockingTasks(config.MaxBlockingTasks),
		ants.WithLogger(logrus.StandardLogger()),
	)
	if err != nil {
		panic(fmt.Sprintf("create ants pool failed: %+v", err))
	}
	return &Pool{
		pool: pool,
	}
}
