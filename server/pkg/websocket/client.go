package websocket

import (
	"encoding/json"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"net/http"
	"time"

	gorilla "github.com/gorilla/websocket"
)

const (
	// writeWait 写操作超时时间
	writeWait = 10 * time.Second
	// pongWait 等待 Pong 消息的超时时间
	pongWait = 60 * time.Second
	// pingPeriod 发送 Ping 消息的周期，必须小于 pongWait
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize 最大消息大小
	maxMessageSize = 65536
	// sendBufferSize 发送通道缓冲区大小
	sendBufferSize = 256
)

// Client 代表一个 WebSocket 客户端连接
type Client struct {
	hub      *Hub              // 所属的 Hub
	conn     *gorilla.Conn     // WebSocket 连接
	send     chan []byte       // 消息发送通道
	userId   uint64            // 用户ID
	userName string            // 用户名
	nickName string            // 用户昵称
	avatar   string            // 用户头像
	deptName string            // 部门名称
	postName string            // 岗位名称
}

// NewClient 创建一个新的客户端连接
func NewClient(hub *Hub, conn *gorilla.Conn, userId uint64, userName, nickName, avatar, deptName, postName string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		userId:   userId,
		userName: userName,
		nickName: nickName,
		avatar:   avatar,
		deptName: deptName,
		postName: postName,
	}
}

// ReadPump 持续读取 WebSocket 消息
// 解析消息类型后交给 Hub 进行路由
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if gorilla.IsUnexpectedCloseError(err, gorilla.CloseGoingAway, gorilla.CloseNormalClosure) {
				pkg.Logger.Warnf("WebSocket 读取异常关闭: userId=%d, error=%v", c.userId, err)
			} else {
				pkg.Logger.Infof("WebSocket 连接关闭: userId=%d", c.userId)
			}
			break
		}

		c.handleMessage(message)
	}
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(message []byte) {
	var chatMsg model.ChatMessage
	if err := json.Unmarshal(message, &chatMsg); err != nil {
		pkg.Logger.Warnf("解析聊天消息失败: userId=%d, error=%v", c.userId, err)
		return
	}

	// 补充发送者信息
	chatMsg.FromId = c.userId
	chatMsg.FromName = c.userName
	chatMsg.FromAvatar = c.avatar
	if chatMsg.Timestamp == 0 {
		chatMsg.Timestamp = time.Now().UnixMilli()
	}

	// 持久化聊天消息（私聊/群聊，落库便于历史加载）
	if chatMsg.Type == "private" || chatMsg.Type == "group" {
		record := &model.ChatRecord{
			FromUserId: c.userId,
			FromName:   c.userName,
			FromAvatar: c.avatar,
			ToUserId:   chatMsg.ToId,
			GroupName:  chatMsg.GroupName,
			Content:    chatMsg.Content,
			MsgType:    chatMsg.MsgType,
			FileUrl:    chatMsg.FileUrl,
			FileName:   chatMsg.FileName,
			Timestamp:  chatMsg.Timestamp,
		}
		// 注意：Go 解析器限制不能将复合字面量方法调用用于 if 初值，需先赋值
		saveErr := model.ChatRecord{}.InsertChatRecord(record)
		if saveErr != nil {
			pkg.Logger.Errorf("持久化聊天消息失败: fromId=%d, error=%v", c.userId, saveErr)
		}
	}

	switch chatMsg.Type {
	case "private":
		// 私聊消息
		if chatMsg.ToId == 0 {
			pkg.Logger.Warnf("私聊消息缺少目标用户ID: fromId=%d", c.userId)
			return
		}
		c.hub.Broadcast(&model.BroadcastMessage{
			Message: &chatMsg,
			FromId:  c.userId,
		})

	case "group":
		// 群聊消息
		if chatMsg.GroupName == "" {
			pkg.Logger.Warnf("群聊消息缺少群组名称: fromId=%d", c.userId)
			return
		}
		c.hub.Broadcast(&model.BroadcastMessage{
			Message: &chatMsg,
			FromId:  c.userId,
		})

	case "system":
		// 系统消息处理（如加入/离开群组）
		c.handleSystemMessage(&chatMsg)

	default:
		pkg.Logger.Warnf("未知的消息类型: type=%s, fromId=%d", chatMsg.Type, c.userId)
	}
}

// handleSystemMessage 处理系统类型消息
func (c *Client) handleSystemMessage(msg *model.ChatMessage) {
	switch msg.Action {
	case "ping":
		// 应用层心跳：前端 ping，后端立即回 pong，用于判断在线状态
		pongMsg := &model.ChatMessage{
			Type:      "system",
			Action:    "pong",
			Timestamp: time.Now().UnixMilli(),
		}
		if data, err := json.Marshal(pongMsg); err == nil {
			select {
			case c.send <- data:
			default:
				pkg.Logger.Warnf("发送 pong 心跳失败，send 通道已满: userId=%d", c.userId)
			}
		}
	case "joinGroup":
		// 加入群组
		if msg.GroupName != "" {
			c.hub.JoinGroup(msg.GroupName, c.userId)
		}
	case "leaveGroup":
		// 离开群组
		if msg.GroupName != "" {
			c.hub.LeaveGroup(msg.GroupName, c.userId)
		}
	case "typing":
		// 正在输入提示（私聊场景）
		if msg.ToId > 0 {
			typingMsg := &model.ChatMessage{
				Type:      "system",
				Action:    "typing",
				FromId:    c.userId,
				FromName:  c.userName,
				ToId:      msg.ToId,
				Timestamp: time.Now().UnixMilli(),
			}
			c.hub.Broadcast(&model.BroadcastMessage{
				Message: typingMsg,
				FromId:  c.userId,
			})
		}
	}
}

// WritePump 持续监听 send channel，将消息写入 WebSocket 连接
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// send channel 已关闭，发送 Close 消息
				c.conn.WriteMessage(gorilla.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(gorilla.TextMessage)
			if err != nil {
				pkg.Logger.Errorf("WebSocket 获取写入器失败: userId=%d, error=%v", c.userId, err)
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				pkg.Logger.Errorf("WebSocket 关闭写入器失败: userId=%d, error=%v", c.userId, err)
				return
			}

		case <-ticker.C:
			// 发送 Ping 心跳
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gorilla.PingMessage, []byte{}); err != nil {
				pkg.Logger.Warnf("WebSocket 发送 Ping 失败: userId=%d, error=%v", c.userId, err)
				return
			}
		}
	}
}

// ServeWs 处理 WebSocket 升级请求
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, userId uint64, userName, nickName, avatar, deptName, postName string) {
	upgrader := gorilla.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pkg.Logger.Errorf("WebSocket 升级失败: userId=%d, error=%v", userId, err)
		return
	}

	client := NewClient(hub, conn, userId, userName, nickName, avatar, deptName, postName)
	hub.Register(client)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}
