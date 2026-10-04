package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPostRoutes(g *gin.RouterGroup) {
	postHandler := system.NewPostHandler()
	postGroup := g.Group("/post")
	{
		postGroup.GET("list", middleware.PermissionMiddleware("system:post:list"), postHandler.ListPost)     // 查询岗位列表
		postGroup.POST("export", middleware.PermissionMiddleware("system:post:export"), postHandler.Export)  // 导出
		postGroup.GET(":id", middleware.PermissionMiddleware("system:post:query"), postHandler.GetPost)      // 查询岗位详细
		postGroup.POST("", middleware.PermissionMiddleware("system:post:add"), postHandler.AddPost)          // 新增岗位
		postGroup.PUT("", middleware.PermissionMiddleware("system:post:edit"), postHandler.UpdatePost)       // 修改岗位
		postGroup.DELETE(":ids", middleware.PermissionMiddleware("system:post:remove"), postHandler.DelPost) // 删除岗位
	}
}
