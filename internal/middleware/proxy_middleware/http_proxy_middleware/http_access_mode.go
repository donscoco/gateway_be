package http_proxy_middleware

import (
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/module/service_manager"
	"github.com/gin-gonic/gin"
)

// 匹配接入方式 基于请求信息
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := service_manager.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			bl.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		//fmt.Println("matched service",public.Obj2Json(service))
		c.Set("service", service)
		c.Next()
	}
}
