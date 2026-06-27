// Package utils @Author larry
// @Date 2025/4/21 17:18
// @Desc

package utils

import (
	"warm-nest/pkg/utils/reflects"
)

func GetID(model any) (string, error) {
	value, err := reflects.GetModelField(model, "ID")
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func IsNew(model any) (bool, error) {
	id, err := GetID(model)
	if err != nil {
		return false, err
	}
	if id == "" {
		return true, nil
	}
	return false, nil
}
