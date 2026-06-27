// Package loaders @Author larry
// @Date 2025/4/28 17:06
// @Desc 带定时刷新的懒加载缓存：首次取值触发加载并启动后台刷新循环。
package loaders

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const defaultRefreshInSec = 60

type Refresher func() (any, error)

type Alert func(err error)

// AlertFunc 全局默认告警器，入参为已格式化的告警文案。
type AlertFunc func(msg string)

// defaultAlert 默认仅打日志，避免 loaders 这一底层包硬依赖具体告警实现（否则
// 告警配置又缓存在 loaders 里，会形成循环调用甚至死锁）。
// 业务侧（app 启动）通过 SetDefaultAlert 注入自定义告警器。
var defaultAlert AlertFunc = func(msg string) {
	logrus.Error(msg)
}

// SetDefaultAlert 注入全局默认告警器。由 app 启动时调用。
func SetDefaultAlert(alert AlertFunc) {
	if alert != nil {
		defaultAlert = alert
	}
}

type Loader struct {
	once  sync.Once
	mutex sync.Mutex

	name         string
	refresher    Refresher
	alert        Alert
	refreshInSec int

	data any
}

type Option func(loader *Loader)

func WithRefresh(refreshInSec int) Option {
	return func(loader *Loader) {
		loader.refreshInSec = refreshInSec
	}
}

func WithAlert(alert Alert) Option {
	return func(loader *Loader) {
		loader.alert = alert
	}
}

func NewLoader(name string, refresher Refresher, options ...Option) *Loader {
	loader := &Loader{
		name:         name,
		refresher:    refresher,
		refreshInSec: defaultRefreshInSec,
	}
	for _, option := range options {
		option(loader)
	}
	return loader
}

func (loader *Loader) GetData() any {
	loader.once.Do(func() {
		loader.refresh()
		loader.startRefreshLoop()
	})
	loader.mutex.Lock()
	defer loader.mutex.Unlock()
	return loader.data
}

// notify 上报失败（刷新失败 / 取值类型转换失败）：优先用 loader 自定义告警器
// （WithAlert），否则走全局 defaultAlert。配了 WithAlert 的 loader 完全旁路
// defaultAlert —— 这是切断「告警 → 读告警配置 → 失败 → 再告警」循环的终止点。
func (loader *Loader) notify(err error) {
	if loader.alert != nil {
		loader.alert(err)
		return
	}
	defaultAlert(fmt.Sprintf("%s loader failed: %v", loader.name, err))
}

func (loader *Loader) startRefreshLoop() {
	go func() {
		logrus.Infof("%s loader refreshLoop start... %d/s", loader.name, loader.refreshInSec)
		ticker := time.NewTicker(time.Second * time.Duration(loader.refreshInSec))
		defer ticker.Stop()
		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						loader.notify(fmt.Errorf("refreshLoop recovered from panic: %v", r))
					}
				}()
				loader.refresh()
			}()
		}
	}()
}

func (loader *Loader) refresh() {
	v, err := loader.refresher()
	if err != nil {
		loader.notify(err)
		return
	}
	loader.mutex.Lock()
	defer loader.mutex.Unlock()
	loader.data = v
	logrus.Infof("%s loader refreshed...", loader.name)
}
