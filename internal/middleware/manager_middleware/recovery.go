package manager_middleware

import (
	"errors"
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/gin-gonic/gin"
	"runtime/debug"
)

// RecoveryMiddleware捕获所有panic，并且返回错误信息
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				//先做一下日志记录
				fmt.Println(string(debug.Stack()))

				//打印日志，判断配置是否debug，返回对应的响应

				// panic 了要记录下错误和栈状态。方便后面排查,todo 格式最后确定一下
				bl.Errorf("panic recover %+v", map[string]interface{}{
					"error": fmt.Sprint(err),
					"stack": string(debug.Stack()),
				})

				bl.ResponseError(c, 500, errors.New(fmt.Sprint(err)))
				return

				//if bl.ServerMode != "debug" {
				//	bl.ResponseError(c, 500, errors.New("内部错误"))
				//	return
				//} else {
				//	bl.ResponseError(c, 500, errors.New(fmt.Sprint(err)))
				//	return
				//}
			}
		}()
		c.Next()
	}
}
