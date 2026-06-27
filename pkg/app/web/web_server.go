package web

import (
	context2 "context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/web/config"
	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/app/web/interceptor"
	"warm-nest/pkg/app/web/utils"
	"warm-nest/pkg/utils/times"
)

type Server struct {
	controllers []Controller
	server      *http.Server
}

var enableMu sync.Mutex
var enableCalled bool

func EnableWeb(params ...interface{}) (func(), error) {
	enableMu.Lock()
	defer enableMu.Unlock()
	if enableCalled {
		return nil, fmt.Errorf("web server can only be enabled once")
	}
	if len(params) != 1 {
		return nil, fmt.Errorf("web server requires 1 parameter: controllers")
	}
	controllers, ok := params[0].([]Controller)
	if !ok {
		return nil, fmt.Errorf("invalid parameter type for web server initialization: controllers")
	}
	webServer := New(controllers)
	webServer.Run()
	return webServer.Stop, nil
}

func New(controllers []Controller) *Server {
	return &Server{
		controllers: controllers,
	}
}

func (s *Server) Run() {
	logrus.Infof("web server starting... host: %s", config.ServerConf().Host)
	start := times.UnixMilli()
	router := gin.New()
	router.Use(interceptor.Context(config.ServerConf().AuthEnabled))
	router.Use(interceptor.RequestLog())
	router.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		webContext := context.GetContext(c)
		logrus.WithField("reqId", webContext.ReqId).WithFields(logrus.Fields{
			"user": webContext.ReqUser,
			"url":  utils.UrlInfo(c),
			"err":  err,
		}).Error("gin recovery")
		logrus.Error(fmt.Sprintf("[Web-Controller] Recovered from panic in: %v %s", err, utils.UrlInfo(c)))
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.Any("/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"pong": time.Now().UnixMilli(),
		})
	})

	for _, controller := range s.controllers {
		controller.Router(router)
	}

	// 初始化性能分析
	s.initProf(router)

	// 启动web服务
	s.server = &http.Server{
		Addr:    config.ServerConf().Host,
		Handler: router,
	}

	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("web server ListenAndServe error: " + err.Error())
		}
	}()
	logrus.Infof("web server started... delay: %d ms", times.Gap(start))
}

// 初始化性能分析
func (s *Server) initProf(engine *gin.Engine) {
	if !config.ServerConf().PprofEnabled {
		return
	}
	if config.ServerConf().PprofPathPrefix != "" {
		ginpprof.WrapGroup(engine.Group(config.ServerConf().PprofPathPrefix))
	} else {
		ginpprof.Wrapper(engine)
	}
}

func (s *Server) Stop() {
	ctxShutDown, cancel := context2.WithTimeout(context2.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctxShutDown); err != nil {
		logrus.Errorf("web server shutdown: %+v", err)
	}
	if err := s.server.Close(); err != nil {
		logrus.Errorf("web server close: %+v", err)
	}
	logrus.Infof("web server stoped...")
}
