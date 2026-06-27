// Package mongos @Author larry
// @Date 2025/4/21 09:54
// @Desc

package mongos

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"warm-nest/pkg/tool/database/query"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/strings"
)

func Parse(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (bson.M, error) {
	filter := bson.M{}
	// 1. 添加条件
	conditions, err := toConditions(req)
	if err != nil {
		return nil, err
	}
	for k, v := range conditions {
		filter[k] = v
	}

	// 2. 添加自定义条件
	customs, err := toCustoms(req, fn...)
	if err != nil {
		return nil, err
	}
	for k, v := range customs {
		filter[k] = v
	}

	return filter, nil
}

func FindOptions(req query.QueryReq) *options.FindOptions {
	opts := options.Find()
	opts.SetProjection(toSelect(req))
	opts.SetSort(toSort(req))
	if req.Size > 0 {
		opts.SetSkip(int64(req.Page-1) * int64(req.Size))
		opts.SetLimit(int64(req.Size))
	}
	return opts
}

func FindOneOptions(req query.QueryReq) *options.FindOneOptions {
	opts := options.FindOne()
	opts.SetProjection(toSelect(req))
	opts.SetSort(toSort(req))
	if req.Size > 0 {
		opts.SetSkip(int64(req.Page-1) * int64(req.Size))
	}
	return opts
}

// toSelect 转换为字段过滤器
func toSelect(req query.QueryReq) bson.M {
	if len(req.Columns) == 0 {
		return nil
	}
	projection := bson.M{}
	for _, column := range req.Columns {
		projection[strings.LowerFirst(column)] = 1
	}
	return projection
}

// toConditions 转换为条件过滤器
func toConditions(req query.QueryReq) (bson.M, error) {
	filter := bson.M{}
	for _, cond := range req.Conditions {
		if err := cond.Validate(); err != nil {
			return nil, err
		}
		filter[strings.LowerFirst(cond.Column)] = toExpression(cond)
	}
	return filter, nil
}

// toCustoms 转换为自定义条件
func toCustoms(req query.QueryReq, fn ...func(c query.Custom) (bson.M, error)) (bson.M, error) {
	if len(fn) == 0 {
		return nil, nil
	}
	filter := bson.M{}
	for _, custom := range req.Customs {
		expr, err := fn[0](custom)
		if err != nil {
			return nil, err
		}
		for k, v := range expr {
			filter[k] = v
		}
	}
	return filter, nil
}

// toSort 实现排序逻辑
func toSort(req query.QueryReq) bson.M {
	if len(req.Sort) == 0 {
		return bson.M{"createdAt": -1}
	}

	sort := bson.M{}
	for _, s := range req.Sort {
		order := 1 // 默认升序
		if s.Order == query.SortTypeDESC {
			order = -1 // 降序
		}
		sort[strings.CamelToUnderline(s.Column)] = order
	}
	return sort
}

// toExpression 转换单个条件为 MongoDB 表达式
func toExpression(cond query.Condition) bson.M {
	switch cond.Operator {
	case query.EQ:
		return bson.M{"$eq": cond.Value}
	case query.NEQ:
		return bson.M{"$ne": cond.Value}
	case query.GT:
		return bson.M{"$gt": cond.Value}
	case query.LT:
		return bson.M{"$lt": cond.Value}
	case query.GTE:
		return bson.M{"$gte": cond.Value}
	case query.LTE:
		return bson.M{"$lte": cond.Value}
	case query.IN:
		return bson.M{"$in": reflects.ToSlice(cond.Value)}
	case query.NIN:
		return bson.M{"$nin": reflects.ToSlice(cond.Value)}
	case query.LIKE:
		return bson.M{"$regex": cond.Value, "$options": "i"}
	case query.NLIKE:
		return bson.M{"$not": bson.M{"$regex": cond.Value, "$options": "i"}}
	case query.NIL:
		return bson.M{"$eq": nil}
	case query.NNIL:
		return bson.M{"$ne": nil}
	case query.BETWEEN:
		values := reflects.ToSlice(cond.Value)
		return bson.M{"$gte": values[0], "$lte": values[1]}
	case query.CONTAIN:
		return bson.M{"$elemMatch": cond.Value}
	case query.HAS:
		return bson.M{"$all": reflects.ToSlice(cond.Value)}
	case query.ALL:
		return bson.M{"$all": reflects.ToSlice(cond.Value)}
	default:
		return bson.M{}
	}
}
