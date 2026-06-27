// Package tool @Author larry
// @Date 2025/6/24 10:55
// @Desc

package widgets

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"warm-nest/pkg/tool/database/mysqls"
	"warm-nest/pkg/utils/reflects"
)

func Batch[T any](batchSize int, processFunc func([]T) error) error {
	var models []T
	var loop int
	mysqls.DB().FindInBatches(&models, batchSize, func(tx *gorm.DB, batch int) error {
		if err := tx.Error; err != nil {
			return err
		}
		loop++
		logrus.WithFields(logrus.Fields{
			"loop":      loop,
			"batch":     batch,
			"modelName": reflects.ModelNameV2[T](),
		}).Info("Fetching batch of models")
		return processFunc(models)
	})
	return nil
}
