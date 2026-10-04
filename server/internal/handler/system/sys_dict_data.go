package system

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/constant"
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

type DictDataHandler struct {
	Services service.Services
	Api      api.Api
}

func NewDictDataHandler() *DictDataHandler {
	return &DictDataHandler{}
}

// ListDictData 查询字典数据列表
//
//	@Summary	查询字典数据列表
//	@Tags		字典数据
//	@Produce	json
//	@Param		dictType	query	string	false	"字典类型"
//	@Param		dictLabel	query	string	false	"字典标签"
//	@Param		status		query	string	false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"字典数据列表"
//	@Router		/system/dict/data/list [get]
//	@Security	BearerAuth
func (s *DictDataHandler) ListDictData(c *gin.Context) {
	response.SetOperTitle(c, "查询字典数据列表")
	req := s.Api.SysDictData.ListDictDataRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	data, total, err := s.Services.SysDictDataService.SelectDictDataList(conditions, sortConditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}
// GetDictData 查询字典数据详细
//
//	@Summary	查询字典数据详细
//	@Tags		字典数据
//	@Produce	json
//	@Param		id	path	int	true	"字典数据ID"
//	@Success	200	{object}	map[string]interface{}"字典数据信息"
//	@Router		/system/dict/data/{id} [get]
//	@Security	BearerAuth
func (s *DictDataHandler) GetDictData(c *gin.Context) {
	response.SetOperTitle(c, "查询字典数据详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysDictDataService.SelectDictDataById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)

}
func (s *DictDataHandler) GetDictDataByType(c *gin.Context) {
	response.SetOperTitle(c, "根据字典类型查询字典数据")
	dictType := c.Param("dictType")
	if dictType == "" {
		response.Error(c, "字典类型不能为空")
		return
	}
	cacheKey := constant.CACHE_SYS_DICT_KEY + dictType
	// 优先从 Redis 读取字典缓存
	cacheData, err := s.Services.SysDictTypeService.GetDictCache(cacheKey)
	if err == nil && len(cacheData) > 0 {
		response.Data(c, cacheData)
		return
	}
	// 缓存未命中，从 MySQL 查询
	data, err := model.SysDictData{}.GetAll("status = ? and dict_type = ?", "0", dictType)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 回填 Redis 缓存
	if len(data) > 0 {
		_ = s.Services.SysDictTypeService.SetDictCache(cacheKey, data)
	}
	response.Data(c, data)
}

// GetDicts 根据字典类型查询字典数据信息
//
//	@Summary	根据字典类型查询字典数据信息
//	@Tags		字典数据
//	@Accept		json
//	@Produce	json
//	@Param		body	body		[]string	true	"字典类型列表"
//	@Success	200		{object}	map[string]interface{}"字典数据"
//	@Router		/system/dict/data/types [post]
func (s *DictDataHandler) GetDicts(c *gin.Context) {
	response.SetOperTitle(c, "根据字典类型列表查询字典数据")
	var types []string
	if err := api.ShouldBindError(c, &types); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data := make(map[string][]model.SysDictDataSimplify)
	for _, t := range types {
		cacheKey := constant.CACHE_SYS_DICT_KEY + t
		// 优先从 Redis 读取字典缓存
		cacheData, err := s.Services.SysDictTypeService.GetDictCache(cacheKey)
		if err == nil && len(cacheData) > 0 {
			// 缓存命中，转换为简化格式
			simplifyList := make([]model.SysDictDataSimplify, 0, len(cacheData))
			for _, d := range cacheData {
				simplifyList = append(simplifyList, model.SysDictDataSimplify{
					DictLabel: d.DictLabel,
					DictValue: d.DictValue,
					ListClass: d.ListClass,
					CssClass:  d.CssClass,
				})
			}
			data[t] = simplifyList
			continue
		}
		// 缓存未命中，兜底从 MySQL 查询
		var dataDict model.SysDictDataSimplify
		d, err := dataDict.GetSimplify("status = ? and dict_type = ?", "0", t)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		data[t] = d
		// 回填 Redis 缓存
		mysqlData, mysqlErr := model.SysDictData{}.GetAll("status = ? and dict_type = ?", "0", t)
		if mysqlErr == nil && len(mysqlData) > 0 {
			_ = s.Services.SysDictTypeService.SetDictCache(cacheKey, mysqlData)
		}
	}
	response.Data(c, data)
}
// AddDictData 新增字典数据
//
//	@Summary	新增字典数据
//	@Tags		字典数据
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"字典数据信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/data [post]
//	@Security	BearerAuth
func (s *DictDataHandler) AddDictData(c *gin.Context) {
	response.SetOperTitle(c, "新增字典数据")
	var req model.SysDictData
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysDictDataService.CheckDictValueUnique(req.DictCode, req.DictValue) {
		response.Error(c, "新增字典数据'"+req.DictType+"'失败，数据键值已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysDictDataService.InsertDictData(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)
}
// UpdateDictData 修改字典数据
//
//	@Summary	修改字典数据
//	@Tags		字典数据
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"字典数据信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/data [put]
//	@Security	BearerAuth
func (s *DictDataHandler) UpdateDictData(c *gin.Context) {
	response.SetOperTitle(c, "修改字典数据")
	req := s.Api.SysDictData.UpdateDictDataRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysDictDataService.CheckDictValueUnique(req.DictCode, req.DictValue) {
		response.Error(c, "修改字典数据'"+req.DictType+"'失败，数据键值已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysDictDataService.UpdateDictData(req.DictCode, req.DictType, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// DelDictData 删除字典数据
//
//	@Summary	删除字典数据
//	@Tags		字典数据
//	@Produce	json
//	@Param		ids	path	string	true	"字典数据ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/dict/data/{ids} [delete]
//	@Security	BearerAuth
func (s *DictDataHandler) DelDictData(c *gin.Context) {
	response.SetOperTitle(c, "删除字典数据")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysDictDataService.DeleteDictDataByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}

// Export 导出字典数据
//
//	@Summary	导出字典数据
//	@Tags		字典数据
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/system/dict/data/export [post]
//	@Security	BearerAuth
func (s *DictDataHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出字典数据")
	req := s.Api.SysDictData.ListDictDataRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysDictDataService.SelectDictDataAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"字典编码", "字典标签", "字典键值", "字典排序", "状态", "备注", "创建时间"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.DictCode),
			obj.DictLabel,
			obj.DictValue,
			fmt.Sprint(obj.DictSort),
			obj.Status,
			obj.Remark,
			fmt.Sprint(obj.CreatedAt),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "字典数据")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)

}
