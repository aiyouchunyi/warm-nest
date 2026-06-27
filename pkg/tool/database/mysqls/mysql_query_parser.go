package mysqls

import (
	"fmt"
	"reflect"

	"github.com/creasty/defaults"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"

	"warm-nest/pkg/tool/database/query"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/transforms"
)

// Parse 转换为表达式
func Parse(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]clause.Expression, error) {
	expressions := make([]clause.Expression, 0)
	expressions = append(expressions, toSelect(req))
	conditions, err := ToConditions(req, fn...)
	if err != nil {
		return nil, err
	}
	expressions = append(expressions, conditions...)
	expressions = append(expressions, toLimit(req))
	expressions = append(expressions, toSort(req))
	return expressions, nil
}

func toSelect(req query.QueryReq) clause.Select {
	columns := make([]clause.Column, 0, len(req.Columns))
	for _, column := range req.Columns {
		columns = append(columns, clause.Column{
			Name: strings.CamelToUnderline(column),
		})
	}
	return clause.Select{Columns: columns}
}

func ToConditions(req query.QueryReq, fn ...func(c query.Custom) (clause.Expression, error)) ([]clause.Expression, error) {
	expressions := make([]clause.Expression, 0)
	conditions, err := condition(req)
	if err != nil {
		return nil, err
	}
	expressions = append(expressions, conditions...)
	if len(fn) != 0 {
		customs, err := custom(req, fn[0])
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, customs...)
	}
	return expressions, nil
}

func toSort(req query.QueryReq) clause.OrderBy {
	if len(req.Sort) == 0 {
		return clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{Column: clause.Column{Name: "created_at"}, Desc: true},
				{Column: clause.Column{Name: "id"}},
			},
		}
	}
	orderBy := clause.OrderBy{}
	for _, sort := range req.Sort {
		orderBy.Columns = append(orderBy.Columns, clause.OrderByColumn{
			Column: clause.Column{Name: strings.CamelToUnderline(sort.Column)},
			Desc:   sort.Order == query.SortTypeDESC,
		})
	}
	return orderBy
}

func toLimit(req query.QueryReq) clause.Limit {
	if req.Size == -1 {
		return clause.Limit{}
	}
	_ = defaults.Set(req)
	return clause.Limit{
		Limit:  &req.Size,
		Offset: (req.Page - 1) * req.Size,
	}
}

func condition(req query.QueryReq) ([]clause.Expression, error) {
	expressions := make([]clause.Expression, 0)
	for _, tCondition := range req.Conditions {
		err := tCondition.Validate()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"condition": tCondition,
			}).WithError(err).Error("Condition validate failed!")
			return nil, err
		}
		expressions = append(expressions, toExpression(tCondition))
	}
	return expressions, nil
}

func custom(req query.QueryReq, fn func(c query.Custom) (clause.Expression, error)) ([]clause.Expression, error) {
	expressions := make([]clause.Expression, 0)
	for _, tCustom := range req.Customs {
		expression, err := fn(tCustom)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"custom": tCustom,
			}).WithError(err).Error("Custom validate failed!")
			return nil, err
		}
		expressions = append(expressions, expression)
	}
	return expressions, nil
}

