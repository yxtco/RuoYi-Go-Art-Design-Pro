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

type OperLogHandler struct {
	Services service.Services
	Api      api.Api
}

func NewOperLogHandler() *OperLogHandler {
	return &OperLogHandler{}
}

// ListOperLog 查询操作日志列表
//
//	@Summary	查询操作日志列表
//	@Tags		操作日志
//	@Produce	json
//	@Param		title		query	string	false	"系统模块"
//	@Param		operName	query	string	false	"操作人员"
//	@Param		businessType	query	int	false	"操作类型"
//	@Param		status		query	int	false	"状态"
//	@Param		pageNum		query	int	false	"页码"
//	@Param		pageSize	query	int	false	"每页数量"
//	@Success	200		{object}	map[string]interface{}"操作日志列表"
//	@Router		/monitor/operLog/list [get]
//	@Security	BearerAuth
func (s *OperLogHandler) ListOperLog(c *gin.Context) {
	response.SetOperTitle(c, "查询操作日志列表")
	req := s.Api.SysOperLog.GetOperLogListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysOperLogService.SelectOperLogList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}

// DelOperLog 删除操作日志
//
//	@Summary	删除操作日志
//	@Tags		操作日志
//	@Produce	json
//	@Param		ids	path	string	true	"日志ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/operLog/{ids} [delete]
//	@Security	BearerAuth
func (s *OperLogHandler) DelOperLog(c *gin.Context) {
	response.SetOperTitle(c, "删除操作日志")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysOperLogService.DeleteOperLogByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}

// CleanOperLog 清空操作日志
//
//	@Summary	清空操作日志
//	@Tags		操作日志
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/operLog/clean [delete]
//	@Security	BearerAuth
func (s *OperLogHandler) CleanOperLog(c *gin.Context) {
	response.SetOperTitle(c, "清空操作日志")
	err := s.Services.SysOperLogService.CleanOperLog()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// Export 导出操作日志
//
//	@Summary	导出操作日志
//	@Tags		操作日志
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/monitor/operLog/export [post]
//	@Security	BearerAuth
func (s *OperLogHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出操作日志")
	req := s.Api.SysOperLog.GetOperLogListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysOperLogService.SelectOperLogAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"日志编号", "系统模块", "操作类型", "操作人员", "操作地址", "操作地点", "操作状态", "操作日期", "消耗时间"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.Id),
			obj.Title,
			fmt.Sprint(obj.BusinessType),
			obj.OperName,
			obj.OperIp,
			obj.OperLocation,
			fmt.Sprint(obj.Status),
			fmt.Sprint(obj.OperTime),
			fmt.Sprint(obj.CostTime),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "操作日志")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
