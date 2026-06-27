// Package query @Author larry
// File Query.go
// @Date 2024/5/31 10:06:00
// @Desc 查询结果
package query

type QueryResult struct {
	Total int64       `gorm:"comment:总数" json:"total"`
	List  interface{} `gorm:"comment:列表" json:"list"`
	Page  int         `gorm:"comment:页数" json:"page"`
	Size  int         `gorm:"comment:每页数量" json:"size"`
}

func Result(total int64, list interface{}, queryReq QueryReq) QueryResult {
	return QueryResult{
		Total: total,
		List:  list,
		Page:  queryReq.Page,
		Size:  queryReq.Size,
	}
}
