package monitor

import (
	"go-fin-server/internal/model"
	"go-fin-server/pkg"

	"gorm.io/gorm"
)

type SysJobService struct {
}

// SelectJobList 获取分页定时任务
func (c SysJobService) SelectJobList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysJob, int64, error) {
	data, count, err := model.SysJob{}.SelectJobList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return nil, count, err
	}
	return data, count, nil
}

// SelectJobAllList 查询定时任务
func (c SysJobService) SelectJobAllList(conditions map[string]interface{}) ([]model.SysJob, error) {
	data, err := model.SysJob{}.SelectJobAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectJobById 获取定时任务信息
func (c SysJobService) SelectJobById(jobId uint64) (model.SysJob, error) {
	data, err := model.SysJob{}.Get("id = ?", jobId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CheckJobNameUnique 校验任务名称是否唯一
func (c SysJobService) CheckJobNameUnique(jobId uint64, jobName string) bool {
	var job model.SysJob
	obj, err := job.Get("job_name = ?", jobName)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if obj.Id != jobId {
		return false
	}
	return true
}

// UpdateJobStatus 修改任务信息状态
func (c SysJobService) UpdateJobStatus(jobId uint64, upd map[string]interface{}) error {
	err := model.SysJob{}.UpdateMap(upd, "id = ?", jobId)
	if err != nil {
		return err
	}
	return nil
}

// UpdateJob 修改任务信息状态
func (c SysJobService) UpdateJob(jobId uint64, upd map[string]interface{}) error {
	err := model.SysJob{}.UpdateMap(upd, "id = ?", jobId)
	if err != nil {
		return err
	}
	return nil
}

// InsertJob 新增定时任务
func (c SysJobService) InsertJob(data *model.SysJob) error {
	err := data.Create(data)
	if err != nil {
		return err
	}
	return nil
}

// DeleteJobByIds 批量删除定时任务
func (c SysJobService) DeleteJobByIds(jobIds []uint64) error {
	err := model.SysJob{}.Delete("id in (?)", jobIds)
	if err != nil {
		return err
	}
	return nil
}
