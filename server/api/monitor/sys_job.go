package monitor

import "go-fin-server/api/common"

type SysJob struct {
	ListJobRequest
	UpdateJobRequest
	RunJobRequest
}

type ListJobRequest struct {
	JobName   string `form:"jobName" query:"job_name:like"`
	JobGroup  string `form:"jobGroup" query:"job_group:="`
	Status    string `form:"status" query:"status:="`
	BeginTime string `form:"beginTime" query:"date(create_time):>="`
	EndTime   string `form:"endTime" query:"date(create_time):<="`
	Sort      string `form:"sort,default=create_time:desc"`
	common.PageReq
}

type RunJobRequest struct {
	Id uint64 `json:"id" binding:"required" err_msg:"任务id 必填"`
}

type UpdateJobRequest struct {
	Id               uint64 `gorm:"column:id;primary_key;" json:"id"        description:"任务id" binding:"required" err_msg:"任务id 必填"`
	JobName          string `gorm:"column:job_name" json:"jobName"  description:"任务名称" binding:"required" err_msg:"任务名称 必填"`
	JobGroup         string `gorm:"column:job_group" json:"jobGroup"  description:"任务组名"`
	InvokeTarget     string `gorm:"column:invoke_target" json:"invokeTarget"  description:"调用目标字符串" binding:"required" err_msg:"调用方法 必填"`
	InvokeTargetArgs string `gorm:"column:invoke_target_args" json:"invokeTargetArgs"  description:"调用目标参数"`
	CronExpression   string `gorm:"column:cron_expression" json:"cronExpression"  description:"cron执行表达式" binding:"required" err_msg:"cron表达式 必填"`
	MisfirePolicy    string `gorm:"column:misfire_policy" json:"MisfirePolicy"  description:"计划执行错误策略（1立即执行 2执行一次 3放弃执行）"`
	Concurrent       string `gorm:"column:concurrent" json:"concurrent" description:"是否并发执行（0允许 1禁止）"`
	Status           string `gorm:"column:status" json:"status"    description:"状态（0正常 1暂停）"`
	Remark           string `gorm:"column:remark" json:"remark"    description:"备注信息"`
}
