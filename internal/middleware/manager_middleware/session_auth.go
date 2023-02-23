package manager_middleware

import (
	"errors"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if adminInfo, ok := session.Get(bl.AdminSessionInfoKey).(string); !ok || adminInfo == "" {
			bl.ResponseError(c, bl.InternalErrorCode, errors.New("user not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
