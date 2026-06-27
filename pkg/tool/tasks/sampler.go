// Package tasks @Author larry
// @Date 2025/9/3 16:25
// @Desc

package tasks

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/times"
)

type Sample func(latestPeriod string) error

type ISampler interface {
	GetLastPeriod() (string, error)
	Sample(latestPeriod string) error
}

type SamplerExecutor struct {
	Sampler    ISampler
	Name       string `gorm:"comment:采样器名称"`
	SampleTime int64  `gorm:"comment:采样时间"`
	LastPeriod string `gorm:"comment:最后采样周期"`
}

func NewSampler(name string, sampleTime int64, Sampler ISampler) *SamplerExecutor {
	return &SamplerExecutor{
		Sampler:    Sampler,
		Name:       name,
		SampleTime: sampleTime,
	}
}

func (h *SamplerExecutor) Snap(_ ...any) error {
	// 1. 获取最新的采样周期
	latestPeriod := h.latestPeriod()
	log := logrus.WithFields(logrus.Fields{
		"sampler":      h.Name,
		"latestPeriod": latestPeriod,
	})
	if h.LastPeriod == "" {
		lastPeriod, err := h.Sampler.GetLastPeriod()
		if err != nil {
			logrus.Error(fmt.Sprintf("%s get last period failed! err: %v", h.Name, err))
			return fmt.Errorf("%s get last period failed! err: %w", h.Name, err)
		}
		h.LastPeriod = lastPeriod
	}

	// 2. 如果最后采样周期等于最新采样周期，则不需要快照
	if h.LastPeriod == latestPeriod {
		log.Info("No need to sample!")
		return nil
	}

	// 4. 进行采样
	err := h.Sampler.Sample(latestPeriod)
	if err != nil {
		logrus.Error(fmt.Sprintf("sample failed! name: %s, latestPeriod: %s, err: %v", h.Name, latestPeriod, err))
		return fmt.Errorf("sample failed! name: %s, latestPeriod: %s, err: %w", h.Name, latestPeriod, err)
	}
	log.Info("Sample success!")
	h.LastPeriod = latestPeriod
	return nil
}

// getLastPeriod 获取采样周期
func (h *SamplerExecutor) latestPeriod() string {
	return times.MilliToPeriodLocal(time.Now().UnixMilli()/(h.SampleTime)*(h.SampleTime), times.YMDHMPeriod)
}
