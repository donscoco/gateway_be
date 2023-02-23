package manager_middleware

import (
	"bytes"
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

// 请求进入日志
func RequestInLog(c *gin.Context) {
	// 打印请求日志

	// todo 后续完成 分布式链路追踪
	traceContext := bl.NewTrace()
	// 上游可能还有网关等服务
	if traceId := c.Request.Header.Get("com-header-rid"); traceId != "" {
		traceContext.TraceId = traceId
	}
	c.Set("trace", traceContext)

	// 用于记录执行时间
	c.Set("startExecTime", time.Now())

	bodyBytes, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Write body back
	req := convert(map[string]interface{}{
		"uri":      c.Request.RequestURI,
		"method":   c.Request.Method,
		"args":     c.Request.PostForm,
		"body":     string(bodyBytes),
		"from":     c.ClientIP(),
		"trace_id": traceContext.TraceId,
	})

	bl.Infof("%s %s", "[http_request_in]", req)
}

// 请求输出日志
func RequestOutLog(c *gin.Context) {
	// 打印响应日志
	// uri,method,args,ip,response_body,process_time, ....
	endExecTime := time.Now()
	st, _ := c.Get("startExecTime")
	startExecTime, _ := st.(time.Time)
	response, _ := c.Get("response")

	ctx, _ := c.Get("trace")
	traceContext := ctx.(*bl.TraceContext)

	bl.Infof("%s %s", "[http_request_out]", convert(map[string]interface{}{
		"uri":       c.Request.RequestURI,
		"method":    c.Request.Method,
		"args":      c.Request.PostForm,
		"from":      c.ClientIP(),
		"response":  response,
		"proc_time": endExecTime.Sub(startExecTime).Seconds(),
		"trace_id":  traceContext.TraceId,
	}))

}

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 看配置是否开启记录请求日志
		if iron_config.Conf.Exist("/log_req") { // todo 这种最好在一开始的时候对外提供一个函数初始化好。通过变量来判断，不要过于外部依赖包
			RequestInLog(c)
			defer RequestOutLog(c)
		}
		c.Next()
	}
}

// // log 使用的函数
// 将 map 转化成 打印的string
func convert(m map[string]interface{}) (result string) {
	for _key, _val := range m {

		result = result + "||" + fmt.Sprintf("%v=%+v", _key, _val)
	}
	//result := strings.Trim(fmt.Sprintf("%q", result), "\"")
	return result
}
