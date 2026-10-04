package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysJob struct {
	Id               uint64 `gorm:"column:id;primary_key;" json:"id"        description:"任务id"`
	JobName          string `gorm:"column:job_name" json:"jobName"  description:"任务名称" binding:"required" err_msg:"任务名称 必填"`
	JobGroup         string `gorm:"column:job_group;default:'DEFAULT'" json:"jobGroup"  description:"任务组名"`
	InvokeTarget     string `gorm:"column:invoke_target" json:"invokeTarget"  description:"调用目标字符串" binding:"required" err_msg:"调用方法 必填"`
	InvokeTargetArgs string `gorm:"column:invoke_target_args" json:"invokeTargetArgs"  description:"调用目标参数"`
	CronExpression   string `gorm:"column:cron_expression" json:"cronExpression"  description:"cron执行表达式" binding:"required" err_msg:"cron表达式 必填"`
	MisfirePolicy    string `gorm:"column:misfire_policy;default:'3'" json:"misfirePolicy"  description:"计划执行错误策略（1立即执行 2执行一次 3放弃执行）"`
	Concurrent       string `gorm:"column:concurrent;default:'1'" json:"concurrent" description:"是否并发执行（0允许 1禁止）"`
	Status           string `gorm:"column:status;default:'0'" json:"status"    description:"状态（0正常 1暂停）"`
	Remark           string `gorm:"column:remark" json:"remark"    description:"备注信息"`
	NextValidTime    string `gorm:"-" json:"nextValidTime"    description:"下次执行时间"`
	BaseNoDelModel
}

func (c SysJob) TableName() string {
	return "sys_job"
}

func (c SysJob) Get(query string, args ...interface{}) (SysJob, error) {
	var data SysJob
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysJob) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(SysJob{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysJob) Create(data *SysJob) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysJob) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysJob{})
	return d.Error
}

func (c SysJob) SelectJobList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysJob, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysJob{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysJob, 0)
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysJob) SelectJobAllList(conditions map[string]interface{}) ([]SysJob, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysJob{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysJob, 0)
	d := query.Order("create_time desc").Find(&data)

	return data, d.Error
}
