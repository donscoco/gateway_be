package redis

import (
	"errors"
	"github.com/go-redis/redis"
	"log"
	"time"
)

var RedisSingleClients map[string]*redis.Client

type RedisConfig struct {
	ProxyName string `json:"proxy_name"`
	Username  string
	Password  string
	Addrs     []string
	Database  int

	IsCluster bool `json:"is_cluster"`

	// todo 根据 redis.ClusterOptions 的配置项添加
	DialTimeout        int `json:"dial_timeout"`
	ReadTimeout        int `json:"read_timeout"`
	WriteTimeout       int `json:"write_timeout"`
	MaxRetries         int `json:"max_retries"`
	PoolSize           int `json:"pool_size"`
	IdleTimeout        int `json:"idle_timeout"`
	IdleCheckFrequency int `json:"idle_check_frequency"`
}

func InitSingleRedisClient(rcs []RedisConfig) (err error) {

	RedisSingleClients = make(map[string]*redis.Client, len(rcs))
	for i := 0; i < len(rcs); i++ {
		// 获取配置路径
		rc := rcs[i]

		proxy, err := CreateSingleReidsClient(&rc)
		if err != nil {
			log.Fatal(err)
		}

		RedisSingleClients[rc.ProxyName] = proxy

	}
	return
}

func CreateSingleReidsClient(conf *RedisConfig) (cli *redis.Client, err error) {

	if len(conf.Addrs) != 1 {
		return nil, errors.New("集群节点数错误")
	}

	opt := &redis.Options{
		Addr:         conf.Addrs[0],
		DB:           conf.Database,
		DialTimeout:  time.Second * time.Duration(conf.DialTimeout),
		ReadTimeout:  time.Second * time.Duration(conf.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(conf.WriteTimeout),
		MaxRetries:   conf.MaxRetries,
		// todo 还有其他选项后续添加，暂时用默认的就行
	}
	cli = redis.NewClient(opt)

	cmd := cli.Ping()
	if cmd.Val() != "PONG" {
		return nil, cmd.Err()
	}

	return cli, nil
}
