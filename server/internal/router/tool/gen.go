package tool

import (
	"go-fin-server/internal/handler/tool"

	"github.com/gin-gonic/gin"
)

func SetupGenRoutes(g *gin.RouterGroup) {
	genHandler := tool.NewGenHandler()
	genGroup := g.Group("/gen")
	{
		genGroup.GET("list", genHandler.ListTable)           // 查询表数据
		genGroup.GET("db/list", genHandler.ListDbTable)      // 查询数据库列表
		genGroup.GET(":id", genHandler.GetTable)             // 查询表详细信息
		genGroup.PUT("", genHandler.UpdateTable)             // 修改代码生成信息
		genGroup.POST("importTable", genHandler.ImportTable) // 导入表
		genGroup.POST("createTable", genHandler.CreateTable) // 创建表
		genGroup.GET("preview/:id", genHandler.Preview)      // 预览代码
		genGroup.DELETE(":id", genHandler.DeleteTable)       // 删除表
		genGroup.GET("genCode/:tableName", genHandler.GenCode)   // 生成代码
		genGroup.GET("synchDb/:tableName", genHandler.SynchDb)   // 同步数据库
	}
}
