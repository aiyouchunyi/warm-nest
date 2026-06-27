// Package api @Author larry
// @Date 2026/06/15
// @Desc 图片上传 API（multipart 文件流，业务实现层）
//
// 上传走 multipart/form-data，框架泛型 web.Handler[T] 只绑 JSON 绑不了，
// 故 controller 用 web.RawHandler + 原生 *gin.Context。但实现（校验/读文件/存储）
// 下沉到本 api 层，controller 只做路由——与其它接口一致的分层。

package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web/resp"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/config"
	"warm-nest/internal/storage"
)

// allowedImageExt 允许的图片扩展名白名单
var allowedImageExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true}

// ApiUpload 图片上传 API
type ApiUpload struct{}

var apiUpload *ApiUpload
var apiUploadOnce sync.Once

// GetApiUpload 获取上传 API 单例
func GetApiUpload() *ApiUpload {
	apiUploadOnce.Do(func() {
		apiUpload = &ApiUpload{}
	})
	return apiUpload
}

// Upload 接收图片，校验后存储，返回访问 URL（RawHandler 入口，自行处理 multipart）
func (a *ApiUpload) Upload(ctx *gin.Context) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return fmt.Errorf("read upload file: %w", err)
	}
	conf := config.StorageConf()
	if file.Size > conf.MaxFileSize {
		return fmt.Errorf("file too large: %d > %d", file.Size, conf.MaxFileSize)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		return fmt.Errorf("unsupported file type: %s", ext)
	}
	data, err := readMultipart(file)
	if err != nil {
		return err
	}
	// 服务端生成随机文件名，防路径穿越；bizDir 固定枚举
	filename := rands.Id16() + ext
	url, err := storage.Get().Save(data, "checkin", filename)
	if err != nil {
		return fmt.Errorf("save upload file: %w", err)
	}
	resp.Success(ctx, map[string]string{"url": url})
	return nil
}

// readMultipart 读取上传文件全部字节
func readMultipart(file *multipart.FileHeader) ([]byte, error) {
	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open upload file: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read upload file content: %w", err)
	}
	return data, nil
}
