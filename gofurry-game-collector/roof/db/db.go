package db

import (
	"fmt"
	"github.com/gofurry/gofurry-game-collector/roof/env"
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
	var url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", dbHost, dbPort, dbUser, dbPassword, dbName)

	db.engine, err = gorm.Open(postgres.Open(url))

	if err != nil {
		log.Fatal("open database error: " + err.Error())
	}
	sqlDB, _ := db.engine.DB()
	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(1000)
	sqlDB.SetConnMaxLifetime(60 * time.Second)
	err = sqlDB.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func (db *orm) DB() *gorm.DB {
	once.Do(initOrm)
	return db.engine
}
