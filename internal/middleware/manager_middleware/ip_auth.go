package manager_middleware

import (
	"errors"
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/gin-gonic/gin"
)

func IPAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isMatched := false

		allow_ip := make([]string, 0, 16)
		iron_config.Conf.GetByScan("/http/allow_ip", &allow_ip)

		for _, host := range allow_ip {
			if c.ClientIP() == host {
				isMatched = true
			}
		}
		if !isMatched {
			bl.ResponseError(c, bl.InternalErrorCode, errors.New(fmt.Sprintf("%v, not in iplist", c.ClientIP())))
			c.Abort()
			return
		}
		c.Next()
	}
}