// toCondExpression 转换为表达式
func toExpression(cond query.Condition) clause.Expression {
	column := strings.CamelToUnderline(cond.Column)
	switch cond.Operator {
	case query.EQ:
		return clause.Eq{Column: column, Value: cond.Value}
	case query.NEQ:
		return clause.Neq{Column: column, Value: cond.Value}
	case query.GT:
		return clause.Gt{Column: column, Value: cond.Value}
	case query.LT:
		return clause.Lt{Column: column, Value: cond.Value}
	case query.GTE:
		return clause.Gte{Column: column, Value: cond.Value}
	case query.LTE:
		return clause.Lte{Column: column, Value: cond.Value}
	case query.IN:
		return clause.IN{Column: column, Values: reflects.ToSlice(cond.Value)}
	case query.NIN:
		return clause.Not(clause.IN{
			Column: column,
			Values: reflects.ToSlice(cond.Value),
		})
	case query.LIKE:
		return clause.Like{Column: column, Value: fmt.Sprintf("%%%s%%", cond.Value)}
	case query.NLIKE:
		return clause.Not(clause.Like{Column: column, Value: fmt.Sprintf("%%%s%%", cond.Value)})
	case query.NIL:
		return clause.Eq{Column: column, Value: nil}
	case query.NNIL:
		return clause.Neq{Column: column, Value: nil}
	case query.BETWEEN:
		val := reflect.ValueOf(cond.Value)
		if val.Kind() == reflect.Slice && val.Len() == 2 {
			return clause.Expr{
				SQL:  "`" + column + "` BETWEEN ? AND ?",
				Vars: []interface{}{val.Index(0).Interface(), val.Index(1).Interface()},
			}
		}
		return nil
	case query.CONTAIN:
		return clause.Expr{
			SQL:  fmt.Sprintf("JSON_CONTAINS(`%s`, ?)", column),
			Vars: []interface{}{transforms.Marshal(cond.Value)},
		}
	case query.HAS:
		var exprs []clause.Expression
		columnExpr := fmt.Sprintf("JSON_CONTAINS(`%s`, ?)", column)
		for _, elem := range reflects.ToSlice(cond.Value) {
			exprs = append(exprs, clause.Expr{
				SQL:  columnExpr,
				Vars: []interface{}{transforms.Marshal([]interface{}{elem})},
			})
		}
		return clause.Or(exprs...)
	case query.ALL:
		return clause.Expr{
			SQL:  fmt.Sprintf("JSON_CONTAINS(`%s`, ?) AND JSON_LENGTH(`%s`)=?", column, column),
			Vars: []interface{}{transforms.Marshal(cond.Value), len(reflects.ToSlice(cond.Value))},
		}
	case query.JEQ:
		jsonPath, lastKey := parseJsonPath(cond.Column)
		return clause.Expr{
			SQL:  fmt.Sprintf("JSON_EXTRACT(`%s`, '$.%s') = ?", jsonPath, lastKey),
			Vars: []interface{}{cond.Value},
		}
	case query.JLIKE:
		jsonPath, lastKey := parseJsonPath(cond.Column)
		return clause.Expr{
			SQL:  fmt.Sprintf("JSON_EXTRACT(`%s`, '$.%s') LIKE ?", jsonPath, lastKey),
			Vars: []interface{}{fmt.Sprintf("%%%s%%", cond.Value)},
		}
	case query.JNLIKE:
		jsonPath, lastKey := parseJsonPath(cond.Column)
		return clause.Not(clause.Expr{
			SQL:  fmt.Sprintf("JSON_EXTRACT(`%s`, '$.%s') LIKE ?", jsonPath, lastKey),
			Vars: []interface{}{fmt.Sprintf("%%%s%%", cond.Value)},
		})
	case query.JNIL:
		jsonPath, lastKey := parseJsonPath(cond.Column)
		return clause.Expr{
			SQL:  fmt.Sprintf("JSON_EXTRACT(`%s`, '$.%s') IS NULL", jsonPath, lastKey),
			Vars: []interface{}{},
		}
	case query.JNNIL:
		jsonPath, lastKey := parseJsonPath(cond.Column)
		return clause.Expr{
			SQL:  fmt.Sprintf("JSON_EXTRACT(`%s`, '$.%s') IS NOT NULL", jsonPath, lastKey),
			Vars: []interface{}{},
		}
	default:
		return nil
	}
}

func parseJsonPath(column string) (string, string) {
	jsonPath := strings.Split(column, ".")
	if len(jsonPath) < 2 {
		return "", ""
	}
	lastKey := jsonPath[len(jsonPath)-1]
	return strings.CamelToUnderline(jsonPath[0]), lastKey
}
