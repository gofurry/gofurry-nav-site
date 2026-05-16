package db

import (
	"fmt"
	"github.com/gofurry/gofurry-user/roof/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

/*
 * @Desc: 数据库
 * @author: 福狼
 * @version: v1.0.0
 */

var Orm = &orm{}
var once sync.Once

func initOrm() {
	Orm.loadDBConfig()
}

type orm struct {
	engine *gorm.DB
}

func (db *orm) loadDBConfig() {
	if db.engine != nil {
		return
	}
	var err error
	var dsn string

	pgsql := env.GetServerConfig().DataBase
	dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgsql.DBHost, pgsql.DBPort, pgsql.DBUsername, pgsql.DBPassword, pgsql.DBName)
	db.engine, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Fatal("open database error: " + err.Error())
	}

	sqlDB, _ := db.engine.DB()
	sqlDB.SetMaxIdleConns(100)                 // SetMaxIdleConns 设置空闲连接池中连接的最大数量。
	sqlDB.SetMaxOpenConns(1000)                // SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetConnMaxLifetime(60 * time.Second) // SetConnMaxLifetime 设置了可以重新使用连接的最大时间。
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
}
