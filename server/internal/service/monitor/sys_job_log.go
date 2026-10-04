package monitor

import "go-fin-server/internal/model"

type SysJobLogService struct {
}

// SelectJobLogList 获取分页定时任务日志
func (c SysJobLogService) SelectJobLogList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysJobLog, int64, error) {
	data, count, err := model.SysJobLog{}.SelectJobLogList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return nil, count, err
	}
	return data, count, nil
}

// SelectJobLogAllList 获取定时任务日志
func (c SysJobLogService) SelectJobLogAllList(conditions map[string]interface{}) ([]model.SysJobLog, error) {
	data, err := model.SysJobLog{}.SelectJobLogAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectJobLogById 获取定时任务日志
func (c SysJobLogService) SelectJobLogById(jobId uint64) (model.SysJobLog, error) {
	data, err := model.SysJobLog{}.Get("id = ?", jobId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// DeleteJobLogByIds 批量删除定时任务日志
func (c SysJobLogService) DeleteJobLogByIds(jobIds []uint64) error {
	err := model.SysJobLog{}.Delete("id in (?)", jobIds)
	if err != nil {
		return err
	}
	return nil
}

// CleanJobLog 清空任务日志
func (c SysJobLogService) CleanJobLog() error {
	err := model.SysJobLog{}.Clean()
	if err != nil {
		return err
	}
	return nil
}
