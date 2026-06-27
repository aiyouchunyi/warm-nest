// Package config @Author larry
// File machine_config.go
// @Date 2024/4/30 15:36:00
// @Desc 状态机配置
package config

import (
	"sync"

	"github.com/creasty/defaults"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/utils/slices"
	"warm-nest/pkg/utils/strings"
)

type MachineConfigs struct {
	Machines []MachineConfig `toml:"machines" json:"machines" gorm:"comment:状态机配置列表"`
}

type MachineConfig struct {
	Name        string `toml:"name" gorm:"comment:状态机名称" json:"name"`
	MaxWait     int64  `toml:"maxWait"  gorm:"comment:结束等待次数" default:"10" json:"maxWait"`
	LockTime    int64  `toml:"lockTime" gorm:"comment:锁定超时时间*ms" default:"300000" json:"lockTime"`
	MonitorTime int64  `toml:"monitorTime" gorm:"comment:未完成任务监控时间*ms" default:"6048000000" json:"monitorTime"`
	ScanLock    int64  `toml:"scanLock" gorm:"comment:扫描锁定时间*s[<0不扫描]" default:"600" json:"scanLock"`
}

var machineConf MachineConfigs
var machineConfOnce sync.Once

// MachineConf 获取状态机配置单例
func MachineConf() MachineConfigs {
	machineConfOnce.Do(func() {
		// 改用 Load 而非 MustLoad，支持配置文件不存在时使用默认值
		conf, err := configs.Load[MachineConfigs](configs.MachineConfPath)
		if err != nil {
			// 配置文件不存在时使用空配置（依赖 defaults 标签提供默认值）
			machineConf = MachineConfigs{Machines: []MachineConfig{}}
			return
		}
		machineConf = conf
	})
	return machineConf
}

func GetMachineConfig(name string) MachineConfig {
	name = strings.LowerFirst(name)
	machineConfig, ok := slices.FindOne(MachineConf().Machines, func(config MachineConfig) bool {
		return config.Name == name
	})
	if !ok {
		machineConfig = MachineConfig{
			Name: name,
		}
	}
	_ = defaults.Set(&machineConfig)
	return machineConfig
}

type Option func(*MachineConfig)

func WithMaxWait(maxWait int64) Option {
	return func(mc *MachineConfig) {
		mc.MaxWait = maxWait
	}
}

func WithLockTime(lockTime int64) Option {
	return func(mc *MachineConfig) {
		mc.LockTime = lockTime
	}
}

func WithMonitorTime(monitorTime int64) Option {
	return func(mc *MachineConfig) {
		mc.MonitorTime = monitorTime
	}
}

func WithScanLock(scanLock int64) Option {
	return func(mc *MachineConfig) {
		mc.ScanLock = scanLock
	}
}
