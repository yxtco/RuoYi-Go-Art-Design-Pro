package system

import (
	"go-fin-server/api"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"strings"

	"github.com/gin-gonic/gin"
)

type DeptHandler struct {
	Services service.Services
	Api      api.Api
}

func NewDeptHandler() *DeptHandler {
	return &DeptHandler{}
}

// ListDept 查询部门列表
//
//	@Summary	查询部门列表
//	@Tags		部门管理
//	@Produce	json
//	@Param		deptName	query	string	false	"部门名称"
//	@Param		status		query	string	false	"状态"
//	@Success	200			{object}	map[string]interface{}"部门列表"
//	@Router		/system/dept/list [get]
//	@Security	BearerAuth
func (s *DeptHandler) ListDept(c *gin.Context) {
	response.SetOperTitle(c, "查询部门列表")
	req := s.Api.SysDept.ListDeptRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	sortConditions := utils.BuildSortConditions(req.Sort)
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	data, err := s.Services.SysDeptService.SelectDataScopeDeptList(conditions, sortConditions, loginUser.User)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
// ListDeptExcludeChild 查询部门列表（排除子节点）
//
//	@Summary	查询部门列表（排除子节点）
//	@Tags		部门管理
//	@Produce	json
//	@Param		id	path	int	true	"部门ID"
//	@Success	200	{object}	map[string]interface{}"部门列表"
//	@Router		/system/dept/list/exclude/{id} [get]
//	@Security	BearerAuth
func (s *DeptHandler) ListDeptExcludeChild(c *gin.Context) {
	response.SetOperTitle(c, "查询部门列表（排除子节点）")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	depts, err := s.Services.SysDeptService.SelectDeptList(map[string]interface{}{}, []string{"parent_id", "order_num"})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	newDepts := make([]model.SysDept, 0)
	for _, dept := range depts {
		if !(dept.Id == req.Id || utils.SliceUint64ContainsElement(stringutils.StringSliceToUint64Slice(strings.Split(dept.Ancestors, ",")), req.Id)) {
			newDepts = append(newDepts, dept)
		}
	}
	response.Data(c, newDepts)
}

// GetDept 查询部门详细
//
//	@Summary	查询部门详细
//	@Tags		部门管理
//	@Produce	json
//	@Param		id	path	int	true	"部门ID"
//	@Success	200	{object}	map[string]interface{}"部门信息"
//	@Router		/system/dept/{id} [get]
//	@Security	BearerAuth
func (s *DeptHandler) GetDept(c *gin.Context) {
	response.SetOperTitle(c, "查询部门详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysDeptService.SelectDeptById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)

}
// AddDept 新增部门
//
//	@Summary	新增部门
//	@Tags		部门管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"部门信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dept [post]
//	@Security	BearerAuth
func (s *DeptHandler) AddDept(c *gin.Context) {
	response.SetOperTitle(c, "新增部门")
	var req model.SysDept
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysDeptService.CheckDeptNameUnique(req.Id, req.DeptName) {
		response.Error(c, "新增部门'"+req.DeptName+"'失败，部门名称已存在")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysDeptService.InsertDept(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// UpdateDept 修改部门
//
//	@Summary	修改部门
//	@Tags		部门管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"部门信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dept [put]
//	@Security	BearerAuth
func (s *DeptHandler) UpdateDept(c *gin.Context) {
	response.SetOperTitle(c, "修改部门")
	req := s.Api.SysDept.UpdateDeptRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err := s.Services.SysDeptService.CheckDeptDataScope(loginUser.UserId, req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysDeptService.CheckDeptNameUnique(req.Id, req.DeptName) {
		response.Error(c, "修改部门'"+req.DeptName+"'失败，部门名称已存在")
		return
	} else if *req.ParentId == req.Id {
		response.Error(c, "修改部门'"+req.DeptName+"'失败，上级部门不能是自己")
		return
	} else if "0" == req.Status {
		count, err := s.Services.SysDeptService.SelectNormalChildrenDeptById(req.Id)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		if count > 0 {
			response.Error(c, "该部门包含未停用的子部门！")
			return
		}
	}
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err = s.Services.SysDeptService.UpdateDept(req.Id, *req.ParentId, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// DelDept 删除部门
//
//	@Summary	删除部门
//	@Tags		部门管理
//	@Produce	json
//	@Param		id	path	int	true	"部门ID"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/dept/{id} [delete]
//	@Security	BearerAuth
func (s *DeptHandler) DelDept(c *gin.Context) {
	response.SetOperTitle(c, "删除部门")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	has, err := s.Services.SysDeptService.HasChildByDeptId(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if has {
		response.Error(c, "存在下级部门,不允许删除")
		return
	}

	has, err = s.Services.SysDeptService.CheckDeptExistUser(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if has {
		response.Error(c, "部门存在用户,不允许删除")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err = s.Services.SysDeptService.CheckDeptDataScope(loginUser.UserId, req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	err = s.Services.SysDeptService.DeleteDeptById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

// UpdateSort 部门排序调整
//
//	@Summary	部门排序调整
//	@Tags		部门管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含deptIds和orderNums"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/dept/updateSort [put]
//	@Security	BearerAuth
func (s *DeptHandler) UpdateSort(c *gin.Context) {
	response.SetOperTitle(c, "修改部门排序")
	var req struct {
		DeptIds   string `json:"deptIds"`
		OrderNums string `json:"orderNums"`
	}
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := stringutils.StringSliceToUint64Slice(strings.Split(req.DeptIds, ","))
	orderNums := stringutils.StringSliceToUint64Slice(strings.Split(req.OrderNums, ","))
	if len(ids) != len(orderNums) {
		response.Error(c, "参数错误：deptIds 和 orderNums 长度不一致")
		return
	}
	for i, id := range ids {
		upd := map[string]interface{}{"order_num": orderNums[i]}
		err := model.SysDept{}.UpdateMap(upd, "id = ?", id)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
	}
	response.Data(c, nil)
}
