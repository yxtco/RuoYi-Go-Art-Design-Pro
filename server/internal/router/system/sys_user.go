package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(g *gin.RouterGroup) {
	userHandler := system.NewUserHandler()
	userGroup := g.Group("/user")
	{
		userGroup.GET("list", middleware.PermissionMiddleware("system:user:list"), userHandler.ListUser)                 // 查询用户列表
		userGroup.POST("export", middleware.PermissionMiddleware("system:user:export"), userHandler.Export)              // 导出excel
		userGroup.POST("importData", middleware.PermissionMiddleware("system:user:import"), userHandler.ImportData)      // 导入excel
		userGroup.POST("importTemplate", userHandler.ImportTemplate)                                                     // excel模版
		userGroup.GET(":id", middleware.PermissionMiddleware("system:user:query"), userHandler.GetUser)                  // 查询用户详细
		userGroup.POST("", middleware.PermissionMiddleware("system:user:add"), userHandler.AddUser)                      // 新增用户
		userGroup.PUT("", middleware.PermissionMiddleware("system:user:edit"), userHandler.UpdateUser)                   // 修改用户
		userGroup.DELETE("/:ids", middleware.PermissionMiddleware("system:user:remove"), userHandler.DelUser)            // 删除用户
		userGroup.PUT("resetPwd", middleware.PermissionMiddleware("system:user:resetPwd"), userHandler.ResetUserPwd)     // 用户密码重置
		userGroup.PUT("changeStatus", middleware.PermissionMiddleware("system:user:edit"), userHandler.ChangeUserStatus) // 用户密码重置
		userGroup.GET("authRole/:id", middleware.PermissionMiddleware("system:user:query"), userHandler.GetAuthRole)     // 查询授权角色
		userGroup.PUT("authRole", middleware.PermissionMiddleware("system:user:edit"), userHandler.UpdateAuthRole)       // 保存授权角色
		userGroup.GET("deptTree", middleware.PermissionMiddleware("system:user:list"), userHandler.DeptTreeSelect)       // 查询部门下拉树结构
		userGroup.GET("profile", userHandler.GetUserProfile)                                                             // 查询用户个人信息
		userGroup.PUT("profile", userHandler.UpdateUserProfile)                                                          // 修改用户个人信息
		userGroup.PUT("profile/updatePwd", userHandler.UpdateUserPwd)                                                    // 用户密码重置
		userGroup.POST("profile/avatar", userHandler.UploadAvatar)                                                       // 用户头像上传

	}

}
