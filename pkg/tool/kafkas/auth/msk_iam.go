package auth

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"
	"github.com/sirupsen/logrus"
)

const AuthIAM = "iam"

// mskTokenTTL token 复用期。AWS MSK IAM token 有效期 15 分钟，
// 这里取 14 分钟留 1 分钟提前续期余量。
const mskTokenTTL = 14 * time.Minute

// ApplyIAM 给 sarama.Config 注入 MSK IAM SASL/OAUTHBEARER + TLS。
// cluster 用于日志区分多集群场景；空串不影响功能，仅日志缺 cluster 字段。
func ApplyIAM(kafkaConfig *sarama.Config, cluster string) error {
	kafkaConfig.Net.SASL.Enable = true
	kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	kafkaConfig.Net.SASL.TokenProvider = &MSKTokenProvider{cluster: cluster}
	kafkaConfig.Net.TLS.Enable = true
	kafkaConfig.Net.TLS.Config = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	return nil
}

func ApplyTLS(kafkaConfig *sarama.Config) {
	kafkaConfig.Net.TLS.Enable = true
	kafkaConfig.Net.TLS.Config = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
}

// MSKTokenProvider 实现 sarama.AccessTokenProvider，按 14 分钟 TTL 缓存 IAM 签名 token。
// sarama 每次新 broker 连接 / SASL 重协商都调一次 Token()，无缓存会反复触发 STS 签名 RPC。
type MSKTokenProvider struct {
	cluster string

	mu     sync.Mutex
	cached string
	expire time.Time
}

func (t *MSKTokenProvider) Token() (*sarama.AccessToken, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cached != "" && time.Now().Before(t.expire) {
		return &sarama.AccessToken{Token: t.cached}, nil
	}

	token, _, err := signer.GenerateAuthToken(context.Background(), "us-west-2")
	if err != nil {
		logrus.WithField("cluster", t.cluster).WithError(err).Error("MSK IAM token generation failed")
		return nil, err
	}
	t.cached = token
	t.expire = time.Now().Add(mskTokenTTL)
	return &sarama.AccessToken{Token: token}, nil
}
