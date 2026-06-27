// Package consumer @Author larry
// @Date 2025/8/8 15:17
// @Desc

package consumer

import (
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type MessageHandler interface {
	Name() string                                     // 消息处理器名称
	Cluster() string                                  // 订阅的集群，空串视作 default
	Topic() string                                    // 订阅的 topic
	BeforeProcess(msg *sarama.ConsumerMessage)        // 事前处理
	Process(msg *sarama.ConsumerMessage) error        // 消息处理
	OnSuccess(msg *sarama.ConsumerMessage)            // 处理成功回调
	OnFailure(msg *sarama.ConsumerMessage, err error) // 处理失败回调
}

type DefaultMessageHandler struct {
}

func (h *DefaultMessageHandler) Name() string {
	return "DefaultMessageHandler"
}

func (h *DefaultMessageHandler) Cluster() string {
	return "" // 默认 default 集群，需子类覆盖以订阅其他集群
}

func (h *DefaultMessageHandler) Topic() string {
	return "" // 默认不订阅任何 topic，需子类覆盖
}

func (h *DefaultMessageHandler) BeforeProcess(msg *sarama.ConsumerMessage) {
	logrus.WithFields(logrus.Fields{
		"cluster":   h.Cluster(),
		"name":      h.Name(),
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"key":       string(msg.Key),
	}).Info("Kafka message prepared for processing!")
}

func (h *DefaultMessageHandler) Process(msg *sarama.ConsumerMessage) error {
	logrus.WithFields(logrus.Fields{
		"cluster":   h.Cluster(),
		"name":      h.Name(),
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"key":       string(msg.Key),
		"value":     string(msg.Value),
	}).Info("Kafka message process finished!")
	return nil
}

func (h *DefaultMessageHandler) OnSuccess(msg *sarama.ConsumerMessage) {
	logrus.WithFields(logrus.Fields{
		"cluster":   h.Cluster(),
		"name":      h.Name(),
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"key":       string(msg.Key),
	}).Info("Kafka message process success!")
}

func (h *DefaultMessageHandler) OnFailure(msg *sarama.ConsumerMessage, err error) {
	logrus.WithFields(logrus.Fields{
		"cluster":   h.Cluster(),
		"name":      h.Name(),
		"topic":     msg.Topic,
		"partition": msg.Partition,
		"offset":    msg.Offset,
		"key":       string(msg.Key),
	}).WithError(err).Error("Kafka message process failed!")
}
