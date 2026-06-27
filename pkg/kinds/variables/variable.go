// Package json @Author larry
// @Date 2025/4/9 15:19
// @Desc

package variables

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/shopspring/decimal"

	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/transforms"
)

type Variables map[string]string

func NewVariables() Variables {
	return make(Variables)
}

// Value implements the driver.Valuer interface
func (vars Variables) Value() (driver.Value, error) {
	return json.Marshal(vars)
}

// Scan implements the sql.Scanner interface
func (vars *Variables) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, vars)
}

// Set 添加变量
func (vars *Variables) Set(name string, value interface{}) *Variables {
	if *vars == nil {
		*vars = make(Variables)
	}
	(*vars)[name] = reflects.ToString(value)
	return vars
}

// AddIfAbsent 添加变量
func (vars *Variables) AddIfAbsent(name string, value interface{}) *Variables {
	if *vars == nil {
		*vars = make(Variables)
	}
	if _, ok := (*vars)[name]; !ok {
		vars.Set(name, value)
	}
	return vars
}

// Get 获取变量
func (vars *Variables) get(name string) string {
	if *vars == nil {
		return ""
	}
	return (*vars)[name]
}

// GetString 获取字符串变量
func (vars *Variables) GetString(name string) string {
	return reflects.MustFrom[string](vars.get(name))
}

// GetBool 获取布尔变量
func (vars *Variables) GetBool(name string) bool {
	return reflects.MustFrom[bool](vars.get(name))
}

// GetInt 获取整数变量
func (vars *Variables) GetInt(name string) int {
	return reflects.MustFrom[int](vars.get(name))
}

// GetInt64 获取整数变量
func (vars *Variables) GetInt64(name string) int64 {
	return reflects.MustFrom[int64](vars.get(name))
}

func (vars *Variables) GetFloat64(name string) float64 {
	return reflects.MustFrom[float64](vars.get(name))
}

func (vars *Variables) GetDecimal(name string) decimal.Decimal {
	return reflects.MustFrom[decimal.Decimal](vars.get(name))
}

// Increment 自增
func (vars *Variables) Increment(name string) int {
	value := vars.GetInt(name)
	value++
	vars.Set(name, value)
	return value
}

// ToString 转换为json字符串
func (vars Variables) ToString() string {
	return transforms.Marshal(vars)
}
