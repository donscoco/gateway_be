package main

import (
	"flag"
	conf "github.com/donscoco/gateway_be/conf"
	"github.com/donscoco/gateway_be/internal/service"
	"github.com/donscoco/gateway_be/internal/service/manager_service"
	"github.com/donscoco/gateway_be/internal/service/proxy_service/grpc_proxy_service"
	"github.com/donscoco/gateway_be/internal/service/proxy_service/http_proxy_service"
	"github.com/donscoco/gateway_be/internal/service/proxy_service/tcp_proxy_service"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/donscoco/gateway_be/pkg/iron_core"
	"github.com/donscoco/gateway_be/pkg/iron_log"
	"os"
	"os/signal"
	"syscall"
)

//var configFile string

const GATEWAY_ENV = "GATEWAY_ENV"

var env string

var GoCore *iron_core.Core
var Config *iron_config.Config
var Logger iron_core.Logger = iron_log.NewLogger("CORE")

func init() {
	flag.StringVar(&env, "env", "dev", "")
	//flag.StringVar(&configFile, "config", "./conf/dev/config.json", "input config ")
}

/*
在根目录下 swag init -g cmd/main.go -o internal/swagger
指定 cmd/main.go 生成swagger文件到 internal/swagger 目录，（在router 设置base目录  internal/swagger ）
*/
func main() {

	flag.Parse()

	gatewayenv := os.Getenv(GATEWAY_ENV)
	if len(gatewayenv) > 0 {
		env = gatewayenv
	}

	// 初始化config
	iron_config.Conf = iron_config.NewConfiguration(conf.Path("./" + env + "/config.json"))

	// 初始化log
	//iron_log.InitLoggerByEnv()
	iron_log.InitLoggerByParam(
		iron_config.Conf.GetString("/log/log_path"),
		iron_config.Conf.GetString("/log/log_level"),
		iron_config.Conf.GetString("/log/log_mode"),
	)

	service.InitRedisService()
	service.InitMysqlService()
	service.InitGORMService()
	manager_service.HttpServerRun() // 后台 增删改查

	// proxy 服务管理
	service.InitAppListService()
	service.InitProxyListService()

	http_proxy_service.HttpServerRun()
	http_proxy_service.HttpsServerRun()
	tcp_proxy_service.TcpServerRun()
	grpc_proxy_service.GrpcServerRun()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	http_proxy_service.HttpServerStop()
	http_proxy_service.HttpsServerStop()
	tcp_proxy_service.TcpServerStop()
	grpc_proxy_service.GrpcServerStop()

	manager_service.HttpServerStop()
	service.StopGORMService()
	service.StopMysqlService()
	service.StopRedisService()

	iron_log.Close()

}
