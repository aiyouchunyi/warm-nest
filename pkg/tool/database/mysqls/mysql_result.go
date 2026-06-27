// Package mysqls @Author larry
// @Date 2025/4/8 14:17
// @Desc

package mysqls

import (
	stdErrors "errors"
	"strings"

	"gorm.io/gorm"

	"warm-nest/pkg/tool/database/code"
)

// UpdateSucc 检查更新是否成功
func UpdateSucc(result *gorm.DB) (bool, error) {
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func UpdateRows(result *gorm.DB) (int64, error) {
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// IsRecordNotFound 判断是否为记录未找到错误
func IsRecordNotFound(err error) bool {
	return err != nil &&
		(stdErrors.Is(err, gorm.ErrRecordNotFound) ||
			code.DBNotFound.Is(err) ||
			strings.Contains(err.Error(), "record not found"))
}

func IsDuplicated(err error) bool {
	return err != nil &&
		(stdErrors.Is(err, gorm.ErrDuplicatedKey) ||
			strings.Contains(err.Error(), "Duplicate entry") ||
			strings.Contains(err.Error(), "duplicate entry") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "Error 1062"))
}

// DoesNotExist 判断是否为表不存在错误
func DoesNotExist(err error) bool {
	return err != nil && strings.Contains(err.Error(), "doesn't exist")
}
