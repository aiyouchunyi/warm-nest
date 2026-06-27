// Package query @Author larry
// File QueryReq.go
// @Date 2024/5/30 17:57:00
// @Desc 查询请求
package query

// QueryReq 查询请求
type QueryReq struct {
	Page       int         `gorm:"comment:页码" form:"page" default:"1" json:"page" validate:"omitempty,min=1"`
	Size       int         `gorm:"comment:每页数量" form:"size" default:"100" json:"size" validate:"omitempty,min=-1,max=1000"`
	Columns    []string    `gorm:"comment:过滤字段列表" json:"columns"`
	Conditions []Condition `gorm:"comment:条件列表" json:"conditions"`
	Customs    []Custom    `gorm:"comment:自定义条件列表" json:"customs"`
	Sort       []Sort      `gorm:"comment:排序列表" json:"sort"`
}

func New() *QueryReq {
	return &QueryReq{
		Page: 1,
		Size: 1000,
	}
}

func (query *QueryReq) AddColumns(columns ...string) *QueryReq {
	if len(columns) == 0 {
		return query
	}
	query.Columns = append(query.Columns, columns...)
	return query
}

// AddCondition 添加条件
func (query *QueryReq) AddCondition(condition Condition) *QueryReq {
	query.Conditions = append(query.Conditions, condition)
	return query
}

// AddConditionIfPresent 当 value 非空时添加 EQ 条件
func (query *QueryReq) AddConditionIfPresent(column string, value string) *QueryReq {
	if value != "" {
		query.Conditions = append(query.Conditions, NewCondition(column, EQ, value))
	}
	return query
}

// AddCustom 添加自定义条件
func (query *QueryReq) AddCustom(custom Custom) *QueryReq {
	query.Customs = append(query.Customs, custom)
	return query
}

func (query *QueryReq) AddSort(sort Sort) *QueryReq {
	query.Sort = append(query.Sort, sort)
	return query
}

func (query *QueryReq) SetPage(page int) *QueryReq {
	query.Page = page
	return query
}

func (query *QueryReq) SetSize(size int) *QueryReq {
	query.Size = size
	return query
}

// LimitOne 设置每页数量
func (query *QueryReq) LimitOne() {
	query.Size = 1
}

func (query *QueryReq) All() {
	query.Size = -1
}
