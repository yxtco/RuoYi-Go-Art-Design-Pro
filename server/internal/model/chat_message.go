package model

// ChatMessage WebSocket 传输的聊天消息体
type ChatMessage struct {
	Type       string `json:"type"`       // 消息类型: "private" | "group" | "system"
	Action     string `json:"action"`     // 动作: "message" | "join" | "leave" | "online" | "typing"
	FromId     uint64 `json:"fromId"`     // 发送者ID
	FromName   string `json:"fromName"`   // 发送者用户名
	FromAvatar string `json:"fromAvatar"` // 发送者头像
	ToId       uint64 `json:"toId"`       // 私聊目标用户ID
	GroupName  string `json:"groupName"`  // 群聊群组名称
	Content    string `json:"content"`    // 消息内容（HTML/富文本）
	MsgType    string `json:"msgType"`    // 消息内容类型: "text" | "image" | "file" | "emoji"
	FileUrl    string `json:"fileUrl"`    // 文件/图片URL
	FileName   string `json:"fileName"`   // 文件名
	Timestamp  int64  `json:"timestamp"`  // 时间戳（毫秒）
}

// OnlineUser 在线用户信息
type OnlineUser struct {
	UserId   uint64 `json:"userId"`   // 用户ID
	UserName string `json:"userName"` // 用户名
	NickName string `json:"nickName"` // 用户昵称
	Avatar   string `json:"avatar"`   // 头像
	DeptName string `json:"deptName"` // 部门名称
	PostName string `json:"postName"` // 岗位名称
}

// BroadcastMessage 广播消息封装
type BroadcastMessage struct {
	Message *ChatMessage
	FromId  uint64
}
