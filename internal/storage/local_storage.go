// Package storage @Author larry
// @Date 2026/06/15
// @Desc 本地文件存储实现（1 期，主机部署目录持久）

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"warm-nest/internal/config"
)

// localStorage 本地文件存储
type localStorage struct {
	basePath string
	baseUrl  string
}

func newLocalStorage() *localStorage {
	conf := config.StorageConf()
	return &localStorage{
		basePath: conf.BasePath,
		baseUrl:  strings.TrimRight(conf.BaseUrl, "/"),
	}
}

// Save 写入 basePath/bizDir/filename，返回 baseUrl/bizDir/filename
func (s *localStorage) Save(data []byte, bizDir, filename string) (string, error) {
	dir := filepath.Join(s.basePath, bizDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("storage mkdir %s: %w", dir, err)
	}
	full := filepath.Join(dir, filename)
	if err := os.WriteFile(full, data, 0o644); err != nil {
		return "", fmt.Errorf("storage write %s: %w", full, err)
	}
	return fmt.Sprintf("%s/%s/%s", s.baseUrl, bizDir, filename), nil
}

// Get 读取 basePath/relPath
func (s *localStorage) Get(relPath string) ([]byte, error) {
	full := filepath.Join(s.basePath, filepath.Clean("/"+relPath))
	data, err := os.ReadFile(full)
	if err != nil {
		return nil, fmt.Errorf("storage read %s: %w", relPath, err)
	}
	return data, nil
}
