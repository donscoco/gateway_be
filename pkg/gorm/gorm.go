package gorm

import (
	"database/sql"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBMapPool map[string]*sql.DB
var GORMMapPool map[string]*gorm.DB
var DefaultDB string

func InitDBPool(dbname string,sqldb *sql.DB) error {

	DBMapPool = make(map[string]*sql.DB)
	GORMMapPool = make(map[string]*gorm.DB)

	//gorm连接方式
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	if err != nil {
		return err
	}

	DBMapPool[dbname] = sqldb
	GORMMapPool[dbname] = gormDB
	return nil
}

func GetGormPool(name string) (*gorm.DB, error) {
	if dbpool, ok := GORMMapPool[name]; ok {
		return dbpool, nil
	}
	return nil, errors.New("get pool error")
}

func CloseDB() error {
	for _, dbpool := range DBMapPool {
		dbpool.Close()
	}
	DBMapPool = make(map[string]*sql.DB)
	GORMMapPool = make(map[string]*gorm.DB)
	return nil
}