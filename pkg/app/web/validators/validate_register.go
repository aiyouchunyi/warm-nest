// Package validators @Author larry
// @Date 2025/12/25 17:24
// @Desc

package validators

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

func EnableValidate(params ...interface{}) error {
	if len(params) != 1 {
		return fmt.Errorf("validate requires 1 parameters: funcs")
	}
	funcs, ok := params[0].(map[string]func(fl validator.FieldLevel) bool)
	if !ok {
		return fmt.Errorf("invalid parameter types for validate initialization: funcs")
	}
	for key, value := range funcs {
		err := GetValidator().RegisterValidation(key, value)
		if err != nil {
			logrus.WithError(err).Errorf("Register validate function %s failed", key)
		}
	}
	return nil
}
