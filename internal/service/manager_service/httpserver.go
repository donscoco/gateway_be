package manager_service

import (
	"context"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	HttpSrvHandler *http.Server
)

func HttpServerRun() {
	config := iron_config.Conf
	gin.SetMode(config.GetString("/http/debug_mode"))
	r := InitRouter()
	HttpSrvHandler = &http.Server{
		Addr:           config.GetString("/http/addr"),
		Handler:        r,
		ReadTimeout:    time.Duration(config.GetInt("/http/read_timeout")) * time.Second,
		WriteTimeout:   time.Duration(config.GetInt("/http/write_timeout")) * time.Second,
		MaxHeaderBytes: 1 << uint(config.GetInt("/http/max_header_bytes")),
	}
	go func() {
		log.Printf(" [INFO] HttpServerRun:%s\n", config.GetString("/http/addr"))
		if err := HttpSrvHandler.ListenAndServe(); err != nil {
			log.Fatalf(" [ERROR] HttpServerRun:%s err:%v\n", config.GetString("/http/addr"), err)
		}
	}()
}

func HttpServerStop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := HttpSrvHandler.Shutdown(ctx); err != nil {
		log.Fatalf(" [ERROR] HttpServerStop err:%v\n", err)
	}
	log.Printf(" [INFO] HttpServerStop stopped\n")
}
