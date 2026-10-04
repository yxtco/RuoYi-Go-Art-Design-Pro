package monitor

import (
	"go-fin-server/internal/handler/monitor"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupCacheRoutes(g *gin.RouterGroup) {
	cacheHandler := monitor.NewCacheHandler()
	CacheGroup := g.Group("/cache")
	{
		CacheGroup.GET("", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.GetCache)                                   // 查询缓存详细
		CacheGroup.GET("getNames", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.ListCacheName)                      // 查询缓存名称列表
		CacheGroup.GET("getKeys/:cacheName", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.GetCacheKeys)             // 查询缓存键名列表
		CacheGroup.GET("getValue/:cacheName/:cacheKey", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.GetCacheValue) // 查询缓存内容
		CacheGroup.DELETE("clearCacheName/:cacheName", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.ClearCacheName) // 清理指定名称缓存
		CacheGroup.DELETE("clearCacheKey/:cacheKey", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.ClearCacheKey)    // 清理指定键名缓存
		CacheGroup.DELETE("clearCacheAll", middleware.PermissionMiddleware("monitor:cache:list"), cacheHandler.ClearCacheAll)              // 清理全部缓存
	}
}
