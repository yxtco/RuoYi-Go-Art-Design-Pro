package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Hub 全局 WebSocket 连接管理中心
// 负责管理所有客户端连接、群组以及消息路由
type Hub struct {
	clients    map[uint64]*Client       // userId -> Client 映射
	groups     map[string]map[uint64]bool // groupName -> set(userId)
	register   chan *Client              // 注册通道
	unregister chan *Client              // 注销通道
	broadcast  chan *model.BroadcastMessage // 广播通道
	mu         sync.RWMutex              // 读写锁
	redisClient *redis.Client            // Redis 客户端，用于存储在线用户信息
}

// NewHub 创建并返回一个新的 Hub 实例
func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		clients:     make(map[uint64]*Client),
		groups:      make(map[string]map[uint64]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *model.BroadcastMessage, 256),
		redisClient: redisClient,
	}
}

// Run 启动 Hub 主循环，处理注册、注销和广播
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case bm := <-h.broadcast:
			h.handleBroadcast(bm)
		}
	}
}

// handleRegister 处理客户端注册
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果用户已在线，先注销旧连接
	if oldClient, exists := h.clients[client.userId]; exists {
		close(oldClient.send)
		delete(h.clients, client.userId)
	}

	h.clients[client.userId] = client
	pkg.Logger.Infof("WebSocket 客户端注册成功: userId=%d, userName=%s, 当前在线人数=%d",
		client.userId, client.userName, len(h.clients))

	// 将用户信息存储到 Redis
	h.addOnlineUserToRedis(client)

	// 发送当前在线用户列表给新连接的用户
	h.sendOnlineList(client)

	// 广播用户上线消息给其他用户
	joinMsg := &model.ChatMessage{
		Type:       "system",
		Action:     "join",
		FromId:     client.userId,
		FromName:   client.userName,
		FromAvatar: client.avatar,
		Timestamp:  time.Now().UnixMilli(),
	}
	h.broadcastToOthers(joinMsg, client.userId)
}

// handleUnregister 处理客户端注销
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.userId]; exists {
		close(client.send)
		delete(h.clients, client.userId)

		// 从 Redis 中移除用户
		h.removeOnlineUserFromRedis(client.userId)

		// 从所有群组中移除该用户
		for groupName, members := range h.groups {
			delete(members, client.userId)
			if len(members) == 0 {
				delete(h.groups, groupName)
			}
		}

		pkg.Logger.Infof("WebSocket 客户端注销: userId=%d, userName=%s, 当前在线人数=%d",
			client.userId, client.userName, len(h.clients))

		// 广播用户下线消息
		leaveMsg := &model.ChatMessage{
			Type:       "system",
			Action:     "leave",
			FromId:     client.userId,
			FromName:   client.userName,
			FromAvatar: client.avatar,
			Timestamp:  time.Now().UnixMilli(),
		}
		h.broadcastToAll(leaveMsg)
	}
}

// handleBroadcast 处理广播消息
func (h *Hub) handleBroadcast(bm *model.BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := bm.Message
	switch msg.Type {
	case "private":
		h.routePrivateMessage(bm)
	case "group":
		h.routeGroupMessage(bm)
	}
}

// routePrivateMessage 路由私聊消息
func (h *Hub) routePrivateMessage(bm *model.BroadcastMessage) {
	msg := bm.Message
	targetClient, exists := h.clients[msg.ToId]
	if !exists {
		// 目标用户不在线，发送系统提示给发送者
		if sender, ok := h.clients[bm.FromId]; ok {
			offlineMsg := &model.ChatMessage{
				Type:      "system",
				Action:    "message",
				Content:   "对方不在线，消息无法送达",
				ToId:      msg.ToId,
				MsgType:   "text",
				Timestamp: time.Now().UnixMilli(),
			}
			data, _ := json.Marshal(offlineMsg)
			select {
			case sender.send <- data:
			default:
			}
		}
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		pkg.Logger.Errorf("序列化私聊消息失败: %v", err)
		return
	}

	select {
	case targetClient.send <- data:
	default:
		pkg.Logger.Warnf("私聊消息发送失败，目标用户 send channel 已满: toId=%d", msg.ToId)
	}
}

// routeGroupMessage 路由群聊消息
func (h *Hub) routeGroupMessage(bm *model.BroadcastMessage) {
	msg := bm.Message
	members, exists := h.groups[msg.GroupName]
	if !exists {
		return
	}

	data, err := json.Marshal(msg)
	if err != nil {
		pkg.Logger.Errorf("序列化群聊消息失败: %v", err)
		return
	}

	for userId := range members {
		if userId == bm.FromId {
			continue // 不发给发送者自己
		}
		if client, ok := h.clients[userId]; ok {
			select {
			case client.send <- data:
			default:
				pkg.Logger.Warnf("群聊消息发送失败，用户 send channel 已满: userId=%d", userId)
			}
		}
	}
}

// broadcastToOthers 广播消息给除指定用户外的所有在线用户
func (h *Hub) broadcastToOthers(msg *model.ChatMessage, excludeUserId uint64) {
	data, err := json.Marshal(msg)
	if err != nil {
		pkg.Logger.Errorf("序列化广播消息失败: %v", err)
		return
	}

	for userId, client := range h.clients {
		if userId == excludeUserId {
			continue
		}
		select {
		case client.send <- data:
		default:
		}
	}
}

