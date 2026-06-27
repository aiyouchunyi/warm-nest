// Package query @Author larry
// File query_operator.go
// @Date 2024/6/6 16:35:00
// @Desc 查询操作符
package query

type Operator string

const (
	EQ      Operator = "EQ"      // 等于
	NEQ     Operator = "NEQ"     // 不等于
	GT      Operator = "GT"      // 大于
	LT      Operator = "LT"      // 小于
	GTE     Operator = "GTE"     // 大于等于
	LTE     Operator = "LTE"     // 小于等于
	IN      Operator = "IN"      // 包含
	NIN     Operator = "NIN"     // 不包含
	LIKE    Operator = "LIKE"    // 匹配
	NLIKE   Operator = "NLIKE"   // 不匹配
	NIL     Operator = "NIL"     // 空
	NNIL    Operator = "NNIL"    // 非空
	BETWEEN Operator = "BETWEEN" // 介于
	CONTAIN Operator = "CONTAIN" // 包含
	HAS     Operator = "HAS"     // JSON包含其中一个
	ALL     Operator = "ALL"     // JSON全部包含
	JEQ     Operator = "JEQ"     // JSON等于
	JLIKE   Operator = "JLIKE"   // JSON匹配
	JNLIKE  Operator = "JNLIKE"  // JSON不匹配
	JNIL    Operator = "JNIL"    // JSON为空
	JNNIL   Operator = "JNNIL"   // JSON非空

)
