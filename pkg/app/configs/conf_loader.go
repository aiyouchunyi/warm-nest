// Package utils @Author larry
// @Date 2025/12/22 16:38
// @Desc

package configs

import (
	"fmt"
	"os"
	"strings"

	"github.com/creasty/defaults"

	"warm-nest/pkg/app/configs/loader"
)

// MustLoad is a generic version of Load that panics on error
func MustLoad[V any](path string) V {
	cfg, err := Load[V](path)
	if err != nil {
		panic(fmt.Sprintf("load config failed: %s", err))
	}
	return cfg
}

// Load is a generic version of Load that returns the configuration object directly
func Load[V any](path string) (V, error) {
	var cfg V
	_ = defaults.Set(&cfg)
	err := loader.LoadTo[V](AppToml, path, &cfg)
	if err != nil {
		return cfg, err
	}
	envFile := fmt.Sprintf("app_%s.toml", strings.ToLower(GetMode()))
	_, err = os.Stat(envFile)
	if err != nil {
		return cfg, nil
	}
	err = loader.LoadTo[V](envFile, path, &cfg)
	if err != nil {
		return cfg, err
	}

	err = loader.LoadEnv(path, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// MustLoadSecret is a generic version of LoadSecret that panics on error
func MustLoadSecret[V any](path string) V {
	cfg, err := LoadSecret[V](path)
	if err != nil {
		panic(fmt.Sprintf("load secret config failed: %s", err))
	}
	return cfg
}

// LoadSecret Config is a generic version of LoadSecret that returns the configuration object directly
func LoadSecret[V any](path string) (V, error) {
	config, err := Load[V](path)
	if err != nil {
		return config, err
	}
	secretId := GetSecretId()
	secretKey := GetSecretKey()
	err = loader.LoadSecretTo(secretId, secretKey, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
