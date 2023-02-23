package service

import (
	"errors"
	"github.com/donscoco/gateway_be/internal/module/service_manager"
	db "github.com/donscoco/gateway_be/pkg/base/mysql"
	"github.com/donscoco/gateway_be/pkg/base/redis"
	"github.com/donscoco/gateway_be/pkg/gorm"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"log"
)

// / 初始化 GORM
func InitGORMService() {

	gorm.DefaultDB = iron_config.Conf.GetString("/server/mysql/0/proxy_name")
	// gorm
	mysqlProxy, ok := db.MySQLManager[gorm.DefaultDB]
	if !ok {
		err := errors.New("init gorm fail, empty mysql")
		log.Fatalln(err)
	}
	err := gorm.InitDBPool(
		gorm.DefaultDB,
		mysqlProxy.Session,
	)
	if err != nil {
		log.Fatalln(err)
	}
	return
}
func StopGORMService() {
	err := gorm.CloseDB()
	if err != nil {
		log.Fatalln(err)
	}
}

// 初始化 mysql
func InitMysqlService() {

	proxys := make([]db.MySQLProxy, 0, 16)
	iron_config.Conf.GetByScan("/server/mysql", &proxys)

	err := db.InitMySQLProxys(proxys)
	if err != nil {
		log.Fatalln(err)
	}
}
func StopMysqlService() {
	for _, p := range db.MySQLManager {
		err := p.Session.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}
	return
}

func InitRedisService() {

	proxys := make([]redis.RedisConfig, 0, 16)
	iron_config.Conf.GetByScan("/server/redis", &proxys)

	err := redis.InitSingleRedisClient(proxys)
	if err != nil {
		log.Fatalln(err)
	}
}
func StopRedisService() {
	for _, p := range redis.RedisSingleClients {
		err := p.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}

	for _, p := range redis.RedisClusterClients {
		err := p.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}

	return
}

func InitAppListService() {
	err := service_manager.AppManagerHandler.LoadOnce()
	if err != nil {
		log.Fatalln(err)
	}
}
func InitProxyListService() {
	err := service_manager.ServiceManagerHandler.LoadOnce()
	if err != nil {
		log.Fatalln(err)
	}
}
