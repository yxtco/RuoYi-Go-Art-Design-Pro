package model

import (
	"go-fin-server/internal/db"
)

// ChatRecord 聊天消息持久化记录（数据库存储）
// 通过 WebSocket 发送的私聊/群聊消息都会落库，便于历史消息加载
type ChatRecord struct {
	Id         uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`
	FromUserId uint64 `gorm:"column:from_user_id" json:"fromUserId"` // 发送者ID
	FromName   string `gorm:"column:from_name" json:"fromName"`      // 发送者名称
	FromAvatar string `gorm:"column:from_avatar" json:"fromAvatar"`  // 发送者头像
	ToUserId   uint64 `gorm:"column:to_user_id" json:"toUserId"`     // 私聊目标用户ID（群聊为0）
	GroupName  string `gorm:"column:group_name" json:"groupName"`    // 群聊名称（私聊为空）
	Content    string `gorm:"column:content;type:text" json:"content"` // 消息内容
	MsgType    string `gorm:"column:msg_type" json:"msgType"`        // 消息类型: text/image/file/emoji
	FileUrl    string `gorm:"column:file_url" json:"fileUrl"`        // 文件/图片URL
	FileName   string `gorm:"column:file_name" json:"fileName"`      // 文件名
	Timestamp  int64  `gorm:"column:timestamp" json:"timestamp"`     // 时间戳(毫秒)
}

// TableName 表名
func (ChatRecord) TableName() string {
	return "chat_message"
}

// InsertChatRecord 保存一条聊天消息
func (ChatRecord) InsertChatRecord(record *ChatRecord) error {
	return db.DBConnections["master"].Create(record).Error
}

// SelectPrivateHistory 查询与指定用户的私聊历史（双向）
// fromUserId 当前用户，toUserId 对方用户，limit 数量限制
func (ChatRecord) SelectPrivateHistory(fromUserId, toUserId uint64, limit int) ([]ChatRecord, error) {
	data := make([]ChatRecord, 0)
	if limit <= 0 {
		limit = 100
	}
	err := db.DBConnections["master"].Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		fromUserId, toUserId, toUserId, fromUserId,
	).Order("timestamp asc").Limit(limit).Find(&data).Error
	return data, err
}

// SelectGroupHistory 查询群聊历史
func (ChatRecord) SelectGroupHistory(groupName string, limit int) ([]ChatRecord, error) {
	data := make([]ChatRecord, 0)
	if limit <= 0 {
		limit = 100
	}
	err := db.DBConnections["master"].Where("group_name = ?", groupName).
		Order("timestamp asc").Limit(limit).Find(&data).Error
	return data, err
}
