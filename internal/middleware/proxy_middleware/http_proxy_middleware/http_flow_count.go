package http_proxy_middleware

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/module/counter"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func HTTPFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			bl.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		//统计项 1 全站 2 服务 3 租户
		totalCounter, err := counter.FlowCounterHandler.GetCounter(bl.FlowTotal)
		if err != nil {
			bl.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		totalCounter.Increase()

		//dayCount, _ := totalCounter.GetDayData(time.Now())
		//fmt.Printf("totalCounter qps:%v,dayCount:%v", totalCounter.QPS, dayCount)
		serviceCounter, err := counter.FlowCounterHandler.GetCounter(bl.FlowServicePrefix + serviceDetail.Info.ServiceName)
		if err != nil {
			bl.ResponseError(c, 4001, err)
			c.Abort()
			return
		}
		serviceCounter.Increase()

		//dayServiceCount, _ := serviceCounter.GetDayData(time.Now())
		//fmt.Printf("serviceCounter qps:%v,dayCount:%v", serviceCounter.QPS, dayServiceCount)
		c.Next()
	}
}
