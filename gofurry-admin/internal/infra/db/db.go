package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	env "github.com/gofurry/awesome-fiber-template/v3/medium/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Name string

const (
	Admin Name = "admin"
	Nav   Name = "nav"
	Game  Name = "game"
)

type connection struct {
	engine  *gorm.DB
	driver  string
	initErr error
	enabled bool
	config  env.DataBaseConfig
}

type manager struct {
	once sync.Once
	dbs  map[Name]*connection
}

var Databases = &manager{}

func InitDatabasesOnStart(models ...any) error {
	if err := Databases.init(); err != nil {
		return err
	}

	admin := Databases.dbs[Admin]
	if admin != nil && admin.enabled && admin.config.AutoMigrate {
		if err := Databases.AutoMigrate(Admin, models...); err != nil {
			return fmt.Errorf("auto migrate admin database failed: %w", err)
		}
	}

	slog.Info("database services initialized")
	return nil
}

func (m *manager) init() error {
	m.once.Do(func() {
		cfg := env.GetServerConfig()
		m.dbs = map[Name]*connection{
			Admin: {enabled: cfg.DataBase.Enabled, config: cfg.DataBase},
			Nav:   {enabled: cfg.BusinessDatabases.Nav.Enabled, config: cfg.BusinessDatabases.Nav},
			Game:  {enabled: cfg.BusinessDatabases.Game.Enabled, config: cfg.BusinessDatabases.Game},
		}

		for name, conn := range m.dbs {
			if !conn.enabled {
				continue
			}
			conn.initErr = conn.load(name)
		}
	})

	var err error
	for name, conn := range m.dbs {
		if conn != nil && conn.initErr != nil {
			err = errors.Join(err, fmt.Errorf("%s database init failed: %w", name, conn.initErr))
		}
	}
	return err
}

func (m *manager) DB(name Name) *gorm.DB {
	_ = m.init()
	if conn, ok := m.dbs[name]; ok {
		return conn.engine
	}
	return nil
}

func (m *manager) Enabled(name Name) bool {
	_ = m.init()
	if conn, ok := m.dbs[name]; ok {
		return conn.enabled
	}
	return false
}

func (m *manager) Err(name Name) error {
	_ = m.init()
	if conn, ok := m.dbs[name]; ok {
		return conn.initErr
	}
	return fmt.Errorf("database %s not configured", name)
}

func (m *manager) Ready(name Name) bool {
	engine := m.DB(name)
	if engine == nil {
		return !m.Enabled(name)
	}
	sqlDB, err := engine.DB()
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx) == nil
}

func (m *manager) ReadyAll() bool {
	_ = m.init()
	for name, conn := range m.dbs {
		if conn == nil || !conn.enabled {
			continue
		}
		if !m.Ready(name) {
			return false
		}
	}
	return true
}

func (m *manager) AutoMigrate(name Name, models ...any) error {
	engine := m.DB(name)
	if engine == nil {
		return fmt.Errorf("database %s is not initialized", name)
	}
	return engine.AutoMigrate(models...)
}

func (m *manager) Close() {
	for _, conn := range m.dbs {
		if conn == nil || conn.engine == nil {
			continue
		}
		sqlDB, err := conn.engine.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
	m.dbs = nil
	m.once = sync.Once{}
}

func (c *connection) load(name Name) error {
	dialector, driver, err := buildDialector(c.config)
	if err != nil {
		return fmt.Errorf("build dialector failed: %w", err)
	}

	engine, err := gorm.Open(dialector)
	if err != nil {
		return fmt.Errorf("open database failed: %w", err)
	}

	sqlDB, err := engine.DB()
	if err != nil {
		return fmt.Errorf("get sql db failed: %w", err)
	}

	configurePool(sqlDB, driver)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return fmt.Errorf("ping database failed: %w", err)
	}

	c.engine = engine
	c.driver = driver
	slog.Info("database connected", "name", string(name), "driver", driver)
	return nil
}

func buildDialector(cfg env.DataBaseConfig) (gorm.Dialector, string, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.DBType))
	switch driver {
	case "", "postgres", "postgresql":
		return postgres.Open(buildPostgresDSN(cfg.Postgres)), "postgres", nil
	case "mysql":
		return mysql.Open(buildMySQLDSN(cfg.MySQL)), "mysql", nil
	case "sqlite":
		dsn, err := buildSQLiteDSN(cfg.SQLite)
		if err != nil {
			return nil, "", err
		}
		return sqlite.Open(dsn), "sqlite", nil
	default:
		return nil, "", fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}
}

func buildPostgresDSN(cfg env.SQLDataBaseConfig) string {
	if strings.TrimSpace(cfg.DSN) != "" {
		return cfg.DSN
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName,
	)
}

func buildMySQLDSN(cfg env.SQLDataBaseConfig) string {
	if strings.TrimSpace(cfg.DSN) != "" {
		return cfg.DSN
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}

func buildSQLiteDSN(cfg env.SQLiteDataBaseConfig) (string, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		dsn = strings.TrimSpace(cfg.Path)
	}
	if dsn == "" {
		dsn = "./data/app.db"
	}

	if dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return dsn, nil
	}

	dir := filepath.Dir(dsn)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create sqlite directory failed: %w", err)
		}
	}

	return dsn, nil
}

func configurePool(sqlDB *sql.DB, driver string) {
	switch driver {
	case "sqlite":
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetConnMaxLifetime(0)
		sqlDB.SetConnMaxIdleTime(0)
	default:
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}
}
