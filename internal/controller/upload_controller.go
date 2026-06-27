// Package controller @Author larry
// @Date 2026/06/15
// @Desc 图片上传 Controller（multipart，走 RawHandler；实现下沉 api.ApiUpload）
//
// web.Handler[T] 只绑 JSON，multipart 必须用 web.RawHandler + 原生 *gin.Context。
// Controller 只做路由 + 静态托管，上传实现（校验/读文件/存储）在 api.ApiUpload.Upload。

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
	"warm-nest/internal/config"
)

// UploadController 图片上传 Controller
type UploadController struct {
	apiUpload *api.ApiUpload
}

var uploadController *UploadController
var uploadOnce sync.Once

// GetUploadController 获取上传 Controller 单例
func GetUploadController() *UploadController {
	uploadOnce.Do(func() {
		uploadController = &UploadController{apiUpload: api.GetApiUpload()}
	})
	return uploadController
}

// Router 注册路由
func (c *UploadController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1")
	group.POST("/upload", web.RawHandler(c.apiUpload.Upload, widgets.Session)) // 上传需登录

	// 本地存储时托管静态文件目录（免登录，图片由 URL 直接读取）；OSS 模式由对象存储自身提供访问
	if conf := config.StorageConf(); conf.Kind == config.StorageKindLocal {
		router.Static("/static", conf.BasePath)
	}
}
