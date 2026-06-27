// Package mysqls @Author larry
// File mysql.go
// @Date 2024/5/22 14:12:00
// @Desc 数据库句柄管理
package mysqls

import (
	"fmt"
	"strings"
	"sync"
	"time"

	mysqld "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"

	"warm-nest/pkg/utils/times"
)

// DB 数据库句柄
var db *gorm.DB

var dbi *gorm.DB
var dbiOnce sync.Once

func DB(opts ...Option) *gorm.DB {
	if db == nil {
		logrus.Errorf("mysql is not start!")
		return nil
	}
	session := NewSession(opts...)
	if session.silent {
		dbiOnce.Do(func() {
			dbi = db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
		})
		return dbi
	}
	return db
}

var enableMu sync.Mutex
var enableCalled bool

// EnableMysql 初始化数据库
// @param url 数据库连接地址
// @return errors 错误信息
func EnableMysql(url string) error {
	enableMu.Lock()
	defer enableMu.Unlock()
	if enableCalled {
		return fmt.Errorf("mysql can only be enabled once")
	}
	start := times.UnixMilli()
	logrus.Info("mysql starting...")
	tdb, err := Connect(url)
	if err != nil {
		logrus.Errorf("mysql start failed! " + err.Error())
		return err
	}
	db = tdb
	logrus.Infof("mysql started... delay: %d ms", times.Gap(start))
	return nil
}

// Connect 链接数据库
// @param url 数据库连接地址
// @return errors 错误信息
func Connect(url string) (*gorm.DB, error) {
	if url == "" {
		logrus.Errorf("mysql url is empty")
		return nil, fmt.Errorf("mysql url is empty")
	}
	dsn, err := mysqld.ParseDSN(url)
	if err != nil {
		logrus.Errorf("mysql parse DSN failed! url: %s, err: %s", url, err.Error())
		return nil, fmt.Errorf("mysql parse DSN failed! url: %s, err: %s", url, err.Error())
	}
	dsn.ParseTime = true
	url = dsn.FormatDSN()
	// replica := parseReplica(url)

	tDB, err := gorm.Open(mysql.Open(url), &gorm.Config{
		Logger: logger.New(
			logrus.StandardLogger(),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				Colorful:                  true,
				IgnoreRecordNotFoundError: true,
			},
		),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		logrus.Errorf("mysql conn failed! " + err.Error())
		return nil, fmt.Errorf("mysql conn failed! " + err.Error())
	}

	err = tDB.Use(dbresolver.Register(
		dbresolver.Config{
			Sources: []gorm.Dialector{mysql.Open(url)},
			// Replicas: []gorm.Dialector{mysql.Open(replica)},
			Policy: dbresolver.RandomPolicy{},
		}).
		SetMaxIdleConns(20).
		SetMaxOpenConns(100).
		SetConnMaxLifetime(time.Hour),
	)

	if err != nil {
		logrus.Errorf("mysql config failed! " + err.Error())
		return nil, fmt.Errorf("mysql config failed! " + err.Error())
	}
	return tDB, nil
}

// 解析分片数据库地址
// @param url 数据库连接地址
// @return string 分片数据库地址
func parseReplica(url string) string {
	if strings.Contains(url, "-ro-") {
		return url
	}
	idx := strings.Index(url, ".cluster")
	if idx < 0 {
		return url
	}
	return fmt.Sprintf("%s-ro%s", url[:idx+8], url[idx+8:])
}
