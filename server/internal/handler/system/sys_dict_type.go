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

type DictTypeHandler struct {
	Services service.Services
	Api      api.Api
}

func NewDictTypeHandler() *DictTypeHandler {
	return &DictTypeHandler{}
}

// ListDictType 查询字典类型列表
//
//	@Summary	查询字典类型列表
//	@Tags		字典类型
//	@Produce	json
//	@Param		dictName	query	string	false	"字典名称"
//	@Param		dictType	query	string	false	"字典类型"
//	@Param		status		query	string	false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"字典类型列表"
//	@Router		/system/dict/type/list [get]
//	@Security	BearerAuth
func (s *DictTypeHandler) ListDictType(c *gin.Context) {
	response.SetOperTitle(c, "查询字典类型列表")
	req := s.Api.SysDictType.ListDictTypeRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysDictTypeService.SelectDictTypeList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)

}
// GetDictType 查询字典类型详细
//
//	@Summary	查询字典类型详细
//	@Tags		字典类型
//	@Produce	json
//	@Param		id	path	int	true	"字典类型ID"
//	@Success	200	{object}	map[string]interface{}"字典类型信息"
//	@Router		/system/dict/type/{id} [get]
//	@Security	BearerAuth
func (s *DictTypeHandler) GetDictType(c *gin.Context) {
	response.SetOperTitle(c, "查询字典类型详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysDictTypeService.SelectDictTypeById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)

}
// AddDictType 新增字典类型
//
//	@Summary	新增字典类型
//	@Tags		字典类型
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"字典类型信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/type [post]
//	@Security	BearerAuth
func (s *DictTypeHandler) AddDictType(c *gin.Context) {
	response.SetOperTitle(c, "新增字典类型")
	var req model.SysDictType
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysDictTypeService.CheckDictTypeUnique(req.Id, req.DictType) {
		response.Error(c, "新增字典'"+req.DictType+"'失败，字典类型已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysDictTypeService.InsertDictType(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// UpdateDictType 修改字典类型
//
//	@Summary	修改字典类型
//	@Tags		字典类型
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"字典类型信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/type [put]
//	@Security	BearerAuth
func (s *DictTypeHandler) UpdateDictType(c *gin.Context) {
	response.SetOperTitle(c, "修改字典类型")
	req := s.Api.SysDictType.UpdateDictTypeRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysDictTypeService.CheckDictTypeUnique(req.Id, req.DictType) {
		response.Error(c, "修改字典'"+req.DictType+"'失败，字典类型已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysDictTypeService.UpdateDictType(req.Id, req.DictType, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// DelDictType 删除字典类型
//
//	@Summary	删除字典类型
//	@Tags		字典类型
//	@Produce	json
//	@Param		ids	path	string	true	"字典类型ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/type/{ids} [delete]
//	@Security	BearerAuth
func (s *DictTypeHandler) DelDictType(c *gin.Context) {
	response.SetOperTitle(c, "删除字典类型")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysDictTypeService.DeleteDictTypeByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}
// RefreshCache 刷新字典缓存
//
//	@Summary	刷新字典缓存
//	@Tags		字典类型
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/type/refreshCache [delete]
//	@Security	BearerAuth
func (s *DictTypeHandler) RefreshCache(c *gin.Context) {
	response.SetOperTitle(c, "刷新字典缓存")
	err := s.Services.SysDictTypeService.ResetDictCache()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// OptionSelect 获取字典选择框列表
//
//	@Summary	获取字典选择框列表
//	@Tags		字典类型
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"字典类型列表"
//	@Router		/system/dict/type/optionSelect [get]
//	@Security	BearerAuth
func (s *DictTypeHandler) OptionSelect(c *gin.Context) {
	response.SetOperTitle(c, "获取字典选择框列表")
	data, err := s.Services.SysDictTypeService.SelectDictTypeAll()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
// Export 导出字典类型数据
//
//	@Summary	导出字典类型数据
//	@Tags		字典类型
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/system/dict/type/export [post]
//	@Security	BearerAuth
func (s *DictTypeHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出字典类型数据")
	req := s.Api.SysDictType.ListDictTypeRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysDictTypeService.SelectDictTypeAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"字典编号", "字典名称", "字典类型", "状态", "备注", "创建时间"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.Id),
			obj.DictName,
			obj.DictType,
			obj.Status,
			obj.Remark,
			fmt.Sprint(obj.CreatedAt),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "字典类型")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
