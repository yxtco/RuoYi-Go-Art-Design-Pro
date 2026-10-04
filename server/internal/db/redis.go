package db

import (
	"github.com/redis/go-redis/v9"
)

var RedisConnections map[string]*redis.Client

type RedisConfig struct {
	Name     string `mapstructure:"name"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func SetupRedis(configs []RedisConfig) {
	RedisConnections = make(map[string]*redis.Client)
	for _, config := range configs {
		client := redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
		})
		// 设置连接池配置
		// 更多的连接池配置可以参考：https://pkg.go.dev/github.com/redis/go-redis/v9#Options
		client.Options().PoolSize = 10
		client.Options().MinIdleConns = 5
		RedisConnections[config.Name] = client
	}
}

// CloseAllRedisConnections 关闭所有数据库连接
func CloseAllRedisConnections() {
	for _, db := range RedisConnections {
		db.Close()
	}
}
