package router

import (
	"go-fin-server/internal/config"
	"go-fin-server/internal/db"
	"go-fin-server/internal/handler"
	"go-fin-server/internal/middleware"
	"go-fin-server/internal/router/monitor"
	"go-fin-server/internal/router/system"
	"go-fin-server/internal/router/tool"
	"go-fin-server/pkg/fileuploadtool"
	ws "go-fin-server/pkg/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-fin-server/docs" // swagger 自动生成的文档
)

// SetupRouter 设置路由
func SetupRouter(hub *ws.Hub) *gin.Engine {
	r := gin.New()
	// 信任所有代理，确保能获取真实客户端 IP（通过 X-Forwarded-For / X-Real-IP）
	_ = r.SetTrustedProxies([]string{"0.0.0.0/0", "::/0"})
	// 使用错误日志中间件
	//r.Use(middleware.ErrorLogMiddleware())
	// 使用跨域中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(gin.Recovery())
	// 使用限流中间件，设置每秒最多请求次数为 10
	//r.Use(middleware.RateLimitMiddleware(time.Second, 10))
	// 使用配置的上传目录作为静态资源服务路径
	r.Static("/assets", fileuploadtool.GetBasePath())
	// Swagger 接口文档（无需认证）
	r.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// WebSocket 聊天路由（需要在 JWT 中间件之前注册，token 通过 URL 参数传递）
	system.SetupChatRoutes(r, hub)
	// 网站设置公开查询（登录页未登录时读取，需在 JWT 中间件之前注册）
	system.SetupSiteSettingPublicRoute(r)

	// 无需认证的路由
	loginHandler := handler.NewLoginHandler()
	r.POST("/login", loginHandler.Login)            // 登录方法
	r.GET("/captchaImage", loginHandler.GetCodeImg) // 获取验证码

	// 需要认证的路由（JWT 中间件）
	r.Use(middleware.JWTMiddleware(db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime))
	r.POST("/logout", loginHandler.LoginOut)        // 退出方法（需要 JWT 认证以获取用户信息）
	r.Use(middleware.OperationLogMiddleware(db.DBConnections["master"]))
	r.GET("/getRouters", loginHandler.GetRouters) // 获取路由
	r.POST("/register", loginHandler.Register)    // 注册方法
	r.GET("/getInfo", loginHandler.GetInfo)       // 获取用户详细信息
	setupSystemRoutes(r)
	setupMonitorRoutes(r)
	setupToolRoutes(r)
	SetupCommonRoutes(r) // 公共上传下载路由

	// 后端直接托管前端编译产物（SPA），置于所有 API 路由之后
	ServeFrontend(r, config.GlobalConfig.Server.WebDistDir)
	return r
}

func setupSystemRoutes(r *gin.Engine) {
	systemGroup := r.Group("/system")
	system.SetupUserRoutes(systemGroup)
	system.SetupRoleRoutes(systemGroup)
	system.SetupDeptRoutes(systemGroup)
	system.SetupPostRoutes(systemGroup)
	system.SetupDictDataRoutes(systemGroup)
	system.SetupDictTypeRoutes(systemGroup)
	system.SetupConfigRoutes(systemGroup)
	system.SetupSiteSettingRoutes(systemGroup)
	system.SetupMenuRoutes(systemGroup)
	system.SetupNoticeRoutes(systemGroup)
}

func setupMonitorRoutes(r *gin.Engine) {
	monitorGroup := r.Group("/monitor")
	monitor.SetupCacheRoutes(monitorGroup)
	monitor.SetupJobRoutes(monitorGroup)
	monitor.SetupJobLogRoutes(monitorGroup)
	monitor.SetupLoginInfoRoutes(monitorGroup)
	monitor.SetupOnlineRoutes(monitorGroup)
	monitor.SetupOperLogRoutes(monitorGroup)
	monitor.SetupServerRoutes(monitorGroup)
	monitor.SetupSQLMonitorRoutes(monitorGroup)
}

func setupToolRoutes(r *gin.Engine) {
	toolGroup := r.Group("/tool")
	tool.SetupGenRoutes(toolGroup)
}
