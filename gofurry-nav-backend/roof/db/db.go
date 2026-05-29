package db

import (
	"log/slog"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-backend/roof/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
 * @Desc: 数据库
 * @author: 福狼
 * @version: v1.0.1
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

	pgsql := env.GetServerConfig().DataBase
	db.engine, err = gorm.Open(postgres.Open(buildPostgresDSN(pgsql)))
	if err != nil {
		slog.Error("open database error: " + err.Error())
		os.Exit(1)
	}

	sqlDB, err := db.engine.DB()
	if err != nil {
		slog.Error("get database pool error: " + err.Error())
		os.Exit(1)
	}
	sqlDB.SetMaxIdleConns(intOrDefault(pgsql.MaxIdleConns, 100))                 // 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxOpenConns(intOrDefault(pgsql.MaxOpenConns, 1000))                // 设置打开数据库连接的最大数量
	sqlDB.SetConnMaxLifetime(secondsOrDefault(pgsql.ConnMaxLifetimeSeconds, 60)) // 设置了可以重新使用连接的最大时间
	sqlDB.SetConnMaxIdleTime(secondsOrDefault(pgsql.ConnMaxIdleTimeSeconds, 30)) // 连接最大空闲时间

	err = sqlDB.Ping()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
}

// Close 关闭数据库连接池
func (db *orm) Close() {

	if db.engine == nil {
		return
	}

	sqlDB, err := db.engine.DB()
	if err != nil {
		slog.Error("failed to get SQL DB instance", "error", err)
		return
	}

	if err = sqlDB.Close(); err != nil {
		slog.Error("failed to close database connection pool", "error", err)
		return
	}

	db.engine = nil
	slog.Info("数据库连接池已关闭")
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

func buildPostgresDSN(pgsql env.DataBaseConfig) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(pgsql.DBUsername, pgsql.DBPassword),
		Host:   net.JoinHostPort(pgsql.DBHost, pgsql.DBPort),
		Path:   pgsql.DBName,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}
