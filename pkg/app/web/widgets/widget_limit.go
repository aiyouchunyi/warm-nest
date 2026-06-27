// Package widget @Author larry
// @Date 2025/2/6 09:55
// @Desc

package widgets

import (
	"time"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/app/web/code"
	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/tool/ratelimits"
)

func Limit(limit int) WidgetFunc {
	return func(c *gin.Context) error {
		ratelimits.Wait(ratelimits.RateKey("controller", c.Request.URL.Path), limit)
		return nil
	}
}

func LimitDuration(limit int, maxDuration time.Duration) WidgetFunc {
	return func(c *gin.Context) error {
		if !ratelimits.WaitDuration(ratelimits.RateKey("controller", c.Request.URL.Path), limit, maxDuration) {
			return errors.NewWithArgs(code.WebRateLimit)
		}
		return nil
	}
}

func UserLimit(limit int) WidgetFunc {
	return func(c *gin.Context) error {
		webContext := context.GetContext(c)
		ratelimits.Wait(ratelimits.RateKey("controller", webContext.ReqUser, c.Request.URL.Path), limit)
		return nil
	}
}

func UserLimitDuration(limit int, maxDuration time.Duration) WidgetFunc {
	return func(c *gin.Context) error {
		webContext := context.GetContext(c)
		if !ratelimits.WaitDuration(ratelimits.RateKey("controller", webContext.ReqUser, c.Request.URL.Path), limit, maxDuration) {
			return errors.NewWithArgs(code.WebRateLimit)
		}
		return nil
	}
}
