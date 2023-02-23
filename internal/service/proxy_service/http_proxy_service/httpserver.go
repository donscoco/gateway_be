package http_proxy_service

import (
	"context"
	"github.com/donscoco/gateway_be/conf/certificate"
	"github.com/donscoco/gateway_be/internal/middleware/manager_middleware"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	HttpSrvHandler  *http.Server
	HttpsSrvHandler *http.Server
)

func HttpServerRun() {
	config := iron_config.Conf

	gin.SetMode(config.GetString("/proxy/http/debug_mode"))
	r := InitRouter(
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
	)
	HttpSrvHandler = &http.Server{
		Addr:           config.GetString("/proxy/http/addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(config.GetInt("/proxy/http/read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(config.GetInt("/proxy/http/write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(config.GetInt("/proxy/http/max_header_bytes")),
	}

	go func() {
		log.Printf(" [INFO] http_proxy_run %s\n", config.GetString("/proxy/http/addr"))
		if err := HttpSrvHandler.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf(" [ERROR] http_proxy_run %s err:%v\n", config.GetString("/proxy/http/addr"), err)
		}
	}()
}
func HttpServerStop() {
	config := iron_config.Conf

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Printf(" [ERROR] http_proxy_stop err:%v\n", err)
	}
	log.Printf(" [INFO] http_proxy_stop %v stopped\n", config.GetString("/proxy/http/addr"))
}

/* https */
func HttpsServerRun() {
	config := iron_config.Conf

	gin.SetMode(config.GetString("/proxy/https/debug_mode"))

	r := InitRouter(
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
	)
	HttpsSrvHandler = &http.Server{
		Addr:           config.GetString("/proxy/https/addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(config.GetInt("/proxy/https/read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(config.GetInt("/proxy/https/write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(config.GetInt("/proxy/https/max_header_bytes")),
	}

	go func() {
		log.Printf(" [INFO] https_proxy_run %s\n", config.GetString("/proxy/https/addr"))
		if err := HttpsSrvHandler.ListenAndServeTLS(certificate.Path("test.pem"), certificate.Path("test.key")); err != nil && err != http.ErrServerClosed {
			log.Fatalf(" [ERROR] https_proxy_run %s err:%v\n", config.GetString("/proxy/https/addr"), err)
		}
	}()
}

func HttpsServerStop() {
	config := iron_config.Conf

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpsSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] https_proxy_stop err:%v\n", err)
	}
	log.Printf(" [INFO] https_proxy_stop %v stopped\n", config.GetString("/proxy/https/addr"))
}
