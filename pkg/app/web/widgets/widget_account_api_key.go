// Package widget @Author larry
// @Date 2025/2/5 19:24
// @Desc

package widgets

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/app/web/code"
)

func AccountApiKey(c *gin.Context) error {
	accountIdStr := c.Query("accountId")
	if accountIdStr == "" {
		return errors.NewWithArgs(code.WebCallParmaErr, "accountId", "不能为空")
	}

	accountId, err := strconv.ParseInt(accountIdStr, 10, 64)
	if err != nil {
		return errors.NewWithArgs(code.WebCallParmaErr, "accountId", "格式错误:"+accountIdStr)
	}

	c.Request.URL.RawQuery = c.Request.URL.RawQuery + "&apiKey=" + fmt.Sprintf("%x", accountId>>32)
	return nil
}
