package http_proxy_middleware

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/module/service_manager"
	"github.com/donscoco/gateway_be/internal/reverse_proxy"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// 匹配接入方式 基于请求信息
func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, ok := c.Get("service")
		if !ok {
			bl.ResponseError(c, 2001, errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail := serverInterface.(*dao.ServiceDetail)

		lb, err := service_manager.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err != nil {
			bl.ResponseError(c, 2002, err)
			c.Abort()
			return
		}
		trans, err := service_manager.TransportorHandler.GetTrans(serviceDetail)
		if err != nil {
			bl.ResponseError(c, 2003, err)
			c.Abort()
			return
		}
		//middleware.ResponseSuccess(c,"ok")
		//return
		//创建 reverseproxy
		//使用 reverseproxy.ServerHTTP(c.Request,c.Response)
		proxy := reverse_proxy.NewLoadBalanceReverseProxy(c, lb, trans)
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Abort()
		return
	}
}
