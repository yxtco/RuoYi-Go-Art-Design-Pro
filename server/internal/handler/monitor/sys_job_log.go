package monitor

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/exceltool"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"strings"

	"github.com/gin-gonic/gin"
)

type JobLogHandler struct {
	Services service.Services
	Api      api.Api
}

func NewJobLogHandler() *JobLogHandler {
	return &JobLogHandler{}
}

// ListJobLog 查询调度日志列表
//
//	@Summary	查询调度日志列表
//	@Tags		调度日志
//	@Produce	json
//	@Param		jobName	query	string	false	"任务名称"
//	@Param		jobGroup	query	string	false	"任务组名"
//	@Param		status		query	string	false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"调度日志列表"
//	@Router		/monitor/jobLog/list [get]
//	@Security	BearerAuth
func (s *JobLogHandler) ListJobLog(c *gin.Context) {
	response.SetOperTitle(c, "查询调度日志列表")
	req := s.Api.SysJobLog.ListJobLogRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysJobLogService.SelectJobLogList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}

// DelJobLog 删除调度日志
//
//	@Summary	删除调度日志
//	@Tags		调度日志
//	@Produce	json
//	@Param		ids	path	string	true	"日志ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/jobLog/{ids} [delete]
//	@Security	BearerAuth
func (s *JobLogHandler) DelJobLog(c *gin.Context) {
	response.SetOperTitle(c, "删除调度日志")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysJobLogService.DeleteJobLogByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}
	response.Data(c, nil)
}

// CleanJobLog 清空调度日志
//
//	@Summary	清空调度日志
//	@Tags		调度日志
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/jobLog/clean [delete]
//	@Security	BearerAuth
func (s *JobLogHandler) CleanJobLog(c *gin.Context) {
	response.SetOperTitle(c, "清空调度日志")
	err := s.Services.SysJobLogService.CleanJobLog()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "清除失败服务错误")
		return
	}
	response.Data(c, nil)
}

// GetJobLog 查询调度日志详细
//
//	@Summary	查询调度日志详细
//	@Tags		调度日志
//	@Produce	json
//	@Param		id	path	int	true	"日志ID"
//	@Success	200	{object}	map[string]interface{}"日志信息"
//	@Router		/monitor/jobLog/{id} [get]
//	@Security	BearerAuth
func (s *JobLogHandler) GetJobLog(c *gin.Context) {
	response.SetOperTitle(c, "查询调度日志详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysJobLogService.SelectJobLogById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}

// Export 导出调度日志
//
//	@Summary	导出调度日志
//	@Tags		调度日志
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/monitor/jobLog/export [post]
//	@Security	BearerAuth
func (s *JobLogHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出调度日志")
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
	headers := []string{"日志编号", "任务名称", "任务组名", "调用方法", "调用方法参数", "日志信息", "执行状态", "执行时间", "状态"}
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
