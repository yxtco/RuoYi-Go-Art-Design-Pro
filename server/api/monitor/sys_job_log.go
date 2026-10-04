package monitor

import "go-fin-server/api/common"

type SysJobLog struct {
	ListJobLogRequest
}

type ListJobLogRequest struct {
	JobName   string `form:"jobName" query:"job_name:like"`
	JobGroup  string `form:"jobGroup" query:"job_group:="`
	Status    string `form:"status" query:"status:="`
	BeginTime string `form:"beginTime" query:"date(create_time):>="`
	EndTime   string `form:"endTime" query:"date(create_time):<="`
	Sort      string `form:"sort,default=create_time:desc"`
	common.PageReq
}
