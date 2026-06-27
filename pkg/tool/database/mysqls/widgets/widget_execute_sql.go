// Package widget @Author Larry
// @Date 2024/10/18 15:02
// @Desc

package widgets

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/database/mysqls"
)

func AutoExec(params ...interface{}) error {
	SQLs := params[0].([]string)
	for _, SQL := range SQLs {
		err := mysqls.DB().Exec(SQL).Error
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"SQL": SQL,
			}).WithError(err).Errorf("Exec SQL failed")
			continue
		}
	}
	return nil
}
