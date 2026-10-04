package monitor

import "go-fin-server/internal/model"

type SysLoginInfoService struct {
}

// SelectLoginInfoList 查询系统登录日志集合
func (c SysLoginInfoService) SelectLoginInfoList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysLoginInfo, int64, error) {
	data, total, err := model.SysLoginInfo{}.SelectLoginInfoList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return data, total, err
	}
	return data, total, nil
}

// SelectLoginInfoAllList 查询系统登录日志集合
func (c SysLoginInfoService) SelectLoginInfoAllList(conditions map[string]interface{}) ([]model.SysLoginInfo, error) {
	data, err := model.SysLoginInfo{}.SelectLoginInfoAllList(conditions)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CleanLoginInfo 清空系统登录日志
func (c SysLoginInfoService) CleanLoginInfo() error {
	var log model.SysLoginInfo
	err := log.Clean()
	if err != nil {
		return err
	}
	return nil
}

// DeleteLoginInfoByIds 批量删除系统登录日志
func (c SysLoginInfoService) DeleteLoginInfoByIds(ids []uint64) error {
	err := model.SysLoginInfo{}.Delete("id in (?)", ids)
	if err != nil {
		return err
	}
	return nil
}
