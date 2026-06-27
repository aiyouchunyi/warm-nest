// Package mongo @Author larry
// @Date 2025/1/2 10:05
// @Desc

package mongos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"warm-nest/pkg/utils/times"
)

var client *mongo.Client

func Client() *mongo.Client {
	if client == nil {
		panic("mongo is not start!")
	}
	return client
}

var enableMu sync.Mutex
var enableCalled bool

func EnableMongo(url string) error {
	enableMu.Lock()
	defer enableMu.Unlock()
	if enableCalled {
		return fmt.Errorf("mongo can only be enabled once")
	}
	start := times.UnixMilli()
	logrus.Info("mongo starting...")
	tClient, err := Connect(url, false)
	if err != nil {
		logrus.Errorf("init mongo failed! " + err.Error())
		return err
	}
	client = tClient
	logrus.Infof("mongo started... delay: %d ms", times.Gap(start))
	return nil
}

func Connect(url string, primary bool) (*mongo.Client, error) {
	if url == "" {
		logrus.Errorf("mongo url is empty")
		return nil, fmt.Errorf("mongo url is empty")
	}

	clientOptions := options.Client().ApplyURI(url).
		SetMaxConnIdleTime(10 * time.Minute).
		SetMaxPoolSize(100).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second).
		SetReadPreference(
			func() *readpref.ReadPref {
				if primary {
					return readpref.Primary()
				}
				return readpref.SecondaryPreferred()
			}(),
		).
		SetMonitor(&event.CommandMonitor{
			Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
				logrus.Infof("MongoDB Command Started: %s %v", evt.CommandName, evt.Command)
			},
			Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
				logrus.Infof("MongoDB Command Succeeded: %s (Duration: %v)", evt.CommandName, evt.Duration)
			},
			Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
				logrus.Infof("MongoDB Command Failed: %s (Duration: %v, Error: %v)", evt.CommandName, evt.Duration, evt.Failure)
			},
		})

	// 直接使用 mongo.Connect 创建并连接客户端
	tClient, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		logrus.Errorf("mongo connect failed! " + err.Error())
		return nil, fmt.Errorf("mongo connect failed! %w", err)
	}

	// 设置连接超时并检测 MongoDB 是否存在
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = tClient.Ping(ctx, readpref.Primary()); err != nil {
		logrus.Errorf("mongo ping failed! " + err.Error())
		return nil, fmt.Errorf("mongo ping failed! %w", err)
	}
	return tClient, nil
}
