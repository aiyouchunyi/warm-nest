package query

import (
	"fmt"
	"reflect"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/app/web/validators"
	"warm-nest/pkg/tool/database/code"
	"warm-nest/pkg/utils/strings"
)

type Condition struct {
	Column   string      `gorm:"comment:字段名" json:"column" validate:"required"`
	Operator Operator    `gorm:"comment:操作符" json:"operator" validate:"required"`
	Value    interface{} `gorm:"comment:操作值" json:"value"`
}

func NewCondition(column string, operator Operator, value interface{}) Condition {
	return Condition{
		Column:   column,
		Operator: operator,
		Value:    value,
	}
}

func (c *Condition) Validate() error {
	if err := validators.Validate(c); err != nil {
		return err
	}
	switch c.Operator {
	case EQ, NEQ, GT, LT, GTE, LTE:
		return c.expectSingle()
	case LIKE, NLIKE:
		return c.expectString()
	case IN, NIN, CONTAIN, HAS, ALL:
		return c.expectSlice()
	case BETWEEN:
		return c.expectSliceLen(2)
	case JEQ, JLIKE, JNLIKE, JNIL, JNNIL:
		jsonPath := strings.Split(c.Column, ".")
		if len(jsonPath) != 2 {
			return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("JSON路径错误: %s", c.Column))
		}
	case NIL, NNIL:
		return nil
	default:
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("操作符错误:%s", c.Operator))
	}
	return nil
}

func (c *Condition) expectSingle() error {
	if c.Value == nil {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值不能为nil  %s:%v", c.Operator, c.Column, c.Value))
	}
	if reflect.TypeOf(c.Value).Kind() == reflect.Slice {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值必须是单值  %s:%v", c.Operator, c.Column, c.Value))
	}
	return nil
}

func (c *Condition) expectString() error {
	if c.Value == nil {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值不能为nil  %s:%v", c.Operator, c.Column, c.Value))
	}
	if reflect.TypeOf(c.Value).Kind() != reflect.String {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值必须是字符串  %s:%v", c.Operator, c.Column, c.Value))
	}
	return nil
}

func (c *Condition) expectSlice() error {
	if c.Value == nil {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值不能为nil  %s:%v", c.Operator, c.Column, c.Value))
	}
	if reflect.TypeOf(c.Value).Kind() != reflect.Slice {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值必须是数组  %s:%v", c.Operator, c.Column, c.Value))
	}
	return nil
}

func (c *Condition) expectSliceLen(length int) error {
	if err := c.expectSlice(); err != nil {
		return err
	}
	if reflect.ValueOf(c.Value).Len() != length {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值必须是长度为%d的数组  %s:%v", c.Operator, length, c.Column, c.Value))
	}
	return nil
}