// broadcastToAll 广播消息给所有在线用户
func (h *Hub) broadcastToAll(msg *model.ChatMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		pkg.Logger.Errorf("序列化广播消息失败: %v", err)
		return
	}

	for _, client := range h.clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

// sendOnlineList 发送当前在线用户列表给指定客户端
func (h *Hub) sendOnlineList(client *Client) {
	onlineUsers := make([]model.OnlineUser, 0, len(h.clients))
	for _, c := range h.clients {
		onlineUsers = append(onlineUsers, model.OnlineUser{
			UserId:   c.userId,
			UserName: c.userName,
			NickName: c.nickName,
			Avatar:   c.avatar,
			DeptName: c.deptName,
			PostName: c.postName,
		})
	}

	onlineMsg := &model.ChatMessage{
		Type:      "system",
		Action:    "online",
		Content:   "",
		Timestamp: time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(onlineMsg)

	// 将在线用户列表作为特殊消息发送
	type onlineListMsg struct {
		*model.ChatMessage
		OnlineUsers []model.OnlineUser `json:"onlineUsers"`
	}
	msg := onlineListMsg{
		ChatMessage: onlineMsg,
		OnlineUsers: onlineUsers,
	}
	msgData, _ := json.Marshal(msg)
	_ = data // 使用 msgData 替代

	select {
	case client.send <- msgData:
	default:
	}
}

// Register 注册客户端到 Hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 从 Hub 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast 提交广播消息到 Hub
func (h *Hub) Broadcast(bm *model.BroadcastMessage) {
	h.broadcast <- bm
}

// JoinGroup 用户加入群组
func (h *Hub) JoinGroup(groupName string, userId uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.groups[groupName]; !exists {
		h.groups[groupName] = make(map[uint64]bool)
	}
	h.groups[groupName][userId] = true
	pkg.Logger.Infof("用户加入群组: userId=%d, group=%s", userId, groupName)
}

// LeaveGroup 用户离开群组
func (h *Hub) LeaveGroup(groupName string, userId uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if members, exists := h.groups[groupName]; exists {
		delete(members, userId)
		if len(members) == 0 {
			delete(h.groups, groupName)
		}
	}
}

// GetOnlineUsers 获取当前在线用户列表
func (h *Hub) GetOnlineUsers() []model.OnlineUser {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]model.OnlineUser, 0, len(h.clients))
	for _, client := range h.clients {
		users = append(users, model.OnlineUser{
			UserId:   client.userId,
			UserName: client.userName,
			NickName: client.nickName,
			Avatar:   client.avatar,
		})
	}
	return users
}

// IsOnline 判断用户是否在线
func (h *Hub) IsOnline(userId uint64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.clients[userId]
	return exists
}

// GetOnlineCount 获取在线用户数量
func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ==================== Redis 在线用户管理 ====================

// addOnlineUserToRedis 将用户信息添加到 Redis
func (h *Hub) addOnlineUserToRedis(client *Client) {
	if h.redisClient == nil {
		return
	}

	userInfo := model.OnlineUser{
		UserId:   client.userId,
		UserName: client.userName,
		NickName: client.nickName,
		Avatar:   client.avatar,
		DeptName: client.deptName,
		PostName: client.postName,
	}

	userJson, err := json.Marshal(userInfo)
	if err != nil {
		pkg.Logger.Errorf("序列化在线用户信息失败: %v", err)
		return
	}

	ctx := context.Background()
	err = h.redisClient.HSet(ctx, constant.CACHE_ONLINE_USERS_KEY, fmt.Sprintf("%d", client.userId), userJson).Err()
	if err != nil {
		pkg.Logger.Errorf("存储在线用户到 Redis 失败: %v", err)
	}
}

// removeOnlineUserFromRedis 从 Redis 中移除用户
func (h *Hub) removeOnlineUserFromRedis(userId uint64) {
	if h.redisClient == nil {
		return
	}

	ctx := context.Background()
	err := h.redisClient.HDel(ctx, constant.CACHE_ONLINE_USERS_KEY, fmt.Sprintf("%d", userId)).Err()
	if err != nil {
		pkg.Logger.Errorf("从 Redis 移除在线用户失败: %v", err)
	}
}

// GetOnlineUsersFromRedis 从 Redis 获取所有在线用户
func (h *Hub) GetOnlineUsersFromRedis() []model.OnlineUser {
	if h.redisClient == nil {
		return h.GetOnlineUsers()
	}

	ctx := context.Background()
	result, err := h.redisClient.HGetAll(ctx, constant.CACHE_ONLINE_USERS_KEY).Result()
	if err != nil {
		pkg.Logger.Errorf("从 Redis 获取在线用户失败: %v", err)
		return h.GetOnlineUsers()
	}

	users := make([]model.OnlineUser, 0, len(result))
	for _, userJson := range result {
		var user model.OnlineUser
		if err := json.Unmarshal([]byte(userJson), &user); err != nil {
			continue
		}
		users = append(users, user)
	}

	return users
}

// GetOnlineCountFromRedis 从 Redis 获取在线用户数量
func (h *Hub) GetOnlineCountFromRedis() int64 {
	if h.redisClient == nil {
		return int64(h.GetOnlineCount())
	}

	ctx := context.Background()
	count, err := h.redisClient.HLen(ctx, constant.CACHE_ONLINE_USERS_KEY).Result()
	if err != nil {
		pkg.Logger.Errorf("从 Redis 获取在线用户数量失败: %v", err)
		return int64(h.GetOnlineCount())
	}

	return count
}
