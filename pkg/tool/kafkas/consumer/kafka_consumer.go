// Package consumer @Author larry
// @Date 2025/8/8 15:18
// @Desc

package consumer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/tool/concurrent"
	"warm-nest/pkg/tool/kafkas/auth"
	"warm-nest/pkg/tool/kafkas/config"
	"warm-nest/pkg/tool/kafkas/monitor"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/times"
)

type KafkaConsumer struct {
	cluster      string
	brokers      []string
	group        string
	router       Router
	cachedTopics []string // cached to avoid repeated allocations
	consumer     sarama.ConsumerGroup
	config       *sarama.Config
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

var enableMu sync.Mutex
var enableCalled bool

func EnableConsumer(params ...any) (func(), error) {
	enableMu.Lock()
	defer enableMu.Unlock()
	if enableCalled {
		return nil, fmt.Errorf("kafka consumer can only be enabled once")
	}
	if configs.IsLocal() {
		logrus.Warn("kafka consumer is not supported in local environment")
		return nil, nil
	}
	if len(params) != 2 {
		return nil, fmt.Errorf("kafka consumer requires 2 parameters: group, handlers")
	}
	group, ok1 := params[0].(string)
	handlers, ok2 := params[1].([]MessageHandler)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid parameter types for kafka consumer initialization: group string, handlers []MessageHandler")
	}

	// 按 handler.Cluster() 分桶，空串归入 default
	buckets := make(map[string][]MessageHandler)
	for _, h := range handlers {
		cluster := h.Cluster()
		if cluster == "" {
			cluster = config.DefaultCluster
		}
		buckets[cluster] = append(buckets[cluster], h)
	}

	started := make([]*KafkaConsumer, 0, len(buckets))
	stopAll := func() {
		for _, c := range started {
			c.Stop()
		}
	}
	for cluster, bucket := range buckets {
		clusterCfg, ok := config.GetCluster(cluster)
		if !ok {
			stopAll()
			return nil, fmt.Errorf("kafka consumer: cluster %q not configured", cluster)
		}
		consumer := New(clusterCfg, cluster, group, bucket)
		if err := consumer.Run(); err != nil {
			stopAll()
			return nil, fmt.Errorf("kafka consumer cluster %q run failed: %w", cluster, err)
		}
		started = append(started, consumer)
	}
	enableCalled = true
	return stopAll, nil
}

func New(cfg config.ClusterConfig, cluster string, group string, handlers []MessageHandler) *KafkaConsumer {
	brokers := strings.Split(cfg.Urls, ",")
	ctx, cancel := context.WithCancel(context.Background())
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true

	if cfg.Auth == auth.AuthIAM {
		_ = auth.ApplyIAM(kafkaConfig, cluster)
	} else if cfg.TLS {
		auth.ApplyTLS(kafkaConfig)
	}

	// Aggregate handlers by topic
	handlersMap := make(map[string][]MessageHandler)
	for _, h := range handlers {
		topic := h.Topic()
		if topic != "" {
			handlersMap[topic] = append(handlersMap[topic], h)
		}
	}

	router := NewTopicRouter(cluster, handlersMap)
	return &KafkaConsumer{
		cluster:      cluster,
		brokers:      brokers,
		group:        group,
		router:       router,
		cachedTopics: router.Topics(),
		config:       kafkaConfig,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// log 返回带 cluster + group 字段的 logrus entry，所有日志统一从这里派生
func (kc *KafkaConsumer) log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"cluster": kc.cluster,
		"group":   kc.group,
	})
}

func (kc *KafkaConsumer) Run() error {
	log := kc.log()
	log.WithField("topics", kc.cachedTopics).Info("kafka consumer starting..")
	start := times.UnixMilli()
	if err := kc.Start(); err != nil {
		log.WithError(err).Error("kafka consumer start error")
		return err
	}
	monitor.GetKafkaMonitor().RegisterCluster(kc.cluster, kc.monitorName(), kc.brokers, kc.config)
	log.WithField("delay_ms", times.Gap(start)).Info("kafka consumer started...")
	return nil
}

func (kc *KafkaConsumer) monitorName() string {
	return "consumer:" + kc.cluster + ":" + kc.group
}

