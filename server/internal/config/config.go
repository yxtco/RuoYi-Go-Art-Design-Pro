package config

import (
	"bytes"
	"fmt"
	"go-fin-server/configs"
	"go-fin-server/internal/db"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

var GlobalConfig Config

// Config 结构体用于存储应用程序配置
type Config struct {
	Log struct {
		Level      string `mapstructure:"level"`       // 日志级别
		Format     string `mapstructure:"format"`      // 日志格式
		Dir        string `mapstructure:"dir"`         // 日志文件存储目录
		MaxSize    int    `mapstructure:"max_size"`    // 单个日志文件最大大小(MB)
		MaxBackups int    `mapstructure:"max_backups"` // 保留的旧日志文件数量
		MaxAge     int    `mapstructure:"max_age"`     // 保留旧日志文件的最大天数
		Compress   bool   `mapstructure:"compress"`    // 是否压缩旧日志文件
	} `mapstructure:"log"`
	RedisList    []db.RedisConfig    `mapstructure:"redis_list"`
	DatabaseList []db.DatabaseConfig `mapstructure:"database_list"`
	Server       struct {
		Port       string `mapstructure:"port"`
		Host       string `mapstructure:"host"`
		WebDistDir string `mapstructure:"web_dist"` // 前端编译产物(dist)目录，后端直接托管静态文件
	} `mapstructure:"server"`
	Rsa struct {
		PublicKey  string `mapstructure:"public_key"`
		PrivateKey string `mapstructure:"private_key"`
	} `mapstructure:"rsa"`
	Jwt struct {
		Secret         string `mapstructure:"secret"`
		ExpirationTime int64  `mapstructure:"expiration_time"`
	} `mapstructure:"jwt"`

	Upload struct {
		BasePath string `mapstructure:"base_path"` // 文件上传基础路径
		MaxSize  int64  `mapstructure:"max_size"`  // 最大文件大小(MB)
	} `mapstructure:"upload"`
	IsAddressEnabled bool `mapstructure:"is_address_enabled"`
	StartTime        time.Time
}

func (c Config) unimplementable() {
	//TODO implement me
	panic("implement me")
}

// SetupConfig 创建并返回应用程序配置
// 流程：
//  1. 检查磁盘运行时配置文件（./configs/<name>.yaml，或 RUOYI_CONFIG_PATH 指定路径）是否存在
//  2. 不存在则执行"初始化脚本"：从内嵌默认配置生成到磁盘，再加载
//  3. 存在则直接加载磁盘配置（可在不改代码的情况下调整连接/端口等）
//  4. 磁盘配置读取失败时回退到内嵌配置，保证二进制自包含可运行
func SetupConfig(c string) error {
	// 初始化/检查运行时配置文件，返回实际生效的磁盘路径
	runtimePath, err := EnsureRuntimeConfig(c)
	if err != nil {
		return err
	}

	// 优先读取磁盘运行时配置
	viper.SetConfigType("yaml")
	viper.SetConfigFile(runtimePath)
	if err := viper.ReadInConfig(); err == nil {
		return viper.Unmarshal(&GlobalConfig)
	}

	// 磁盘配置不可用（只读文件系统等）时回退内嵌配置
	configFile, err := configs.ConfigFs.ReadFile(fmt.Sprintf("%s.yaml", c))
	if err != nil {
		return fmt.Errorf("failed to read embedded config file: %v", err)
	}
	if err := viper.ReadConfig(bytes.NewBuffer(configFile)); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	return viper.Unmarshal(&GlobalConfig)
}

// EnsureRuntimeConfig 检查磁盘运行时配置文件是否存在，缺失则用内嵌模板生成（初始化脚本）
// 返回生效的磁盘配置路径
func EnsureRuntimeConfig(c string) (string, error) {
	path := runtimeConfigPath(c)

	// 已存在则无需初始化
	if fileExists(path) {
		return path, nil
	}

	// 读取内嵌默认配置作为初始化模板
	template, err := configs.ConfigFs.ReadFile(fmt.Sprintf("%s.yaml", c))
	if err != nil {
		return "", fmt.Errorf("配置文件 %s.yaml 内嵌模板不存在: %v", c, err)
	}

	// 初始化脚本：创建目录并写出默认配置
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %v", err)
	}
	if err := os.WriteFile(path, template, 0o600); err != nil {
		return "", fmt.Errorf("生成默认配置文件失败: %v", err)
	}
	fmt.Printf("[config] 配置文件 %s 不存在，已从内嵌模板生成默认配置\n", path)
	return path, nil
}

// runtimeConfigPath 计算运行时配置文件路径
// 优先使用环境变量 RUOYI_CONFIG_PATH，否则默认 ./configs/<name>.yaml
func runtimeConfigPath(c string) string {
	if p := os.Getenv("RUOYI_CONFIG_PATH"); p != "" {
		return p
	}
	return filepath.Join("configs", c+".yaml")
}

// fileExists 判断文件是否存在
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
