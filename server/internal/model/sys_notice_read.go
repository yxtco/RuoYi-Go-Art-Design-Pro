package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/types"
	"time"

	"gorm.io/gorm/clause"
)

// SysNoticeRead 公告已读记录
type SysNoticeRead struct {
	Id         uint64          `gorm:"column:id;primary_key" json:"id" description:"主键"`
	NoticeId   uint64          `gorm:"column:notice_id" json:"noticeId" description:"公告ID"`
	UserId     uint64          `gorm:"column:user_id" json:"userId" description:"用户ID"`
	ReadTime   types.LocalTime `gorm:"column:read_time" json:"readTime" description:"阅读时间"`
	CreateBy   string          `gorm:"column:create_by" json:"createBy" description:"创建者"`
	CreateTime types.LocalTime `gorm:"column:create_time" json:"createTime" description:"创建时间"`
}

func (c SysNoticeRead) TableName() string {
	return "sys_notice_read"
}

// Insert 插入已读记录
func (c SysNoticeRead) Insert(data *SysNoticeRead) error {
	return db.DBConnections["master"].Create(&data).Error
}

// BatchInsert 批量插入已读记录
func (c SysNoticeRead) BatchInsert(records []SysNoticeRead) error {
	if len(records) == 0 {
		return nil
	}
	return db.DBConnections["master"].Create(&records).Error
}

// Exists 判断是否已存在已读记录
func (c SysNoticeRead) Exists(noticeId, userId uint64) (bool, error) {
	var count int64
	err := db.DBConnections["master"].Model(&SysNoticeRead{}).
		Where("notice_id = ? AND user_id = ?", noticeId, userId).
		Count(&count).Error
	return count > 0, err
}

// MarkRead 标记单条公告已读（使用 INSERT IGNORE 避免唯一索引冲突）
func (c SysNoticeRead) MarkRead(noticeId, userId uint64, userName string) error {
	now := types.LocalTime{Time: time.Now()}
	record := SysNoticeRead{
		NoticeId:   noticeId,
		UserId:     userId,
		ReadTime:   now,
		CreateBy:   userName,
		CreateTime: now,
	}
	// 使用 INSERT IGNORE 避免并发场景下唯一索引冲突报错
	return db.DBConnections["master"].
		Clauses(clause.Insert{Modifier: "IGNORE"}).
		Create(&record).Error
}

// CountUnread 统计当前用户未读的正常公告数量
func (c SysNoticeRead) CountUnread(userId uint64) (int64, error) {
	var count int64
	err := db.DBConnections["master"].Model(&SysNotice{}).
		Where("status = ?", "0").
		Where("id NOT IN (?)",
			db.DBConnections["master"].Model(&SysNoticeRead{}).
				Select("notice_id").
				Where("user_id = ?", userId),
		).
		Count(&count).Error
	return count, err
}

// MarkReadBatch 批量标记公告已读
func (c SysNoticeRead) MarkReadBatch(noticeIds []uint64, userId uint64, userName string) error {
	records := make([]SysNoticeRead, 0, len(noticeIds))
	now := types.LocalTime{Time: time.Now()}
	for _, nid := range noticeIds {
		records = append(records, SysNoticeRead{
			NoticeId:   nid,
			UserId:     userId,
			ReadTime:   now,
			CreateBy:   userName,
			CreateTime: now,
		})
	}
	// 使用 ON DUPLICATE KEY UPDATE 避免唯一索引冲突
	return db.DBConnections["master"].
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "notice_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"read_time"}),
		}).
		Create(&records).Error
}
