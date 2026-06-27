// Package concurrent @Author larry
// @Date 2025/7/10 15:54
// @Desc

package concurrent

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"warm-nest/pkg/tool/caches"
	"warm-nest/pkg/utils/times"
)

// SignalManager 管理信号量并支持超时机制
type SignalManager struct {
	// sem 信号量实例
	sem *semaphore.Weighted
}

// NewSignal 创建一个新的信号量管理实例
func NewSignal(maxConcurrency int64) *SignalManager {
	return &SignalManager{
		sem: semaphore.NewWeighted(maxConcurrency),
	}
}

// Signal 获取信号管理器
func Signal(maxConcurrency int64, ttlInSec int64, prefix string, params ...string) *SignalManager {
	manager, _ := caches.GetOrLoad[*SignalManager](SingleKey(prefix, params...), ttlInSec, func() (interface{}, error) {
		return NewSignal(maxConcurrency), nil
	})
	return manager
}

// WaitDo 在超时时间内尝试获取信号量并执行指定的操作
func (sm *SignalManager) WaitDo(timeoutInMils int64, action func() (interface{}, error)) (interface{}, error) {
	return sm.WaitDoTrace("", timeoutInMils, action)
}

// WaitDoTrace 在超时时间内尝试获取信号量并执行指定的操作
func (sm *SignalManager) WaitDoTrace(traceId string, timeoutInMils int64, action func() (interface{}, error)) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutInMils)*time.Millisecond)
	defer cancel()

	inTime := times.UnixMilli()
	if err := sm.sem.Acquire(ctx, 1); err != nil {
		return nil, fmt.Errorf("failed to acquire semaphore: %w", err)
	}
	defer sm.sem.Release(1)

	log := logrus.WithFields(logrus.Fields{
		"waitTime": fmt.Sprintf("%dms", times.Gap(inTime)),
	})
	if traceId != "" {
		log = log.WithField("traceId", traceId)
	}
	log.Info("Real do action")

	// 执行动作
	return action()
}
