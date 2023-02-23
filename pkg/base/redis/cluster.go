package redis

import (
	"github.com/go-redis/redis"
	"log"
	"time"
)

var RedisClusterClients map[string]*redis.ClusterClient

func InitClusterRedisClient(rcs []RedisConfig) (err error) {
	RedisClusterClients = make(map[string]*redis.ClusterClient, len(rcs))
	for i := 0; i < len(rcs); i++ {
		// 获取配置路径
		rc := rcs[i]

		proxy, err := CreateClusterRedisClient(&rc)
		if err != nil {
			log.Fatal(err)
		}

		RedisClusterClients[rc.ProxyName] = proxy
	}
	return
}

func CreateClusterRedisClient(conf *RedisConfig) (cli *redis.ClusterClient, err error) {
	opt := &redis.ClusterOptions{
		Addrs:        conf.Addrs,
		Password:     conf.Password,
		DialTimeout:  time.Second * time.Duration(conf.DialTimeout),
		ReadTimeout:  time.Second * time.Duration(conf.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(conf.WriteTimeout),
		MaxRetries:   conf.MaxRetries,
	}
	if len(conf.Password) > 0 {
		opt.Password = conf.Password
	}
	cli = redis.NewClusterClient(opt)
	cmd := cli.Ping()
	if cmd.Val() != "PONG" {
		return nil, cmd.Err()
	}
	return cli, nil
}
