package monitor

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/exceltool"
	"go-fin-server/pkg/taskregistry"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/cronutils"
	"go-fin-server/pkg/utils/stringutils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	Services service.Services
	Api      api.Api
}

func NewJobHandler() *JobHandler {
	return &JobHandler{}
}

// ListJob 查询定时任务列表
//
//	@Summary	查询定时任务列表
//	@Tags		定时任务
//	@Produce	json
//	@Param		jobName	query	string	false	"任务名称"
//	@Param		jobGroup	query	string	false	"任务组名"
//	@Param		status		query	string	false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"任务列表"
//	@Router		/monitor/job/list [get]
//	@Security	BearerAuth
func (s *JobHandler) ListJob(c *gin.Context) {
	response.SetOperTitle(c, "查询定时任务列表")
	req := s.Api.SysJob.ListJobRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysJobService.SelectJobList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}

// GetJob 查询定时任务详细
//
//	@Summary	查询定时任务详细
//	@Tags		定时任务
//	@Produce	json
//	@Param		id	path	int	true	"任务ID"
//	@Success	200	{object}	map[string]interface{}"任务信息"
//	@Router		/monitor/job/{id} [get]
//	@Security	BearerAuth
func (s *JobHandler) GetJob(c *gin.Context) {
	response.SetOperTitle(c, "查询定时任务详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysJobService.SelectJobById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	nextValidTime, err := cronutils.GetNextExecution(data.CronExpression)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data.NextValidTime = nextValidTime
	response.Data(c, data)
}

// AddJob 新增定时任务
//
//	@Summary	新增定时任务
//	@Tags		定时任务
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"任务信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/job [post]
//	@Security	BearerAuth
func (s *JobHandler) AddJob(c *gin.Context) {
	response.SetOperTitle(c, "新增定时任务")
	var req model.SysJob
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysJobService.CheckJobNameUnique(req.Id, req.JobName) {
		response.Error(c, "新增任务'"+req.JobName+"'失败，任务名称已存在")
		return
	} else if !cronutils.IsValid(req.CronExpression) {
		response.Error(c, "新增任务'"+req.JobName+"'失败，Cron表达式不正确")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysJobService.InsertJob(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// UpdateJob 修改定时任务
//
//	@Summary	修改定时任务
//	@Tags		定时任务
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"任务信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/job [put]
//	@Security	BearerAuth
func (s *JobHandler) UpdateJob(c *gin.Context) {
	response.SetOperTitle(c, "修改定时任务")
	req := s.Api.SysJob.UpdateJobRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysJobService.CheckJobNameUnique(req.Id, req.JobName) {
		response.Error(c, "修改任务'"+req.JobName+"'失败，任务名称已存在")
		return
	} else if !cronutils.IsValid(req.CronExpression) {
		response.Error(c, "修改任务'"+req.JobName+"'失败，Cron表达式不正确")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysJobService.UpdateJob(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// DelJob 删除定时任务
//
//	@Summary	删除定时任务
//	@Tags		定时任务
//	@Produce	json
//	@Param		ids	path	string	true	"任务ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/job/{ids} [delete]
//	@Security	BearerAuth
func (s *JobHandler) DelJob(c *gin.Context) {
	response.SetOperTitle(c, "删除定时任务")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysJobService.DeleteJobByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}
	response.Data(c, nil)
}
// ChangeJobStatus 修改任务状态
//
//	@Summary	修改任务状态
//	@Tags		定时任务
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含id和status"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/job/changeStatus [put]
//	@Security	BearerAuth
func (s *JobHandler) ChangeJobStatus(c *gin.Context) {
	response.SetOperTitle(c, "修改定时任务状态")
	req := s.Api.Common.ChangeStatusRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := map[string]interface{}{
		"status":    req.Status,
		"update_by": loginUser.UserName,
	}
	err := s.Services.SysJobService.UpdateJobStatus(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// RunJob 立即执行一次任务
//
//	@Summary	立即执行一次任务
//	@Tags		定时任务
//	@Router		/monitor/job/run [put]
//	@Security	BearerAuth
func (s *JobHandler) RunJob(c *gin.Context) {
	response.SetOperTitle(c, "立即执行一次定时任务")
	req := s.Api.SysJob.RunJobRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 从数据库获取任务信息
	job, err := s.Services.SysJobService.SelectJobById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 记录调度日志
	jobLog := model.SysJobLog{
		JobName:          job.JobName,
		JobGroup:         job.JobGroup,
		InvokeTarget:     job.InvokeTarget,
		InvokeTargetArgs: job.InvokeTargetArgs,
		JobUpdatedTime:   job.UpdatedAt,
	}
	err = jobLog.Create(&jobLog)
	if err != nil {
		pkg.Logger.Error(err)
	}
	// 执行任务
	startTime := time.Now()
	err = taskregistry.Execute(job.InvokeTarget, job.InvokeTargetArgs)
	duration := time.Since(startTime).Milliseconds()
	if err != nil {
		// 记录失败日志
		_ = jobLog.UpdateMap(map[string]interface{}{
			"exception_info": err.Error(),
			"status":         "1",
			"job_message":    job.JobName + " 总共耗时：" + strconv.FormatInt(duration, 10) + "毫秒",
		}, "id = ?", jobLog.Id)
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 记录成功日志
	_ = jobLog.UpdateMap(map[string]interface{}{
		"job_message": job.JobName + " 总共耗时：" + strconv.FormatInt(duration, 10) + "毫秒",
	}, "id = ?", jobLog.Id)
	response.Data(c, nil)
}
// Export 导出定时任务数据
//
//	@Summary	导出定时任务数据
//	@Tags		定时任务
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/monitor/job/export [post]
//	@Security	BearerAuth
func (s *JobHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出定时任务数据")
	req := s.Api.SysJob.ListJobRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysJobService.SelectJobAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"任务编号", "任务名称", "任务组名", "调用目标字符串", "cron执行表达式", "状态"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.Id),
			obj.JobName,
			obj.JobGroup,
			obj.InvokeTarget,
			obj.CronExpression,
			obj.Status,
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "定时任务")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
