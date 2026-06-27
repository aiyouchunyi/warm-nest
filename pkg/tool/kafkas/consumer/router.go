// Package consumer @Author larry
// @Date 2026/4/28
// @Desc Message router abstraction

package consumer

import (
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// Router defines the interface for routing messages to handlers
type Router interface {
	// Route returns the list of handlers that should process the given message
	Route(msg *sarama.ConsumerMessage) []MessageHandler
	// Topics returns all topics managed by this router
	Topics() []string
}

// TopicRouter routes messages based on their topic
type TopicRouter struct {
	cluster     string
	handlersMap map[string][]MessageHandler
	topics      []string
}

// NewTopicRouter creates a router that dispatches messages by topic.
// cluster 用于日志区分多集群场景；空串视作 default。
func NewTopicRouter(cluster string, handlersMap map[string][]MessageHandler) *TopicRouter {
	if len(handlersMap) == 0 {
		logrus.WithField("cluster", cluster).Warn("creating TopicRouter with empty handlersMap")
	}

	topics := make([]string, 0, len(handlersMap))
	for topic := range handlersMap {
		topics = append(topics, topic)
	}

	return &TopicRouter{
		cluster:     cluster,
		handlersMap: handlersMap,
		topics:      topics,
	}
}

func (r *TopicRouter) Route(msg *sarama.ConsumerMessage) []MessageHandler {
	handlers, ok := r.handlersMap[msg.Topic]
	if !ok {
		logrus.WithFields(logrus.Fields{
			"cluster": r.cluster,
			"topic":   msg.Topic,
		}).Warn("no handlers found for topic")
		return nil
	}
	return handlers
}

func (r *TopicRouter) Topics() []string {
	// Return a copy to prevent external modification
	result := make([]string, len(r.topics))
	copy(result, r.topics)
	return result
}
