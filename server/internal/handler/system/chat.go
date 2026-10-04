package system

import (
	"go-fin-server/internal/config"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/redistool"
	ws "go-fin-server/pkg/websocket"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ChatHandler WebSocket 聊天处理器
type ChatHandler struct {
	hub *ws.Hub
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(hub *ws.Hub) *ChatHandler {
	return &ChatHandler{hub: hub}
}

// HandleWebSocket 处理 WebSocket 连接升级
//
//	@Summary	WebSocket 聊天连接
//	@Tags		在线聊天
//	@Produce	json
//	@Param		token	query	string	true	"JWT Token"
//	@Router		/ws/chat [get]
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	// 从 URL 参数中获取 token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 401, "msg": "缺少认证 token"})
		return
	}

	// 解析 JWT token 获取 LoginUserKey
	loginUserKey, err := parseTokenAndGetLoginUserKey(token, config.GlobalConfig.Jwt.Secret)
	if err != nil {
		pkg.Logger.Warnf("WebSocket 认证失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "无效的 token"})
		return
	}

	// 从 Redis 获取 LoginUser 信息
	userKey := constant.CACHE_LOGIN_TOKEN_KEY + loginUserKey
	var loginUser model.LoginUser
	err = redistool.GetCacheObject(db.RedisConnections["master"], userKey, &loginUser)
	if err != nil {
		pkg.Logger.Warnf("WebSocket 获取用户信息失败: userKey=%s, error=%v", userKey, err)
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未登录或登录已过期"})
		return
	}

	// 查询用户岗位（岗位名称），用于在线用户面板展示
	deptName := loginUser.DeptName
	postName := ""
	posts, postErr := model.SysPost{}.SelectPostsByUserName(loginUser.UserName)
	if postErr == nil && len(posts) > 0 {
		postName = posts[0].PostName
	}

	// 升级到 WebSocket 连接
	ws.ServeWs(h.hub, c.Writer, c.Request, loginUser.UserId, loginUser.UserName, loginUser.User.NickName, loginUser.User.Avatar, deptName, postName)
}

// GetOnlineUsers 获取在线用户列表
//
//	@Summary	获取在线用户列表
//	@Tags		在线聊天
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"在线用户列表"
//	@Router		/chat/online [get]
//	@Security	BearerAuth
func (h *ChatHandler) GetOnlineUsers(c *gin.Context) {
	// 优先从 Redis 获取在线用户（支持多服务器部署）
	users := h.hub.GetOnlineUsersFromRedis()
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": users,
	})
}

// GetOnlineCount 获取在线用户数量
//
//	@Summary	获取在线用户数量
//	@Tags		在线聊天
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"在线用户数量"
//	@Router		/chat/online/count [get]
//	@Security	BearerAuth
func (h *ChatHandler) GetOnlineCount(c *gin.Context) {
	// 优先从 Redis 获取在线用户数量（支持多服务器部署）
	count := h.hub.GetOnlineCountFromRedis()
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": count,
	})
}

// GetHistory 获取聊天历史消息（私聊/群聊，按时间升序返回最近 limit 条）
//
//	@Summary	获取聊天历史消息
//	@Tags		在线聊天
//	@Produce	json
//	@Param		otherId		query		int		false	"私聊对方用户ID"
//	@Param		groupName	query		string	false	"群聊名称"
//	@Param		limit		query		int		false	"数量限制(默认100)"
//	@Success	200	{object}	map[string]interface{}"聊天历史"
//	@Router		/chat/history [get]
//	@Security	BearerAuth
func (h *ChatHandler) GetHistory(c *gin.Context) {
	// /chat 分组注册在 JWT 中间件之前，需手动解析 Authorization: Bearer 获取当前用户
	loginUser, err := h.loginUserFromRequest(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": err.Error()})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}

	var (
		records []model.ChatRecord
		err2    error
	)

	groupName := c.Query("groupName")
	otherIdStr := c.Query("otherId")
	if groupName != "" {
		// 群聊历史
		records, err2 = model.ChatRecord{}.SelectGroupHistory(groupName, limit)
	} else if otherIdStr != "" {
		otherId, parseErr := strconv.ParseUint(otherIdStr, 10, 64)
		if parseErr != nil {
			c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "otherId 参数格式错误", "data": nil})
			return
		}
		records, err2 = model.ChatRecord{}.SelectPrivateHistory(loginUser.UserId, otherId, limit)
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请提供 otherId 或 groupName 参数", "data": nil})
		return
	}

	if err2 != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "查询聊天历史失败: " + err2.Error(), "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": records,
	})
}

// loginUserFromRequest 从 Authorization: Bearer 解析当前登录用户
func (h *ChatHandler) loginUserFromRequest(c *gin.Context) (*model.LoginUser, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return nil, fmt.Errorf("缺少认证信息")
	}
	token := authHeader[7:]
	loginUserKey, err := parseTokenAndGetLoginUserKey(token, config.GlobalConfig.Jwt.Secret)
	if err != nil {
		return nil, fmt.Errorf("无效的 token")
	}
	userKey := constant.CACHE_LOGIN_TOKEN_KEY + loginUserKey
	var loginUser model.LoginUser
	if err := redistool.GetCacheObject(db.RedisConnections["master"], userKey, &loginUser); err != nil {
		return nil, fmt.Errorf("用户未登录或登录已过期")
	}
	return &loginUser, nil
}

// parseTokenAndGetLoginUserKey 解析 JWT token 并返回 LoginUserKey
func parseTokenAndGetLoginUserKey(tokenString string, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ChatClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*ChatClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	return claims.LoginUserKey, nil
}

// ChatClaims JWT Claims 结构（仅用于 WebSocket 认证）
type ChatClaims struct {
	LoginUserKey string `json:"loginUserKey"`
	jwt.RegisteredClaims
}
