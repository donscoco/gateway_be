package http_proxy_middleware

import (
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/module/counter"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appInterface, ok := c.Get("app")
		if !ok {
			c.Next()
			return
		}
		appInfo := appInterface.(*dao.App)
		appCounter, err := counter.FlowCounterHandler.GetCounter(bl.FlowAppPrefix + appInfo.AppID)
		if err != nil {
			bl.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		appCounter.Increase()
		if appInfo.Qpd > 0 && appCounter.TotalCount > appInfo.Qpd {
			bl.ResponseError(c, 2003, errors.New(fmt.Sprintf("租户日请求量限流 limit:%v current:%v", appInfo.Qpd, appCounter.TotalCount)))
			c.Abort()
			return
		}
		c.Next()
	}
}