func (kc *KafkaConsumer) Start() error {
	log := kc.log()
	if len(kc.brokers) == 0 || kc.group == "" || len(kc.cachedTopics) == 0 {
		log.WithFields(logrus.Fields{
			"brokers": len(kc.brokers),
			"topics":  len(kc.cachedTopics),
		}).Error("kafka consumer configuration is incomplete")
		return fmt.Errorf("kafka consumer configuration is incomplete")
	}

	cg, err := sarama.NewConsumerGroup(kc.brokers, kc.group, kc.config)
	if err != nil {
		return err
	}
	kc.consumer = cg

	// 可选：如果 kc.config.Consumer.Return.Errors 为 true，建议 drain 错误通道
	if kc.config != nil && kc.config.Consumer.Return.Errors {
		kc.wg.Add(1)
		go func() {
			defer kc.wg.Done()
			for {
				select {
				case err, ok := <-kc.consumer.Errors():
					if !ok {
						return
					}
					if err != nil {
						entry := log.WithError(err)
						// sarama 的 ConsumerError 自带 Topic/Partition，提取后单独成字段方便排查
						isEOF := false
						if ce, ok := err.(*sarama.ConsumerError); ok {
							entry = entry.WithFields(logrus.Fields{
								"topic":     ce.Topic,
								"partition": ce.Partition,
							})
							isEOF = errors.Is(ce.Err, io.EOF)
						}
						// MSK broker 端 idle ~60min 主动断 TCP 是常态，sarama 内部会
						// abort+reconnect 自愈，<1s 恢复消费；降级 WARN 避免污染告警。
						if isEOF {
							entry.Warn("kafka consumer group eof, sarama auto-reconnecting")
						} else {
							entry.Error("kafka consumer group error")
						}
					}
				case <-kc.ctx.Done():
					return
				}
			}
		}()
	}

	kc.wg.Add(1)
	go func() {
		defer kc.wg.Done()

		backoff := 1 * time.Second
		const maxBackoff = 30 * time.Second

		for {
			// sarama 要求重复调用 Consume，内部会处理 rebalance 等
			if err := kc.consumer.Consume(kc.ctx, kc.cachedTopics, kc); err != nil {
				// 如果不是因为 ctx 取消导致的错误，则退避后重试
				if kc.ctx.Err() == nil {
					log.WithError(err).Error("kafka consumer consume error")
					// 简单指数退避 + 抖动
					jitter := time.Duration(rand.Int63n(int64(backoff/2 + time.Millisecond)))
					time.Sleep(backoff + jitter)
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
					continue
				}
			}

			// 会话正常结束或 ctx 取消
			if kc.ctx.Err() != nil {
				return
			}
			// 会话正常返回（如 rebalance 结束），重置退避继续下一轮
			backoff = 1 * time.Second
		}
	}()
	return nil
}

// Stop 停止 Kafka 消费者
// 注意：此方法会取消上下文，导致 ConsumeClaim 中的循环退出
func (kc *KafkaConsumer) Stop() {
	log := kc.log()
	kc.cancel()
	monitor.GetKafkaMonitor().Unregister(kc.monitorName())
	if kc.consumer != nil {
		if err := kc.consumer.Close(); err != nil {
			log.WithError(err).Error("kafka consumer close error!")
		} else {
			log.Info("kafka consumer closed success!")
		}
	}

	done := make(chan struct{})
	go func() { kc.wg.Wait(); close(done) }()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Warn("kafka consumer stop timeout, continue shutdown")
	}

	log.Info("kafka consumer stopped...")
}

func (kc *KafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (kc *KafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (kc *KafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	clusterLog := kc.log()
	for {
		select {
		case <-session.Context().Done():
			clusterLog.Info("kafka consumer session context done")
			return nil

		case msg, ok := <-claim.Messages():
			if !ok {
				clusterLog.Info("kafka consumer claim messages channel closed")
				return nil
			}

			now := times.UnixMilli()
			log := clusterLog.WithFields(logrus.Fields{
				"topic":     msg.Topic,
				"partition": msg.Partition,
				"offset":    msg.Offset,
				"key":       string(msg.Key),
				"time":      now,
			})

			func() {
				// 仅保护当前 goroutine，按你的要求不对每个 handler 额外做 panic recover
				defer func() {
					if r := recover(); r != nil {
						log.WithField("recover", r).Error("kafka consumer panic, auto recovered")
					}
				}()

				// Route message to appropriate handlers
				handlers := kc.router.Route(msg)
				if len(handlers) == 0 {
					log.Warn("kafka consumer message has no handlers for this topic, marking as processed")
					session.MarkMessage(msg, fmt.Sprintf("ok ts=%d", now))
					return
				}

				limit := min(4, len(handlers))
				group := concurrent.NewCGroup(limit)

				for i := range handlers {
					h := handlers[i]
					group.Go(func() error {
						h.BeforeProcess(msg)
						if err := h.Process(msg); err != nil {
							h.OnFailure(msg, err)
							return err
						}
						h.OnSuccess(msg)
						return nil
					})
				}

				if err := group.Wait(); err != nil {
					// 有任一 handler 失败则不标记，等待自动重试或由上层策略处理
					log.WithError(err).Error("kafka consumer message process error")
					return
				}

				log.Info("kafka consumer message process success")
				// 仅标记成功消息；提交由 AutoCommit.Interval 触发
				session.MarkMessage(msg, fmt.Sprintf("ok ts=%d", now))
			}()
		}
	}
}
