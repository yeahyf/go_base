///mysql数据库接口封装
package rds

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLClient struct {
	*sql.DB
}

///构建一个新的SQLClient
func NewMySQLClient(maxsize, maxlife, maxidlecon int, address string) *MySQLClient {
	db, err := sql.Open("mysql", address)
	if err != nil {
		panic("Get MySQL Client Error! info:" + err.Error())
	}

	db.SetMaxOpenConns(maxsize)
	db.SetConnMaxLifetime(time.Duration(maxlife) * time.Second)
	db.SetMaxIdleConns(maxidlecon)
	db.Ping()

	client := &MySQLClient{
		db,
	}

	return client
}

//关闭数据库
func (client *MySQLClient) CloseMySQL() {
	if client != nil {
		client.Close()
	}
}
