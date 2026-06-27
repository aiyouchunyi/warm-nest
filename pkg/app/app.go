package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/log"
	"warm-nest/pkg/app/web"
	"warm-nest/pkg/tool/database"
	"warm-nest/pkg/tool/loaders"
	"warm-nest/pkg/utils/times"
)

type Do func(params ...interface{}) error

type DoCtx func(params ...interface{}) (func(), error)

type App struct {
	name      string
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	start     int64
	waitClose int64
}

func New(name string) *App {
	logrus.Infof(" %s app starting...", name)
	log.Init()
	// 把 loaders 的默认告警器接到日志：loaders 自身不依赖具体告警实现（避免循环），由此处注入。
	loaders.SetDefaultAlert(func(msg string) { logrus.Error(msg) })
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		name:      name,
		ctx:       ctx,
		cancel:    cancel,
		start:     times.UnixMilli(),
		waitClose: 60,
	}
}

// Web 启动web服务
func (a *App) Web(controllers []web.Controller) *App {
	a.DoCtx(web.EnableWeb, controllers)
	return a
}

func (a *App) DB(drivers ...string) *App {
	params := make([]interface{}, len(drivers))
	for i, driver := range drivers {
		params[i] = driver
	}
	a.Do(database.EnableDriver, params...)
	return a
}

// Do 执行自定义函数
func (a *App) Do(do Do, params ...interface{}) *App {
	if err := do(params...); err != nil {
		logrus.WithError(err).Error("do task failed!")
	}
	return a
}

func (a *App) DoCtx(do DoCtx, params ...interface{}) *App {
	go func() {
		stop, err := do(params...)
		if err != nil {
			logrus.WithError(err).Error("do ctx task failed!")
		}
		if stop == nil {
			return
		}
		a.wg.Add(1)

		<-a.ctx.Done()
		stop()
		a.wg.Done()
	}()
	return a
}

func (a *App) AsyncDo(do Do, params ...interface{}) *App {
	go func() {
		a.Do(do, params...)
	}()
	return a
}

func (a *App) WaitClose(seconds int64) *App {
	a.waitClose = seconds
	return a
}

// Run 启动应用
func (a *App) Run() {

	logrus.Infof("%s app started... delay: %d ms", a.name, times.Gap(a.start))

	// 等待终止信号
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	// 开始关闭应用
	a.cancel()

	// 等待所有任务完成，设置超时时间
	timeout := time.After(time.Duration(a.waitClose) * time.Second)
	done := make(chan struct{})

	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logrus.Infof("%s app stopped...", a.name)
	case <-timeout:
		logrus.Errorf("%s app stop timeout, force stopped...", a.name)
	}
}
