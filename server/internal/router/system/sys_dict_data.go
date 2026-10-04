package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDictDataRoutes(g *gin.RouterGroup) {
	dictDataHandler := system.NewDictDataHandler()
	DictDataGroup := g.Group("/dict/data")
	{
		DictDataGroup.GET("list", middleware.PermissionMiddleware("system:dict:list"), dictDataHandler.ListDictData)     // 查询字典数据列表
		DictDataGroup.POST("export", middleware.PermissionMiddleware("system:dict:export"), dictDataHandler.Export)      // 导出
		DictDataGroup.GET(":id", middleware.PermissionMiddleware("system:dict:query"), dictDataHandler.GetDictData)      // 查询字典数据详细
		DictDataGroup.POST("types", dictDataHandler.GetDicts)                                                            // 根据字典类型查询字典数据信息
		DictDataGroup.POST("", middleware.PermissionMiddleware("system:dict:add"), dictDataHandler.AddDictData)          // 新增字典数据
		DictDataGroup.PUT("", middleware.PermissionMiddleware("system:dict:edit"), dictDataHandler.UpdateDictData)       // 修改字典数据
		DictDataGroup.DELETE(":ids", middleware.PermissionMiddleware("system:dict:remove"), dictDataHandler.DelDictData) // 删除字典数据

	}
}
