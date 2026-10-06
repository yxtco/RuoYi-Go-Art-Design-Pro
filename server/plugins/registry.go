package plugins

import "github.com/gin-gonic/gin"

// RouteRegistrar 插件路由注册函数签名
// 插件在 init() 中注册此函数，启动时由 PluginLoader 统一调用
// 参数为 gin.Engine，插件可自行创建 Group
type RouteRegistrar func(r *gin.Engine)

// registry 全局插件注册表（init 阶段自动填充）
var registry []RouteRegistrar

// Register 注册插件路由（由插件的 init() 函数调用）
//
// 示例（在插件的 fin_router.go 中）：
//
//	func init() {
//	    plugins.Register(SetupFinRoutes)
//	}
func Register(registrar RouteRegistrar) {
	registry = append(registry, registrar)
}

// GetAllRegistrars 获取所有已注册的插件路由注册函数
func GetAllRegistrars() []RouteRegistrar {
	return registry
}
