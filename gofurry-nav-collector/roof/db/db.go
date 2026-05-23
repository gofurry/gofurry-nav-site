package db

import (
	"net"
	"net/url"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
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
	dbName := env.GetServerConfig().DataBase.DBName
	dbUser := env.GetServerConfig().DataBase.DBUsername
	dbPassword := env.GetServerConfig().DataBase.DBPassword
	dbHost := env.GetServerConfig().DataBase.DBHost
	dbPort := env.GetServerConfig().DataBase.DBPort
	dsn := buildPostgresDSN(dbHost, dbPort, dbUser, dbPassword, dbName)

	db.engine, err = gorm.Open(postgres.Open(dsn))

	if err != nil {
		log.Fatal("open database error: " + err.Error())
	}
	sqlDB, err := db.engine.DB()
	if err != nil {
		log.Fatal("get database handle error: " + err.Error())
	}
	sqlDB.SetMaxIdleConns(intOrDefault(env.GetServerConfig().DataBase.MaxIdleConns, 100))
	sqlDB.SetMaxOpenConns(intOrDefault(env.GetServerConfig().DataBase.MaxOpenConns, 1000))
	sqlDB.SetConnMaxLifetime(secondsOrDefault(env.GetServerConfig().DataBase.ConnMaxLifetimeSeconds, 60))
	if env.GetServerConfig().DataBase.ConnMaxIdleTimeSeconds > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(env.GetServerConfig().DataBase.ConnMaxIdleTimeSeconds) * time.Second)
	}
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func buildPostgresDSN(host string, port string, user string, password string, dbName string) string {
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbName,
	}
	return dsn.String()
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
}

func intOrDefault(value int, def int) int {
	if value > 0 {
		return value
	}
	return def
}

func secondsOrDefault(value int, def int) time.Duration {
	return time.Duration(intOrDefault(value, def)) * time.Second
}
