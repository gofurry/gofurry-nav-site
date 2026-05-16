package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofurry/gofurry-game-collector/roof/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

/*
 * @Desc: MongoDB 通用连接服务
 * @author: 福狼
 * @version: v1.0.0
 */

// Mongo 全局MongoDB连接实例
var Mongo = &mongoDB{}
var mongoOnce sync.Once

// 初始化 MongoDB 连接
func initMongo() {
	Mongo.loadMongoConfig()
}

// mongoDB MongoDB 客户端和连接配置
type mongoDB struct {
	client *mongo.Client // 客户端
	dbName string        // 默认数据库名
}

// loadMongoConfig 加载配置并初始化MongoDB连接
func (m *mongoDB) loadMongoConfig() {
	// 防止重复初始化
	if m.client != nil {
		return
	}

	// 从环境配置中读取 MongoDB 参数
	mongoCfg := env.GetServerConfig().Mongodb
	if mongoCfg.Host == "" {
		log.Fatal("mongodb config error: host is empty")
	}

	// 构建 MongoDB 连接字符串
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/",
		mongoCfg.Username,
		mongoCfg.Password,
		mongoCfg.Host,
		mongoCfg.Port,
	)

	// 设置客户端连接选项
	clientOpts := options.Client().ApplyURI(uri).SetAuth(options.Credential{
		Username:      mongoCfg.Username,
		Password:      mongoCfg.Password,
		AuthSource:    mongoCfg.AuthDB, // 认证数据库
		AuthMechanism: "SCRAM-SHA-256", // 认证机制
	})

	// 配置连接池参数
	clientOpts.SetMaxPoolSize(100)                  // 最大连接数
	clientOpts.SetMinPoolSize(10)                   // 最小空闲连接数
	clientOpts.SetMaxConnIdleTime(60 * time.Second) // 连接最大空闲时间
	clientOpts.SetConnectTimeout(10 * time.Second)  // 连接超时时间
	clientOpts.SetTimeout(15 * time.Second)         // 操作超时时间

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 建立连接
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("mongodb connect error: %v", err)
	}

	// 验证连接
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("mongodb ping error: %v", err)
	}

	// 赋值到单例
	m.client = client
	m.dbName = mongoCfg.DBName // 默认数据库名
	log.Println("mongodb connect success!")
}

// Client 获取MongoDB客户端
func (m *mongoDB) Client() *mongo.Client {
	mongoOnce.Do(initMongo)
	return m.client
}

// DB 获取指定名称的数据库实例
// 如果dbName为空, 使用配置中的默认数据库名
func (m *mongoDB) DB(dbName ...string) *mongo.Database {
	mongoOnce.Do(initMongo)
	name := m.dbName
	if len(dbName) > 0 && dbName[0] != "" {
		name = dbName[0]
	}
	return m.client.Database(name)
}

// Collection 获取指定集合
// collName 集合名, dbName 数据库名
func (m *mongoDB) Collection(collName string, dbName ...string) *mongo.Collection {
	return m.DB(dbName...).Collection(collName)
}

// Close 关闭MongoDB连接
func (m *mongoDB) Close() error {
	if m.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}
