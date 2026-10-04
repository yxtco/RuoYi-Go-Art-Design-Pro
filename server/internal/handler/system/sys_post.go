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

type PostHandler struct {
	Services service.Services
	Api      api.Api
}

func NewPostHandler() *PostHandler {
	return &PostHandler{}
}

// ListPost 查询岗位列表
//
//	@Summary	查询岗位列表
//	@Tags		岗位管理
//	@Produce	json
//	@Param		postCode	query	string	false	"岗位编码"
//	@Param		postName	query	string	false	"岗位名称"
//	@Param		status		query	string	false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"岗位列表"
//	@Router		/system/post/list [get]
//	@Security	BearerAuth
func (s *PostHandler) ListPost(c *gin.Context) {
	response.SetOperTitle(c, "查询岗位列表")
	req := s.Api.SysPost.ListPostRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	data, total, err := s.Services.SysPostService.SelectPostList(conditions, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)

}
// GetPost 查询岗位详细
//
//	@Summary	查询岗位详细
//	@Tags		岗位管理
//	@Produce	json
//	@Param		id	path	int	true	"岗位ID"
//	@Success	200	{object}	map[string]interface{}"岗位信息"
//	@Router		/system/post/{id} [get]
//	@Security	BearerAuth
func (s *PostHandler) GetPost(c *gin.Context) {
	response.SetOperTitle(c, "查询岗位详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysPostService.SelectPostById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
// AddPost 新增岗位
//
//	@Summary	新增岗位
//	@Tags		岗位管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"岗位信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/post [post]
//	@Security	BearerAuth
func (s *PostHandler) AddPost(c *gin.Context) {
	response.SetOperTitle(c, "新增岗位")
	var req model.SysPost
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysPostService.CheckPostNameUnique(req.Id, req.PostName) {
		response.Error(c, "新增岗位'"+req.PostName+"'失败，岗位名称已存在")
		return
	} else if !s.Services.SysPostService.CheckPostCodeUnique(req.Id, req.PostCode) {
		response.Error(c, "新增岗位'"+req.PostName+"'失败，岗位编码已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysPostService.InsertPost(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// UpdatePost 修改岗位
//
//	@Summary	修改岗位
//	@Tags		岗位管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"岗位信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/post [put]
//	@Security	BearerAuth
func (s *PostHandler) UpdatePost(c *gin.Context) {
	response.SetOperTitle(c, "修改岗位")
	req := s.Api.SysPost.UpdatePostRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysPostService.CheckPostNameUnique(req.Id, req.PostName) {
		response.Error(c, "修改岗位'"+req.PostName+"'失败，岗位名称已存在")
		return
	} else if !s.Services.SysPostService.CheckPostCodeUnique(req.Id, req.PostCode) {
		response.Error(c, "修改岗位'"+req.PostName+"'失败，岗位编码已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysPostService.UpdatePost(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// DelPost 删除岗位
//
//	@Summary	删除岗位
//	@Tags		岗位管理
//	@Produce	json
//	@Param		ids	path	string	true	"岗位ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/post/{ids} [delete]
//	@Security	BearerAuth
func (s *PostHandler) DelPost(c *gin.Context) {
	response.SetOperTitle(c, "删除岗位")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	err := s.Services.SysPostService.DeletePostByIds(stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)
}
// Export 导出岗位数据
//
//	@Summary	导出岗位数据
//	@Tags		岗位管理
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/system/post/export [post]
//	@Security	BearerAuth
func (s *PostHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出岗位数据")
	req := s.Api.SysPost.ListPostRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	dataList, err := s.Services.SysPostService.SelectPostAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"岗位编号", "岗位编码", "岗位名称", "岗位排序", "状态", "创建时间"}
	var data [][]string
	for _, obj := range dataList {
		data = append(data, []string{
			fmt.Sprint(obj.Id),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "岗位数据")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
