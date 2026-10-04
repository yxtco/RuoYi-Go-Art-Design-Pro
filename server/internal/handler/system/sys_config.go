package system

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/exceltool"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"strings"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	Services service.Services
	Api      api.Api
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// ListConfig 查询参数列表
//
//	@Summary	查询参数列表
//	@Tags		参数管理
//	@Produce	json
//	@Param		configName	query	string	false	"参数名称"
//	@Param		configKey		query	string	false	"参数键名"
//	@Param		configType	query	string	false	"系统内置"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"参数列表"
//	@Router		/system/config/list [get]
//	@Security	BearerAuth
func (s *ConfigHandler) ListConfig(c *gin.Context) {
	response.SetOperTitle(c, "查询参数列表")
	req := s.Api.SysConfig.ListConfigRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysConfigService.SelectConfigList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}
// GetConfig 查询参数详细
//
//	@Summary	查询参数详细
//	@Tags		参数管理
//	@Produce	json
//	@Param		id	path	int	true	"参数ID"
//	@Success	200	{object}	map[string]interface{}"参数信息"
//	@Router		/system/config/{id} [get]
//	@Security	BearerAuth
func (s *ConfigHandler) GetConfig(c *gin.Context) {
	response.SetOperTitle(c, "查询参数详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysConfigService.SelectConfigById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)

}
// GetConfigKey 根据参数键名查询参数值
//
//	@Summary	根据参数键名查询参数值
//	@Tags		参数管理
//	@Produce	json
//	@Param		configKey	path	string	true	"参数键名"
//	@Success	200		{object}	map[string]interface{}"参数值"
//	@Router		/system/config/configKey/{configKey} [get]
func (s *ConfigHandler) GetConfigKey(c *gin.Context) {
	response.SetOperTitle(c, "根据参数键名查询参数值")
	req := s.Api.SysConfig.ConfigKeyUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysConfigService.SelectConfigByKey(req.ConfigKey)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
// AddConfig 新增参数配置
//
//	@Summary	新增参数配置
//	@Tags		参数管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"参数信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/config [post]
//	@Security	BearerAuth
func (s *ConfigHandler) AddConfig(c *gin.Context) {
	response.SetOperTitle(c, "新增参数配置")
	var req model.SysConfig
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysConfigService.CheckConfigKeyUnique(req.Id, req.ConfigKey) {
		response.Error(c, "新增参数'"+req.ConfigKey+"'失败，参数键名已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysConfigService.InsertConfig(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// UpdateConfig 修改参数配置
//
//	@Summary	修改参数配置
//	@Tags		参数管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"参数信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/config [put]
//	@Security	BearerAuth
func (s *ConfigHandler) UpdateConfig(c *gin.Context) {
	response.SetOperTitle(c, "修改参数配置")
	req := s.Api.SysConfig.UpdateConfigRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysConfigService.CheckConfigKeyUnique(req.Id, req.ConfigKey) {
		response.Error(c, "修改参数'"+req.ConfigKey+"'失败，参数键名已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)

	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysConfigService.UpdateConfig(req.Id, req.ConfigKey, req.ConfigValue, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)
}
// DelConfig 删除参数配置
//
//	@Summary	删除参数配置
//	@Tags		参数管理
//	@Produce	json
//	@Param		ids	path	string	true	"参数ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/config/{ids} [delete]
//	@Security	BearerAuth
func (s *ConfigHandler) DelConfig(c *gin.Context) {
	response.SetOperTitle(c, "删除参数配置")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysConfigService.DeleteConfigByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}
// RefreshCache 刷新参数缓存
//
//	@Summary	刷新参数缓存
//	@Tags		参数管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/config/refreshCache [delete]
//	@Security	BearerAuth
func (s *ConfigHandler) RefreshCache(c *gin.Context) {
	response.SetOperTitle(c, "刷新参数缓存")
	err := s.Services.SysConfigService.ResetConfigCache()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// Export 导出参数数据
//
//	@Summary	导出参数数据
//	@Tags		参数管理
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/system/config/export [post]
//	@Security	BearerAuth
func (s *ConfigHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出参数数据")
	req := s.Api.SysConfig.ListConfigRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysConfigService.SelectConfigAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"参数主键", "参数名称", "参数键名", "参数键值", "系统内置", "备注", "创建时间"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.Id),
			obj.ConfigName,
			obj.ConfigKey,
			obj.ConfigValue,
			obj.ConfigType,
			obj.Remark,
			fmt.Sprint(obj.CreatedAt),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "参数数据")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
