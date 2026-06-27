// Package widget @Author Larry
// @Date 2024/11/13 17:45
// @Desc

package widgets

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/database/mysqls"
)

func EnableModel(params ...interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("model requires 1 parameters: models")
	}
	models, ok := params[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid parameter types for model initialization: models")
	}
	for comment, m := range models {
		err := CreateModel(comment, m)
		if err != nil {
			continue
		}
	}
	return nil
}

func CreateModel(comment string, model interface{}) error {
	if mysqls.DB().Migrator().HasTable(model) {
		return nil
	}
	err := mysqls.DB().Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").Migrator().CreateTable(model)
	if err != nil {
		logrus.WithError(err).Errorf("Create table %s failed", comment)
		return err
	}

	tableType, _ := mysqls.DB().Migrator().TableType(model)
	sql := fmt.Sprintf("ALTER TABLE %s COMMENT '%s'", tableType.Name(), comment)
	mysqls.DB().Exec(sql)

	sql2 := fmt.Sprintf("ALTER TABLE %s DEFAULT CHARACTER SET %s", tableType.Name(), "utf8mb4")
	mysqls.DB().Exec(sql2)

	logrus.Infof("Create table %s success", comment)
	return nil
}

func EnableTable(params ...interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("tableSql requires 1 parameters: sqls")
	}
	modelSQLs, ok := params[0].(map[string]string)
	if !ok {
		return fmt.Errorf("invalid parameter types for tableSql initialization: sqls")
	}
	for tableName, SQL := range modelSQLs {
		err := CreateTable(tableName, SQL)
		if err != nil {
			continue
		}
	}
	return nil
}

// CreateTable 创建表
func CreateTable(tableName string, SQL string) error {
	if mysqls.DB().Migrator().HasTable(tableName) {
		return nil
	}
	err := mysqls.DB().Exec(SQL).Error
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"SQL": SQL,
		}).WithError(err).Errorf("Exec Create table SQL failed!")
	}
	return err
}
