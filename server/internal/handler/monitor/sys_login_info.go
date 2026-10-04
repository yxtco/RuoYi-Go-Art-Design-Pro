package monitor

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/db"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/exceltool"
	"go-fin-server/pkg/redistool"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"strings"

	"github.com/gin-gonic/gin"
)

type LoginInfoHandler struct {
	Services service.Services
	Api      api.Api
}

func NewLoginInfoHandler() *LoginInfoHandler {
	return &LoginInfoHandler{}
}

// ListLoginInfo 查询登录日志列表
//
//	@Summary	查询登录日志列表
//	@Tags		登录日志
//	@Produce	json
//	@Param		ipaddr		query	string	false	"IP地址"
//	@Param		userName	query	string	false	"用户名称"
//	@Param		status		query	int		false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"登录日志列表"
//	@Router		/monitor/logininfor/list [get]
//	@Security	BearerAuth
func (s *LoginInfoHandler) ListLoginInfo(c *gin.Context) {
	response.SetOperTitle(c, "查询登录日志列表")
	req := s.Api.SysLoginInfo.GetLoginInfoListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysLoginInfoService.SelectLoginInfoList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}

// DelLoginInfo 删除登录日志
//
//	@Summary	删除登录日志
//	@Tags		登录日志
//	@Produce	json
//	@Param		ids	path	string	true	"日志ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/logininfor/{ids} [delete]
//	@Security	BearerAuth
func (s *LoginInfoHandler) DelLoginInfo(c *gin.Context) {
	response.SetOperTitle(c, "删除登录日志")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysLoginInfoService.DeleteLoginInfoByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}

// UnlockLoginInfo 解锁用户登录状态
//
//	@Summary	解锁用户登录状态
//	@Tags		登录日志
//	@Produce	json
//	@Param		userName	path	string	true	"用户名"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/logininfor/unlock/{userName} [get]
//	@Security	BearerAuth
func (s *LoginInfoHandler) UnlockLoginInfo(c *gin.Context) {
	response.SetOperTitle(c, "账户解锁")
	req := s.Api.SysLoginInfo.UserNameUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	exists, err := redistool.Exists(db.RedisConnections["master"], constant.CACHE_PWD_ERR_CNT_KEY+req.UserName)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if exists {
		err := redistool.Del(db.RedisConnections["master"], constant.CACHE_PWD_ERR_CNT_KEY+req.UserName)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
	}
	response.Data(c, nil)
}

// CleanLoginInfo 清空登录日志
//
//	@Summary	清空登录日志
//	@Tags		登录日志
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/logininfor/clean [delete]
//	@Security	BearerAuth
func (s *LoginInfoHandler) CleanLoginInfo(c *gin.Context) {
	response.SetOperTitle(c, "清空登录日志")
	err := s.Services.SysLoginInfoService.CleanLoginInfo()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// Export 导出登录日志
//
//	@Summary	导出登录日志
//	@Tags		登录日志
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/monitor/logininfor/export [post]
//	@Security	BearerAuth
func (s *LoginInfoHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出登录日志")
	req := s.Api.SysLoginInfo.GetLoginInfoListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysLoginInfoService.SelectLoginInfoAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"访问编号", "用户名称", "登录地址", "登录地点", "浏览器", "操作系统", "登录状态", "操作信息", "登录日期"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.Id),
			obj.UserName,
			obj.Ipaddr,
			obj.LoginLocation,
			obj.Browser,
			obj.Os,
			obj.Status,
			obj.Msg,
			fmt.Sprint(obj.LoginTime),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "登录日志")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
