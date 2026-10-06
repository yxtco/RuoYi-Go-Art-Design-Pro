package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 系统级 SugaredLogger（system.log + 控制台，error 级别同步写入 error.log）
var Logger *zap.SugaredLogger

// OperationLogger 操作级 SugaredLogger（写入 operation.log）
var OperationLogger *zap.SugaredLogger

// systemAtomicLevel 系统日志的原子级别，支持运行时动态调整
var systemAtomicLevel zap.AtomicLevel

// mu 保护 Logger / OperationLogger 重建时的并发安全
var mu sync.Mutex

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Dir        string `mapstructure:"dir"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// SetupLogger 初始化多级日志系统
// - system.log：系统级日志（debug/info/warn）+ 控制台输出
// - error.log：错误级日志（error/panic/fatal）
// - operation.log：操作级日志
func SetupLogger(logConfig LogConfig) {
	mu.Lock()
	defer mu.Unlock()

	// 解析日志级别
	level := parseLevel(logConfig.Level)

	// 确保日志目录存在
	logDir := logConfig.Dir
	if logDir == "" {
		logDir = "./logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Println("Error creating log directory:", err)
	}

	// 默认 lumberjack 配置
	maxSize := logConfig.MaxSize
	if maxSize <= 0 {
		maxSize = 100
	}
	maxBackups := logConfig.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 30
	}
	maxAge := logConfig.MaxAge
	if maxAge <= 0 {
		maxAge = 7
	}

	// 选择编码格式
	encoding := "json"
	if logConfig.Format != "json" {
		encoding = "console"
	}

	// 文件编码器配置（无彩色）
	fileEncoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台编码器配置（带彩色）
	consoleEncoderConfig := fileEncoderConfig
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var fileEncoder, consoleEncoder zapcore.Encoder
	if encoding == "json" {
		fileEncoder = zapcore.NewJSONEncoder(fileEncoderConfig)
		consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
	} else {
		fileEncoder = zapcore.NewConsoleEncoder(fileEncoderConfig)
		consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
	}

	// 创建原子级别（支持运行时动态调整）
	systemAtomicLevel = zap.NewAtomicLevelAt(level)

	// --- 系统级日志 Core（system.log + 控制台）---
	systemWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "system.log"),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   logConfig.Compress,
	}
	// 控制台 Core（带彩色，使用原子级别）
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), systemAtomicLevel)
	// 文件 Core（无彩色，使用原子级别）
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(systemWriter), systemAtomicLevel)
	// 合并控制台和文件 Core
	systemCore := zapcore.NewTee(consoleCore, fileCore)

	// --- 错误级日志 Core（error.log）---
	errorWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "error.log"),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   logConfig.Compress,
	}
	errorCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(errorWriter), zapcore.ErrorLevel)

	// 合并系统日志和错误日志 Core
	combinedCore := zapcore.NewTee(systemCore, errorCore)

	// 创建系统 Logger（带调用者信息）
	zapLogger := zap.New(combinedCore, zap.AddCaller(), zap.AddCallerSkip(0))
	Logger = zapLogger.Sugar()

	// --- 操作级日志（operation.log）---
	operationWriter := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "operation.log"),
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   logConfig.Compress,
	}
	operationCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(operationWriter), zapcore.InfoLevel)
	operationZapLogger := zap.New(operationCore)
	OperationLogger = operationZapLogger.Sugar()
}

// SetLogLevel 运行时动态调整系统日志级别
// levelStr 支持：quiet（安静，仅 error）、standard（标准，info）、detailed（详细，debug），或直接传 zap 级别名
func SetLogLevel(levelStr string) {
	mu.Lock()
	defer mu.Unlock()

	level := parseLevel(levelStr)
	systemAtomicLevel.SetLevel(level)
}

// SyncLogger 刷新日志缓冲区（程序退出前调用）
func SyncLogger() {
	if Logger != nil {
		_ = Logger.Sync()
	}
	if OperationLogger != nil {
		_ = OperationLogger.Sync()
	}
}

// parseLevel 解析日志级别字符串为 zapcore.Level
// 支持前端友好名称：quiet（安静）→ error、standard（标准）→ info、detailed（详细）→ debug
func parseLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "quiet":
		return zapcore.ErrorLevel
	case "standard", "":
		return zapcore.InfoLevel
	case "detailed":
		return zapcore.DebugLevel
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
