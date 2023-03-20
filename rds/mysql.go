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

// Option 使用Option设计模式修改配置
type Option func(db *sql.DB)

func WithMaxOpenConn(maxOpenConn int) Option {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(maxOpenConn)
	}
}

func WithConnMaxLifetime(connMaxLife int) Option {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(time.Duration(connMaxLife) * time.Second)
	}
}

func WithMaxIdleConn(maxIdleConn int) Option {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(maxIdleConn)
	}
}

func WithConnMaxIdleTime(maxIdleTime int) Option {
	return func(db *sql.DB) {
		db.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)
	}
}

// NewSQLClient 构建一个新的SQLClient
func NewSQLClient(address string, opts ...Option) *MySQLClient {
	db, err := sql.Open("mysql", address)
	if err != nil {
		panic("couldn't get rds client" + err.Error())
	}
	for _, option := range opts {
		option(db)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return &MySQLClient{db}
}

// CloseMySQL 关闭数据库
func (client *MySQLClient) CloseMySQL() {
	if client != nil {
		utils.CloseAction(client)
	}
}
