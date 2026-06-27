// Package utils @Author larry
// @Date 2025/12/22 14:29
// @Desc

package loader

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/reflects"
	strings2 "warm-nest/pkg/utils/strings"
)

// LoadEnv 加载环境变量到配置结构体中
func LoadEnv(prefix string, cfg any) error {
	if cfg == nil {
		return nil
	}
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}
	v := reflect.ValueOf(cfg).Elem()
	if v.Kind() == reflect.Map {
		return loadEnvMap(prefix, v)
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("toml")
		if tag == "" {
			tag = field.Name
		}
		if tag == "-" {
			continue
		}
		envKey := prefix
		if envKey != "" {
			envKey += "."
		}
		envKey += tag

		fv := v.Field(i)
		if fv.Kind() == reflect.Struct {
			err := LoadEnv(envKey, fv.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		}

		if fv.Kind() == reflect.Map {
			if err := loadEnvMap(envKey, fv); err != nil {
				return err
			}
			continue
		}

		envVal := getEnvMulti(genEnvKeys(envKey))
		if envVal == "" {
			continue
		}

		err := reflects.SetFieldValue(fv, envVal)
		if err != nil {
			logrus.WithError(err).Errorf("Set env field value failed! key=%s, val=%s", envKey, envVal)
			continue
		}
	}
	return nil
}

// loadEnvMap 递归把 env 覆盖到 map[string]Struct 或 map[string]*Struct 的每个 value。
// 仅在 key 是 string、value 是 struct/*struct 时生效；其他形状（如 map[string]string）静默跳过。
// map value 不可寻址，需要取出 → 新建副本 → 递归 → SetMapIndex 写回。
func loadEnvMap(prefix string, fv reflect.Value) error {
	if fv.Type().Key().Kind() != reflect.String {
		return nil
	}
	elemType := fv.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr
	if isPtr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return nil
	}
	if fv.IsNil() || fv.Len() == 0 {
		return nil
	}
	iter := fv.MapRange()
	for iter.Next() {
		subKey := prefix + "." + iter.Key().String()
		ptr := reflect.New(elemType)
		if isPtr {
			if !iter.Value().IsNil() {
				ptr.Elem().Set(iter.Value().Elem())
			}
		} else {
			ptr.Elem().Set(iter.Value())
		}
		if err := LoadEnv(subKey, ptr.Interface()); err != nil {
			return err
		}
		if isPtr {
			fv.SetMapIndex(iter.Key(), ptr)
		} else {
			fv.SetMapIndex(iter.Key(), ptr.Elem())
		}
	}
	return nil
}

// 获取第一个有值的 env
func getEnvMulti(keys []string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

// genEnvKeys key 输入是 serverConfig.mysqlDb 也可能是 server_config.mysql_db
// 输出固定五种形式（无原始值）：驼峰点、小写点、大写点、小写下划线、大写下划线
func genEnvKeys(key string) []string {
	segments := strings.Split(key, ".")
	var snakeSegs, camelSegs []string
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		snakeSegs = append(snakeSegs, strings2.CamelToSnake(seg))
		camelSegs = append(camelSegs, strings2.SnakeToCamel(seg))
	}

	// 1. 驼峰点分隔（serverConfig.mysqlDb）
	camelDot := strings.Join(camelSegs, ".")
	// 2. 小写下划线点分隔（server_config.mysql_db）
	lowerDot := strings.Join(snakeSegs, ".")
	// 3. 全大写点分隔（SERVER_CONFIG.MYSQL_DB）
	upperDot := strings.ToUpper(lowerDot)
	// 4. 全小写下划线（server_config_mysql_db）
	lowerUnderscore := strings.ReplaceAll(lowerDot, ".", "_")
	// 5. 全大写下划线（SERVER_CONFIG_MYSQL_DB）
	upperUnderscore := strings.ToUpper(lowerUnderscore)

	keys := []string{
		lowerDot,
		upperDot,
		lowerUnderscore,
		upperUnderscore,
	}
	if camelDot != lowerDot {
		keys = append([]string{camelDot}, keys...)
	}
	return keys
}
