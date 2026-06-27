// Package widget @Author larry
// @Date 2025/8/1 14:30
// @Desc

package widgets

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type DoMigrate func() error

func EnableMigrate(params ...interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("migrate requires 1 parameters: doMigrates")
	}
	doMigrates, ok := params[0].(map[string]DoMigrate)
	if !ok {
		return fmt.Errorf("invalid parameter types for Migrate initialization: doMigrates")
	}
	if len(doMigrates) == 0 {
		return nil
	}

	go func() {
		for name, doMigrate := range doMigrates {
			logrus.Infof("migration: %s start", name)
			err := doMigrate()
			if err != nil {
				logrus.WithError(err).Errorf("Executing migration: %s failed", name)
			}
			logrus.Infof("migration: %s end", name)
		}
	}()
	return nil
}
