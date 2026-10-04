package system

import (
	"go-fin-server/internal/handler/system"
	"go-fin-server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupNoticeRoutes(g *gin.RouterGroup) {
	noticeHandler := system.NewNoticeHandler()
	NoticeGroup := g.Group("/notice")
	{
		NoticeGroup.GET("unreadCount", noticeHandler.GetUnreadCount) // 获取未读公告数量
		NoticeGroup.GET("list", middleware.PermissionMiddleware("system:notice:list"), noticeHandler.ListNotice)     // 查询公告列表
		NoticeGroup.GET(":id", middleware.PermissionMiddleware("system:notice:query"), noticeHandler.GetNotice)      // 查询公告详细
		NoticeGroup.POST("", middleware.PermissionMiddleware("system:notice:add"), noticeHandler.AddNotice)          // 新增公告
		NoticeGroup.PUT("", middleware.PermissionMiddleware("system:notice:edit"), noticeHandler.UpdateNotice)       // 修改公告
		NoticeGroup.DELETE(":ids", middleware.PermissionMiddleware("system:notice:remove"), noticeHandler.DelNotice) // 删除公告
		NoticeGroup.GET("readUsers/list", middleware.PermissionMiddleware("system:notice:list"), noticeHandler.ListNoticeReadUsers) // 查询公告已读用户列表
		NoticeGroup.POST("markRead", noticeHandler.MarkNoticeRead)       // 标记公告已读
		NoticeGroup.POST("markReadAll", noticeHandler.MarkNoticeReadAll) // 批量标记公告已读
	}
}
