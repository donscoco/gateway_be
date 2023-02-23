package main

import (
	"github.com/donscoco/gateway_be/conf"
	"github.com/donscoco/gateway_be/internal/service"
	"github.com/donscoco/gateway_be/pkg/base/redis"
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"log"
	"time"
)

func main() {

	// 初始化config
	iron_config.Conf = iron_config.NewConfiguration(conf.Path("dev/config.json"))

	service.InitRedisService()

	p, ok := redis.RedisSingleClients["default"]
	if !ok {
		//todo log
		log.Printf("err")
		return
	}
	//cmd := p.Ping()
	//fmt.Println(cmd.Val())
	pipe := p.Pipeline()

	incrcmd := pipe.Incr("test_pipe")
	expirecmd := pipe.Expire("test_pipe", time.Second*100)

	cmds, err := pipe.Exec()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%+v", cmds)
	log.Printf("%+v", incrcmd)
	log.Printf("%+v", expirecmd)

}
