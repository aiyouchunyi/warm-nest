// Package loaders @Author larry
// @Date 2025/4/28 17:28
// @Desc Loader 全局注册表：按名注册 / 刷新 / 取值。

package loaders

import (
	"fmt"

	"warm-nest/pkg/tool/concurrent"
	"warm-nest/pkg/utils/reflects"
)

var loaderPool = concurrent.NewCMap[*Loader]()

// Register 注册一个Loader
func Register(name string, refresher Refresher, options ...Option) {
	loaderPool.Store(name, NewLoader(name, refresher, options...))
}

// Refresh 刷新Loader
func Refresh(name string) {
	loader, ok := loaderPool.Load(name)
	if !ok {
		defaultAlert(fmt.Sprintf("Loader %s not found!", name))
		return
	}
	loader.refresh()
}

// GetValue 获取Loader的值
func GetValue[V any](name string) V {
	loader, ok := loaderPool.Load(name)
	if !ok {
		defaultAlert(fmt.Sprintf("Loader %s not found!", name))
		var v V
		return v
	}
	return cast[V](loader)
}

// GetValueV2 获取Loader的值，如果不存在则按 refresher/options 创建一个新的Loader。
// 注意：refresher 与 options 仅在首次创建该 name 的 Loader 时生效，后续同名调用会复用
// 已有 Loader 并忽略传入的 refresher/options。
func GetValueV2[V any](name string, refresher Refresher, options ...Option) V {
	loader, ok := loaderPool.Load(name)
	if !ok {
		loader = NewLoader(name, refresher, options...)
		loaderPool.Store(name, loader)
	}
	return cast[V](loader)
}

func cast[V any](loader *Loader) V {
	v, err := reflects.Cast[V](loader.GetData())
	if err != nil {
		loader.notify(fmt.Errorf("cast failed for loader %s: %w", loader.name, err))
		var zeroValue V
		return zeroValue
	}
	return v
}
