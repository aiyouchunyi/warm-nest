// Package auth @Author larry
// @Date 2025/5/22 15:55
// @Desc

package constant

const (
	AuthKindUser = "USER"
	AuthKindRole = "ROLE"
	AuthKindExpr = "EXPR"
)

const (
	AuditNotCreator  = "NOT_CREATOR"  // 表达式不属于创建者
	AuditCreator     = "CREATOR"      // 表达式属于创建者
	AuditNotPrevious = "NOT_PREVIOUS" // 表达式不属于上一个审批人
)
