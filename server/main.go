// @title			RuoYi-Go API
// @version		1.0
// @description		若依 Go 版本接口文档
// @BasePath		/
// @securityDefinitions.apikey	BearerAuth
// @in				header
// @name				Authorization
// @description		Bearer {token}
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go-fin-server/internal/config"
	"go-fin-server/internal/db"
	"go-fin-server/internal/initdb"
	"go-fin-server/internal/jobscheduler"
	"go-fin-server/internal/migrate"
	"go-fin-server/internal/router"
	"go-fin-server/internal/service/system"
	"go-fin-server/pkg"
	"go-fin-server/pkg/fileuploadtool"
	ws "go-fin-server/pkg/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	c string
)

func init() {
	flag.StringVar(&c, "c", "release", "配置文件")
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	flag.Parse()
	// 设置启动时间
	config.GlobalConfig.StartTime = time.Now()
	// 初始化配置文件
	if err := config.SetupConfig(c); err != nil {
		panic(err)
	}
	// 初始化日志配置
	logCfg := config.GlobalConfig.Log
	pkg.SetupLogger(pkg.LogConfig{
		Level:      logCfg.Level,
		Format:     logCfg.Format,
		Dir:        logCfg.Dir,
		MaxSize:    logCfg.MaxSize,
		MaxBackups: logCfg.MaxBackups,
		MaxAge:     logCfg.MaxAge,
		Compress:   logCfg.Compress,
	})
	defer pkg.SyncLogger()
	// 初始化上传目录
	if err := fileuploadtool.EnsureDir(fileuploadtool.GetBasePath()); err != nil {
		pkg.Logger.Warn("创建上传目录失败: " + err.Error())
	} else {
		pkg.Logger.Info("上传目录初始化完成: " + fileuploadtool.GetBasePath())
	}
	// 初始化 Redis 连接
	db.SetupRedis(config.GlobalConfig.RedisList)
	defer db.CloseAllRedisConnections()
	// 初始化数据库（确保库存在；若为空库则自动导入 data/ry_init_full.sql 完成初始化）
	if err := initdb.EnsureInitialized(); err != nil {
		pkg.Logger.Error("数据库自动初始化失败: " + err.Error())
		panic(err)
	}
	// 初始化 数据库 连接
	if err := db.SetupDatabase(config.GlobalConfig.DatabaseList); err != nil {
		panic(err)
	}
	defer db.CloseAllDBConnections()
	// 执行 golang-migrate 数据库迁移（迁移文件内嵌于二进制，幂等）
	if err := migrate.Run(); err != nil {
		pkg.Logger.Error("数据库迁移失败: " + err.Error())
		panic(err)
	}
	pkg.Logger.Info("数据库迁移完成")
	// 启动时加载 MySQL 字典数据到 Redis 缓存
	dictService := system.SysDictTypeService{}
	if err := dictService.LoadingDictCache(); err != nil {
		pkg.Logger.Warn("加载字典缓存失败: " + err.Error())
	} else {
		pkg.Logger.Info("字典缓存加载完成")
	}
	// 从数据库加载日志级别（DB 为准，yaml 仅作为启动兜底）
	configService := system.SysConfigService{}
	if dbLogLevel, err := configService.SelectConfigByKey("sys.log.level"); err == nil && dbLogLevel != "" {
		pkg.SetOperationLogLevel(dbLogLevel)
		pkg.Logger.Infof("日志级别已从数据库加载: %s", dbLogLevel)
	}
	// 启动定时任务调度器
	taskScheduler := jobscheduler.NewTaskScheduler()
	taskScheduler.Start()
	defer taskScheduler.Stop()
	pkg.Logger.Info("定时任务调度器已启动")
	// 设置 GIN 模式，release 模式下不输出路由注册调试日志
	gin.SetMode(gin.ReleaseMode)
	// 初始化 WebSocket Hub（传入 Redis 客户端用于存储在线用户）
	chatHub := ws.NewHub(db.RedisConnections["master"])
	go chatHub.Run()
	pkg.Logger.Info("WebSocket Hub 已启动")
	// 初始化路由
	r := router.SetupRouter(chatHub)
	// 优雅重启或停止
	server := NewServer(fmt.Sprintf(":%s", config.GlobalConfig.Server.Port), r)
	server.Start()
	// 打印启动成功信息
	port := config.GlobalConfig.Server.Port
	ips := getLocalIPs()
	fmt.Println("========================================")
	fmt.Println("  服务启动成功!")
	for _, ip := range ips {
		fmt.Printf("  前端访问地址: https://%s:%s\n", ip, port)
	}
	fmt.Println("========================================")
	server.GracefulStop()
}

type Server struct {
	httpServer *http.Server
}

// NewServer 创建一个新的服务器实例
func NewServer(port string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    port,
			Handler: handler,
		},
	}
}

// Start 启动服务器
func (s *Server) Start() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

// GracefulStop 优雅地停止服务器
func (s *Server) GracefulStop() {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

// getLocalIPs 获取本机所有非回环 IPv4 地址
func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{"127.0.0.1"}
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ips = append(ips, ipNet.IP.String())
		}
	}
	if len(ips) == 0 {
		ips = append(ips, "127.0.0.1")
	}
	return ips
}
