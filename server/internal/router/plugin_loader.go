package router

import (
	"go-fin-server/pkg"
	"go-fin-server/plugins"

	"github.com/gin-gonic/gin"
)

// setupPluginRoutes 自动加载并注册所有插件路由
// 插件通过 init() 函数自动注册到 plugins.Registry，此处统一遍历调用
// 新增插件只需在 router.go 中空白导入即可，无需修改本文件
func setupPluginRoutes(r *gin.Engine) {
	registrars := plugins.GetAllRegistrars()
	if len(registrars) == 0 {
		return
	}
	pkg.Logger.Infof("检测到 %d 个插件，开始注册路由...", len(registrars))
	for _, registrar := range registrars {
		registrar(r)
	}
}
