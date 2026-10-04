package system

import (
	"go-fin-server/internal/handler/system"
	ws "go-fin-server/pkg/websocket"

	"github.com/gin-gonic/gin"
)

// SetupChatRoutes 设置聊天相关路由
// 注意：WebSocket 路由需要在 JWT 中间件之前注册
func SetupChatRoutes(r *gin.Engine, hub *ws.Hub) {
	chatHandler := system.NewChatHandler(hub)

	// WebSocket 聊天连接（无需 JWT 中间件，token 通过 URL 参数传递）
	r.GET("/ws/chat", chatHandler.HandleWebSocket)

	// 聊天相关 HTTP API（需要 JWT 认证，走常规中间件）
	chatGroup := r.Group("/chat")
	{
		chatGroup.GET("online", chatHandler.GetOnlineUsers)       // 获取在线用户列表
		chatGroup.GET("online/count", chatHandler.GetOnlineCount) // 获取在线用户数量
		chatGroup.GET("history", chatHandler.GetHistory)          // 获取聊天历史（私聊/群聊）
	}
}
