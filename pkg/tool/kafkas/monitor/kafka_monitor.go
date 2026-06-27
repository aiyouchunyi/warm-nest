// Package monitor @Author larry
// @Date 2025/8/11 19:37
// @Desc

package monitor

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/configs"
)

// entry 描述一个被监控的 kafka 注册项
type entry struct {
	cluster string         // 可选：所属集群名，用于日志区分（default / financial 等）
	brokers []string       // broker 地址列表
	cfg     *sarama.Config // 上层使用的 sarama.Config（用于复用 TLS/SASL 配置）；nil 表示明文
	rr      int            // 该 entry 内部的 broker 轮询索引
}

// KafkaMonitor 是全局单例：仅起一个 goroutine 做健康检查
type KafkaMonitor struct {
	mu sync.Mutex
	wg sync.WaitGroup

	interval time.Duration
	monitors map[string]*entry
	stopCh   chan struct{}
	running  bool
	rrIndex  int
}

var kafkaMonitor *KafkaMonitor
var kafkaMonitorOnce sync.Once

// GetKafkaMonitor 返回单例实例
func GetKafkaMonitor() *KafkaMonitor {
	kafkaMonitorOnce.Do(func() {
		kafkaMonitor = &KafkaMonitor{
			interval: 60 * time.Second, // 默认检查周期
			rrIndex:  0,
			monitors: make(map[string]*entry),
		}
	})
	return kafkaMonitor
}

// Register 启动健康检查（兼容旧签名，不区分 cluster，按明文连接）。
// 新代码请使用 RegisterCluster 传入 cluster 名与 sarama.Config。
func (m *KafkaMonitor) Register(name string, brokers []string) {
	m.RegisterCluster("", name, brokers, nil)
}

// RegisterCluster 启动健康检查（幂等）。
// - cluster: 集群名（用于日志区分，可空）
// - name: 注册项唯一键，与 Unregister 配对
// - brokers: broker 地址列表
// - cfg: 上层 sarama.Config，用于复用 TLS/SASL；nil 表示明文连接
// - 第一次调用时启动后台 goroutine；后续调用仅合并注册项。
func (m *KafkaMonitor) RegisterCluster(cluster string, name string, brokers []string, cfg *sarama.Config) {
	if configs.IsLocal() {
		return
	}
	if len(brokers) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.monitors[name]; !exists {
		m.monitors[name] = &entry{cluster: cluster, brokers: brokers, cfg: cfg}
	}

	// 后台 goroutine 未运行时启动；支持 Unregister 全部释放后再次注册
	if !m.running {
		m.stopCh = make(chan struct{})
		m.running = true
		m.wg.Add(1)
		go m.loop(m.stopCh)
	}
}

// Unregister 停止健康检查（幂等）。
// 当所有注册项被释放后，停止后台 goroutine；之后再次 Register 会重新启动。
func (m *KafkaMonitor) Unregister(name string) {
	m.mu.Lock()
	if _, exists := m.monitors[name]; exists {
		delete(m.monitors, name)
	}
	if len(m.monitors) > 0 || !m.running {
		m.mu.Unlock()
		return
	}
	stopCh := m.stopCh
	m.running = false
	m.mu.Unlock()

	close(stopCh)
	m.wg.Wait()
}

// loop 接收 stopCh 作为参数而非读 m.stopCh，避免重启时新旧 channel 混用导致 race
func (m *KafkaMonitor) loop(stopCh <-chan struct{}) {
	defer m.wg.Done()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	logrus.Info("kafka monitor: started")
	for {
		select {
		case <-ticker.C:
			cluster, _, err := m.checkOnce()
			log := logrus.WithField("cluster", cluster)
			if err != nil {
				log.WithError(err).Warn("kafka monitor: check failed")
			} else {
				log.Info("kafka monitor: ok")
			}
		case <-stopCh:
			logrus.Info("kafka monitor: stopped")
			return
		}
	}
}

// buildBrokerCfg 基于 entry.cfg 构造一个 monitor 专用的 sarama.Config，
// 复用 TLS/SASL 字段，但用更短的超时确保健康检查快速失败，避免污染上层配置。
func buildBrokerCfg(src *sarama.Config) *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Net.DialTimeout = 2 * time.Second
	cfg.Net.ReadTimeout = 2 * time.Second
	cfg.Net.WriteTimeout = 2 * time.Second
	if src != nil {
		cfg.Net.SASL = src.Net.SASL
		cfg.Net.TLS = src.Net.TLS
	}
	return cfg
}

func (m *KafkaMonitor) checkOnce() (string, string, error) {
	m.mu.Lock()
	if len(m.monitors) == 0 {
		m.mu.Unlock()
		return "", "", fmt.Errorf("no brokers configured")
	}
	names := make([]string, 0, len(m.monitors))
	for k := range m.monitors {
		names = append(names, k)
	}
	sort.Strings(names) // 稳定顺序，避免 map 迭代乱序
	picked := m.monitors[names[m.rrIndex%len(names)]]
	m.rrIndex = (m.rrIndex + 1) % len(names)
	addr := picked.brokers[picked.rr%len(picked.brokers)]
	picked.rr = (picked.rr + 1) % len(picked.brokers)
	cluster := picked.cluster
	cfg := buildBrokerCfg(picked.cfg)
	m.mu.Unlock()

	b := sarama.NewBroker(addr)
	if err := b.Open(cfg); err != nil {
		return cluster, addr, err
	}
	_, err := b.ApiVersions(&sarama.ApiVersionsRequest{})
	_ = b.Close()
	if err != nil {
		return cluster, addr, err
	}
	return cluster, addr, nil
}
