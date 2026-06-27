// Package database @Author larry
// @Date 2025/12/25 16:46
// @Desc

package database

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/database/config"
	"warm-nest/pkg/tool/database/mongos"
	"warm-nest/pkg/tool/database/mysqls"
)

const (
	MysqlDriver = "mysql"
	MongoDriver = "mongo"
)

// EnableDriver 初始化数据库工具包
func EnableDriver(params ...interface{}) error {
	for _, param := range params {
		switch param.(string) {
		case MysqlDriver:
			if err := mysqls.EnableMysql(config.DatabaseConf().Mysql); err != nil {
				return err
			}
		case MongoDriver:
			if err := mongos.EnableMongo(config.DatabaseConf().Mongo); err != nil {
				return err
			}
		default:
			logrus.Errorf("Unknown database driver: %v", param)
		}
	}
	return nil
}
