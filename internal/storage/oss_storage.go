// Package storage @Author larry
// @Date 2026/06/27
// @Desc 阿里云 OSS 图片存储实现（公共读 Bucket）
//
// 公共读：Save 上传后返回完整永久公网 URL，直接入库、展示无需签名。
// 上传走内网 endpoint（服务在 ECS、与 OSS 同地域，免流量费）；返回 URL 用外网域名（用户公网访问）。

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"warm-nest/internal/config"
)

// ossStorage 阿里云 OSS 存储
type ossStorage struct {
	client    *oss.Client
	bucket    string
	publicUrl string // 公网访问 URL 前缀 https://<bucket>.<外网endpoint>
}

func newOssStorage() *ossStorage {
	conf := config.OssConf()
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.AccessKeyId, conf.AccessKeySecret)).
		WithRegion(conf.Region).
		WithEndpoint(conf.Endpoint)
	return &ossStorage{
		client:    oss.NewClient(cfg),
		bucket:    conf.Bucket,
		publicUrl: fmt.Sprintf("https://%s.%s", conf.Bucket, conf.PublicEndpoint),
	}
}

// Save 上传到 bizDir/filename，返回完整公网 URL（公共读，永久可访问，入库用）
func (s *ossStorage) Save(data []byte, bizDir, filename string) (string, error) {
	key := fmt.Sprintf("%s/%s", bizDir, filename)
	_, err := s.client.PutObject(context.Background(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(s.bucket),
		Key:    oss.Ptr(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("oss put object %s: %w", key, err)
	}
	return fmt.Sprintf("%s/%s", s.publicUrl, key), nil
}

// Get 按完整 URL 反解 key 下载字节流（公共读也可直接 HTTP，这里走 SDK 统一）
func (s *ossStorage) Get(relPath string) ([]byte, error) {
	key := relPath
	if prefix := s.publicUrl + "/"; len(relPath) > len(prefix) && relPath[:len(prefix)] == prefix {
		key = relPath[len(prefix):]
	}
	result, err := s.client.GetObject(context.Background(), &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		return nil, fmt.Errorf("oss get object %s: %w", key, err)
	}
	defer result.Body.Close()
	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("oss read object %s: %w", key, err)
	}
	return data, nil
}
