// Package model @Author larry
// @Date 2026/06/20
// @Desc 收货地址值对象（结构化 JSON，存 ElderProfile.Address 列）
//
// 设计要点（问题11 子女端结构化地址）：
//   - 地址从单 string 升级为固定字段结构（省/市/区/街道/详细地址 + 收货人 + 收货电话），
//     前端按字段分别录入/校验，比单文本框更可控。测试期不兼容旧单字符串，直接改列类型为 json。
//   - 收货人 ReceiverName / 收货电话 ReceiverPhone 独立于老人本人（RealName/ElderPhone）：
//     支持子女代收（收货人填子女）。领取下单时这两字段拍平进 RewardClaim 快照。
//   - 用强类型 struct + driver.Valuer/sql.Scanner 落 json 列（非 jsons.JSONObject）：
//     字段固定且对外是明确契约，强类型让取字段不易写错、前端 schema 清晰。
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// Address 结构化收货地址（存为 json 列）
type Address struct {
	Province      string `json:"province"`      // 省
	City          string `json:"city"`          // 市
	District      string `json:"district"`      // 区/县
	Street        string `json:"street"`        // 街道/乡镇
	Detail        string `json:"detail"`        // 详细地址（门牌等）
	ReceiverName  string `json:"receiverName"`  // 收货人（可为子女代收，未必=老人真名）
	ReceiverPhone string `json:"receiverPhone"` // 收货电话（未必=老人本人电话）
}

// Value 实现 driver.Valuer：写库时序列化为 json 字符串（空结构体也存 {}，不存 NULL）
func (a Address) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("marshal address: %w", err)
	}
	return string(b), nil
}

// Scan 实现 sql.Scanner：读库时从 json 反序列化；NULL/空串视为零值地址
func (a *Address) Scan(value any) error {
	if value == nil {
		*a = Address{}
		return nil
	}
	var raw []byte
	switch v := value.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("scan address: unsupported type %T", value)
	}
	if len(raw) == 0 {
		*a = Address{}
		return nil
	}
	if err := json.Unmarshal(raw, a); err != nil {
		return fmt.Errorf("unmarshal address: %w", err)
	}
	return nil
}

// MissingShippingField 收货信息是否完整（领取实物前置校验用）：
// 详细地址 + 收货人 + 收货电话三者齐备才算可发货。返回缺失字段名（空=完整）。
func (a Address) MissingShippingField() string {
	if strings.TrimSpace(a.Detail) == "" && strings.TrimSpace(a.Province) == "" {
		return "address" // 地址主体为空
	}
	if strings.TrimSpace(a.ReceiverName) == "" {
		return "receiverName"
	}
	if strings.TrimSpace(a.ReceiverPhone) == "" {
		return "receiverPhone"
	}
	return ""
}
