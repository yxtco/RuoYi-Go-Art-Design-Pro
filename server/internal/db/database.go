package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBConnections map[string]*gorm.DB

// DatabaseConfig 数据库配置结构体
type DatabaseConfig struct {
	Name     string `mapstructure:"name"`
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	LogLevel int    `mapstructure:"logLevel"`
}

// BuildDSN 根据独立字段构建 DSN 连接字符串
func (c *DatabaseConfig) BuildDSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			c.Host, c.Username, c.Password, c.Database, c.Port)
	default:
		return ""
	}
}

// SetupDatabase 根据配置初始化并返回数据库连接
func SetupDatabase(configs []DatabaseConfig) error {
	DBConnections = make(map[string]*gorm.DB)
	for _, config := range configs {
		var db *gorm.DB
		var err error

		dsn := config.BuildDSN()
		switch config.Driver {
		case "mysql":
			db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		case "postgres":
			db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		default:
			return fmt.Errorf("unsupported database driver: %s", config.Driver)
		}

		if err != nil {
			return fmt.Errorf("error opening database: %w", err)
		}

		// 配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("error getting database instance: %w", err)
		}
		// 当你没有开启debug模式的时候，gorm底层默认的log级别是Warn，当你的SQL语句执行时间超过了100ms的时候就会触发Warn日志打印，同时错误的SQL语句也会触发。
		// 如果你想要在没有开启debug的时候什么语句也不要打印
		db.Logger = db.Logger.LogMode(logger.LogLevel(config.LogLevel))
		// 根据需求配置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		// 将初始化的数据库连接添加到全局变量中，以 name 作为键
		DBConnections[config.Name] = db

		// 注册 SQL 监控插件
		if err := db.Use(&SQLMonitorPlugin{}); err != nil {
			return fmt.Errorf("error registering SQL monitor plugin: %w", err)
		}
	}
	return nil
}

// GetDBConnection 根据 name 获取数据库连接
func GetDBConnection(name string) (*gorm.DB, error) {
	db, ok := DBConnections[name]
	if !ok {
		return nil, fmt.Errorf("database connection with name '%s' not found", name)
	}
	return db, nil
}

// CloseAllDBConnections 关闭所有数据库连接
func CloseAllDBConnections() {
	for _, db := range DBConnections {
		sqlDB, err := db.DB()
		if err != nil {
			// 处理错误
			continue
		}
		sqlDB.Close()
	}
}
