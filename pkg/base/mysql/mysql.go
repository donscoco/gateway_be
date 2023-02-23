package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var MySQLManager map[string]*MySQLProxy

type MySQLProxy struct {
	ProxyName string `json:"proxy_name"`
	Username  string
	Password  string
	Addr      string
	Database  string

	Session *sql.DB

	// 连接时设置的参数，参考mysql driver 包的 parseDSNParams 进行添加
	ReadTimeoutSec  int `json:"read_timeout_sec"`
	WriteTimeoutSec int `json:"write_timeout_sec"`
	ConnTimeoutSec  int `json:"conn_timeout_sec"`

	// 建立连接后设置的参数
	ConnMaxLifetime int `json:"conn_max_lifetime"` // 设置链接的生命周期
	MaxIdleConns    int `json:"max_idle_conns"`    // 设置闲置链接数
	MaxOpenConns    int `json:"max_open_conns"`    // 设置最大链接数

}

func InitMySQLProxys(ps []MySQLProxy) (err error) {
	MySQLManager = make(map[string]*MySQLProxy, len(ps))
	for i := 0; i < len(ps); i++ {
		// 获取配置路径
		p := ps[i]

		proxy, err := CreateMySQLProxy(&p)
		if err != nil {
			log.Fatal(err)
		}

		MySQLManager[p.ProxyName] = proxy
	}
	return
}

func CreateMySQLProxy(p *MySQLProxy) (proxy *MySQLProxy, err error) {
	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?timeout=%ds&readTimeout=%ds&writeTimeout=%ds&parseTime=true",
		p.Username, p.Password, p.Addr, p.Database,
		p.ConnTimeoutSec,
		p.ReadTimeoutSec,
		p.WriteTimeoutSec,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		//todo
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// 设置链接的生命周期
	if p.ConnMaxLifetime != 0 {
		db.SetConnMaxLifetime(time.Second * time.Duration(int64(p.ConnMaxLifetime)))
	}
	// 设置闲置链接数
	if p.MaxIdleConns != 0 {
		db.SetMaxIdleConns(p.MaxIdleConns)
	}
	// 设置最大链接数
	if p.MaxOpenConns != 0 {
		db.SetMaxOpenConns(p.MaxOpenConns)
	}
	p.Session = db

	return p, err
}
