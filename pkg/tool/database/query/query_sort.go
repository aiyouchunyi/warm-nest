// Package query @Author larry
// File query_sort.go
// @Date 2024/6/6 16:35:00
// @Desc 查询排序
package query

type SortType string

const (
	SortTypeASC  SortType = "ASC"  // 升序
	SortTypeDESC SortType = "DESC" // 降序
)

type Sort struct {
	Column string   `gorm:"comment:字段名" json:"column"`
	Order  SortType `gorm:"comment:排序方式" json:"order"`
}
