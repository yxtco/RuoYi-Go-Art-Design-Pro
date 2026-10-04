package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysNotice struct {
	Id            uint64 `gorm:"column:id;primary_key" json:"id"    description:"公告ID"`
	NoticeTitle   string `gorm:"column:notice_title" json:"noticeTitle"  description:"公告标题" binding:"required" err_msg:"公告标题 必填"`
	NoticeType    string `gorm:"column:notice_type" json:"noticeType"  description:"公告类型（1通知 2公告）" binding:"required" err_msg:"公告类型 必填"`
	NoticeContent string `gorm:"column:notice_content" json:"noticeContent"  description:"公告内容"`
	Status        string `gorm:"column:status;default:'0'" json:"status"    description:"公告状态（0正常 1关闭）"`
	Remark        string `gorm:"column:remark" json:"remark"    description:"备注"`
	BaseNoDelModel
}

func (c SysNotice) TableName() string {
	return "sys_notice"
}

func (c SysNotice) Get(query string, args ...interface{}) (SysNotice, error) {
	var data SysNotice
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysNotice) Create(data *SysNotice) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysNotice) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(&SysNotice{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysNotice) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysNotice{})
	return d.Error
}

func (c SysNotice) SelectNoticeList(conditions map[string]interface{}) ([]SysNotice, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysNotice{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysNotice, 0)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	d := query.Order("create_time desc").Find(&data)
	return data, count, d.Error
}
