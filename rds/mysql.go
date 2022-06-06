package rds

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/yeahyf/go_base/utils"
)

type MySQLClient struct {
	*sql.DB
}

//NewMySQLClient 构建一个新的SQLClient
func NewMySQLClient(maxsize, maxLife, maxIdleCon int, address string) *MySQLClient {
	db, err := sql.Open("mysql", address)
	if err != nil {
		panic("couldn't get rds client" + err.Error())
	}

	db.SetMaxOpenConns(maxsize)
	db.SetConnMaxLifetime(time.Duration(maxLife) * time.Second)
	db.SetMaxIdleConns(maxIdleCon)
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	client := &MySQLClient{
		db,
	}

	return client
}

//CloseMySQL 关闭数据库
func (client *MySQLClient) CloseMySQL() {
	if client != nil {
		utils.CloseAction(client)
	}
}
