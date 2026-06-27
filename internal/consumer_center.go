// Package internal @Author larry
// @Date 2025/8/8 17:03
// @Desc

package internal

import (
	"warm-nest/pkg/tool/kafkas/consumer"
)

// RegisterConsumer 注册 Kafka 消费者
func RegisterConsumer() []consumer.MessageHandler {
	return []consumer.MessageHandler{
		//listener.GetReserveAdjustHandler(),
	}
}
