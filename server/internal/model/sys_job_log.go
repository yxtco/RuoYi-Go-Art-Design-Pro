package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/types"
	"go-fin-server/pkg/utils"
)

type SysJobLog struct {
	Id               uint64          `gorm:"column:id;primary_key;" json:"id"        description:"任务id"`
	JobName          string          `gorm:"column:job_name" json:"jobName"  description:"任务名称"`
	JobGroup         string          `gorm:"column:job_group" json:"jobGroup"  description:"任务组名"`
	InvokeTarget     string          `gorm:"column:invoke_target" json:"invokeTarget"  description:"调用目标字符串"`
	InvokeTargetArgs string          `gorm:"column:invoke_target_args" json:"invokeTargetArgs"  description:"调用目标参数"`
	JobMessage       string          `gorm:"column:job_message" json:"jobMessage"  description:"日志信息"`
	ExceptionInfo    string          `gorm:"column:exception_info" json:"exceptionInfo"  description:"调用目标参数"`
	Status           string          `gorm:"column:status;default:'0'" json:"status"    description:"状态（0正常 1暂停）"`
	JobUpdatedTime   types.LocalTime `gorm:"column:job_update_time" json:"jobUpdateTime"     description:"任务更新时间"`
	CreatedAt        types.LocalTime `gorm:"column:create_time" json:"createTime"     description:"创建时间"`
	FinishTime       types.LocalTime `gorm:"column:finish_time" json:"finish_time"     description:"完成时间"`
}

func (c SysJobLog) TableName() string {
	return "sys_job_log"
}

func (c SysJobLog) Create(data *SysJobLog) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysJobLog) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	d := db.DBConnections["master"].Model(SysJobLog{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysJobLog) SelectJobLogList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysJobLog, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysJobLog{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysJobLog, 0)
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

func (c SysJobLog) Get(query string, args ...interface{}) (SysJobLog, error) {
	var data SysJobLog
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysJobLog) SelectJobLogAllList(conditions map[string]interface{}) ([]SysJobLog, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysJobLog{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysJobLog, 0)
	d := query.Order("create_time desc").Find(&data)

	return data, d.Error
}

func (c SysJobLog) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysJobLog{})
	return d.Error
}

func (c SysJobLog) Clean() error {
	sql := `
		truncate table sys_job_log;
	`
	d := db.DBConnections["master"].Exec(sql)
	return d.Error
}
