// Package constant @Author larry
// @Date 2025/4/9 16:18
// @Desc

package constant

const (
	EventActionApprove  = "APPROVE"  // 通过(肯定到往后节点)
	EventActionReject   = "REJECT"   // 拒绝(否定到往后节点)
	EventActionTransfer = "TRANSFER" // 转交(中性到往后节点)
	EventActionDismiss  = "DISMISS"  // 驳回(到往前非上一个节点)
	EventActionRollback = "ROLLBACK" // 回退(到上一个节点状态)
	EventActionCancel   = "CANCEL"   // 取消(到停止状态)
	EventActionForce    = "FORCE"    // 强制[自动节点和手动节点](不经过条件直接流转到指定节点)
)

var EventActionMap = map[string]string{
	EventActionApprove:  "通过",
	EventActionReject:   "拒绝",
	EventActionTransfer: "转交",
	EventActionDismiss:  "驳回",
	EventActionRollback: "回滚",
	EventActionCancel:   "取消",
	EventActionForce:    "强制",
}
