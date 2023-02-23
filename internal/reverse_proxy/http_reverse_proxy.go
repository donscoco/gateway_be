package reverse_proxy

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"github.com/donscoco/gateway_be/conf/certificate"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/reverse_proxy/load_balance"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func NewLoadBalanceReverseProxy(c *gin.Context, lb load_balance.LoadBalance, trans *http.Transport) *httputil.ReverseProxy {
	//请求协调者
	director := func(req *http.Request) {
		nextAddr, err := lb.Get(req.URL.String())
		if err != nil || nextAddr == "" {
			panic("get next addr fail")

		}
		target, err := url.Parse(nextAddr)
		if err != nil {
			panic(err)
		}
		targetQuery := target.RawQuery
		// todo 后续改成可配置代理的目标服务是否是https，现在默认都是内网http的服务
		if target.Scheme == "https" {
			target.Scheme = "http"
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}

	//更改内容
	modifyFunc := func(resp *http.Response) error {

		//兼容websocket
		if strings.Contains(resp.Header.Get("Connection"), "Upgrade") {
			return nil
		}

		//兼容gzip
		var payload []byte
		var readErr error

		if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			payload, readErr = ioutil.ReadAll(gr)
			resp.Header.Del("Content-Encoding")
		} else {
			payload, readErr = ioutil.ReadAll(resp.Body)
		}
		if readErr != nil {
			return readErr
		}
		resp.Body = io.NopCloser(bytes.NewBuffer(payload))
		resp.ContentLength = int64(len(payload))
		resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
		return nil
	}

	//错误回调 ：关闭real_server时测试，错误回调
	//范围：transport.RoundTrip发生的错误、以及ModifyResponse发生的错误
	errFunc := func(w http.ResponseWriter, r *http.Request, err error) {
		bl.ResponseError(c, 999, err)
	}

	// 如果网关代理后面的服务不需要https，只是网关这里使用https，那就不用指定信任的ca证书签发机构
	// todo 后续改成可配置成反向代理的目标服务是否用https，目前默认代理的目标服务器是在内网部署的http

	// 反向代理的目标服务只是http时使用。如果是https需要指定信任的证书签发机构ca
	return &httputil.ReverseProxy{Director: director, ModifyResponse: modifyFunc, ErrorHandler: errFunc}

	// 测试用的ca证书：用于发送的节点是https时使用（购买浏览器内置ca签发的证书后，注释掉，用上面的就行）
	//return ReverseProxyForHttpsNode(director,modifyFunc,errFunc)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func ReverseProxyForHttpNode() {

}

func ReverseProxyForHttpsNode(director func(req *http.Request), modifyFunc func(resp *http.Response) error, errFunc func(w http.ResponseWriter, r *http.Request, err error)) *httputil.ReverseProxy {
	// 测试用的ca证书：（购买浏览器内置ca签发的证书后，注释掉，用上面的就行）
	var transportForHTTPS = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second, //连接超时
			KeepAlive: 30 * time.Second, //长连接超时时间
		}).DialContext,
		//TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},// 方法1：不需要证书的配置，跳过证书信任检查，也就是不管请求收到的服务器证书是谁签发的。都无条件信任。
		TLSClientConfig: func() *tls.Config {
			pool := x509.NewCertPool()
			caCertPath := certificate.Path("ca.crt") // 方法2：配置信任的ca机构。相当于浏览器中内置的ca机构证书。这样只要是这个指定的证书签发的服务器证书，这次请求就当作是可以信任的
			caCrt, _ := ioutil.ReadFile(caCertPath)
			pool.AppendCertsFromPEM(caCrt)
			return &tls.Config{RootCAs: pool}
		}(),
		MaxIdleConns:          100,              //最大空闲连接
		IdleConnTimeout:       90 * time.Second, //空闲超时时间
		TLSHandshakeTimeout:   10 * time.Second, //tls握手超时时间
		ExpectContinueTimeout: 1 * time.Second,  //100-continue 超时时间
	}
	return &httputil.ReverseProxy{Director: director, Transport: transportForHTTPS, ModifyResponse: modifyFunc, ErrorHandler: errFunc}

}
