package system

import (
	"go-fin-server/internal/model"
)

type SysNoticeService struct {
}

// SelectNoticeList 查询公告列表
func (c SysNoticeService) SelectNoticeList(conditions map[string]interface{}) ([]model.SysNotice, int64, error) {
	data, count, err := model.SysNotice{}.SelectNoticeList(conditions)
	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// SelectNoticeById 查询公告信息
func (c SysNoticeService) SelectNoticeById(noticeId uint64) (model.SysNotice, error) {
	data, err := model.SysNotice{}.Get("id = ?", noticeId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// InsertNotice 新增通知公告
func (c SysNoticeService) InsertNotice(data *model.SysNotice) error {
	err := data.Create(data)
	if err != nil {
		return err
	}
	return nil
}

// UpdateNotice 修改通知公告
func (c SysNoticeService) UpdateNotice(noticeId uint64, upd map[string]interface{}) error {
	err := model.SysNotice{}.UpdateMap(upd, "id = ?", noticeId)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNoticeByIds 批量删除公告信息
func (c SysNoticeService) DeleteNoticeByIds(noticeIds []uint64) error {
	err := model.SysNotice{}.Delete("id in (?)", noticeIds)
	if err != nil {
		return err
	}
	return nil
}

// MarkNoticeRead 标记单条公告已读
func (c SysNoticeService) MarkNoticeRead(noticeId, userId uint64, userName string) error {
	return model.SysNoticeRead{}.MarkRead(noticeId, userId, userName)
}

// MarkNoticeReadBatch 批量标记公告已读
func (c SysNoticeService) MarkNoticeReadBatch(noticeIds []uint64, userId uint64, userName string) error {
	return model.SysNoticeRead{}.MarkReadBatch(noticeIds, userId, userName)
}

// CountUnreadNotice 统计当前用户未读公告数量
func (c SysNoticeService) CountUnreadNotice(userId uint64) (int64, error) {
	return model.SysNoticeRead{}.CountUnread(userId)
}
