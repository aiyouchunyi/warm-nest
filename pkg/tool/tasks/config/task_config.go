// Package config @Author larry
// File task_config.go
// @Date 2024/4/29 09:55:00
// @Desc 任务配置
package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/utils/slices"
)

type TaskConfigs struct {
	Tasks []TaskConfig `toml:"tasks"`
}

type TaskConfig struct {
	Name     string `toml:"name"`
	Spec     string `toml:"spec"`
	Disabled bool   `toml:"disabled"`
}

var taskConf TaskConfigs
var taskConfOnce sync.Once

// TaskConf 获取任务配置单例
func TaskConf() TaskConfigs {
	taskConfOnce.Do(func() {
		taskConf = configs.MustLoad[TaskConfigs](configs.TaskConfPath)
	})
	return taskConf
}

// TaskConfig 通过名称获取任务配置sta
func (t TaskConfigs) TaskConfig(name string) (TaskConfig, bool) {
	return slices.FindOne(t.Tasks, func(task TaskConfig) bool {
		return task.Name == name
	})
}
