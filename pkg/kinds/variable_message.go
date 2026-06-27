// Package msg @Author larry
// @Date 2025/8/18 11:15
// @Desc

package kinds

import (
	"fmt"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"

	"warm-nest/pkg/utils/reflects"
)

type VariableMessage struct {
	Data *orderedmap.OrderedMap[string, any]
}

func New() *VariableMessage {
	return &VariableMessage{
		Data: orderedmap.New[string, any](),
	}
}

func (m *VariableMessage) Add(key string, value any) *VariableMessage {
	m.Data.Set(key, value)
	return m
}

func (m *VariableMessage) LineString() string {
	var result strings.Builder
	for pair := m.Data.Oldest(); pair != nil; pair = pair.Next() {
		if result.Len() > 0 {
			result.WriteByte('\n')
		}
		result.WriteString(fmt.Sprintf("%s: %s", pair.Key, reflects.ToString(pair.Value)))
	}
	return result.String()
}
