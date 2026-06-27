// Package mysqls @Author larry
// @Date 2025/11/10 09:46
// @Desc

package mysqls

type MySQLSplit interface {
	GetTableNameByIndex(tableIndex int) string
}
