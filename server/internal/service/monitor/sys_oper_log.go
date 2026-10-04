package monitor

import "go-fin-server/internal/model"

type SysOperLogService struct {
}

// SelectOperLogList 根据条件分页查询操作日志数据
func (c SysOperLogService) SelectOperLogList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysOperLog, int64, error) {
	var log model.SysOperLog
	data, total, err := log.SelectOperLogList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return data, total, err
	}
	return data, total, nil
}

// SelectOperLogAllList 根据条件查询操作日志数据
func (c SysOperLogService) SelectOperLogAllList(conditions map[string]interface{}) ([]model.SysOperLog, error) {
	var log model.SysOperLog
	data, err := log.SelectOperLogAllList(conditions)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CleanOperLog 清空操作日志
func (c SysOperLogService) CleanOperLog() error {
	var log model.SysOperLog
	err := log.Clean()
	if err != nil {
		return err
	}
	return nil
}

// DeleteOperLogByIds 批量删除系统操作日志
func (c SysOperLogService) DeleteOperLogByIds(ids []uint64) error {
	err := model.SysOperLog{}.Delete("id in (?)", ids)
	if err != nil {
		return err
	}
	return nil
}
