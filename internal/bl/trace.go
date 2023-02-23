package bl

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net"
	"os"
	"time"
)

// 用于生成 traceid 等信息打印在日志中，方便排查，
// todo 后续添加 分布式链路追踪

func GetGinTraceContext(c *gin.Context) *TraceContext {
	if c == nil {
		return NewTrace()
	}
	traceContext, exists := c.Get("trace")
	if exists {
		if tc, ok := traceContext.(*TraceContext); ok {
			return tc
		}
	}
	return NewTrace()
}

func SetGinTraceContext(c *gin.Context, trace *TraceContext) error {
	if trace == nil || c == nil {
		return errors.New("context is nil")
	}
	c.Set("trace", trace)
	return nil
}

func SetTraceContext(ctx context.Context, trace *TraceContext) context.Context {
	if trace == nil {
		return ctx
	}
	return context.WithValue(ctx, "trace", trace)
}

func GetTraceContext(ctx context.Context) *TraceContext {
	// 判断 ctx 是否 gin.Context
	// 是就直接获取 trace
	// 不是就转为 gin.Context 后再返回
	if ginCtx, ok := ctx.(*gin.Context); ok {
		traceIntraceContext, exists := ginCtx.Get("trace")
		if !exists {
			return NewTrace()
		}
		traceContext, ok := traceIntraceContext.(*TraceContext)
		if ok {
			return traceContext
		}
		return NewTrace()
	}

	if contextInterface, ok := ctx.(context.Context); ok {
		traceContext, ok := contextInterface.Value("trace").(*TraceContext)
		if ok {
			return traceContext
		}
		return NewTrace()

	}
	return NewTrace()
}

type Trace struct {
	TraceId     string
	SpanId      string
	Caller      string
	SrcMethod   string
	HintCode    int64
	HintContent string
}

type TraceContext struct {
	Trace
	CSpanId string
}

func NewTrace() *TraceContext {
	trace := &TraceContext{}
	trace.TraceId = GetTraceId()
	trace.SpanId = NewSpanId()
	return trace
}
func NewSpanId() string {
	timestamp := uint32(time.Now().Unix())
	ipToLong := binary.BigEndian.Uint32(GetLocalIPs()[0].To4())
	b := bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%08x", ipToLong^timestamp))
	b.WriteString(fmt.Sprintf("%08x", rand.Int31()))
	return b.String()
}
func GetTraceId() (traceId string) {
	return calcTraceId(GetLocalIPs()[0].String())
}

// 根据ip 生成 traceid
func calcTraceId(ip string) (traceId string) {
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()

	b := bytes.Buffer{}
	netIP := net.ParseIP(ip)
	if netIP == nil {
		b.WriteString("00000000")
	} else {
		b.WriteString(hex.EncodeToString(netIP.To4()))
	}
	b.WriteString(fmt.Sprintf("%08x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", rand.Int31n(1<<24)))
	b.WriteString("b0") // 末两位标记来源,b0为go

	return b.String()
}

// 获取localIps
func GetLocalIPs() (ips []net.IP) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	return ips
}
