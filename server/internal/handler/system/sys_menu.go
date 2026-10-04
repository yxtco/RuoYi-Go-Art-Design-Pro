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

type MenuHandler struct {
	Services service.Services
	Api      api.Api
}

func NewMenuHandler() *MenuHandler {
	return &MenuHandler{}
}

// ListMenu 查询菜单列表
//
//	@Summary	查询菜单列表
//	@Tags		菜单管理
//	@Produce	json
//	@Param		menuName	query	string	false	"菜单名称"
//	@Param		status		query	string	false	"状态"
//	@Success	200			{object}	map[string]interface{}"菜单列表"
//	@Router		/system/menu/list [get]
//	@Security	BearerAuth
func (s *MenuHandler) ListMenu(c *gin.Context) {
	response.SetOperTitle(c, "查询菜单列表")
	req := s.Api.SysMenu.ListMenuRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	data, err := s.Services.SysMenuService.SelectMenuList(loginUser.UserId, conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)

}
// GetMenu 查询菜单详细
//
//	@Summary	查询菜单详细
//	@Tags		菜单管理
//	@Produce	json
//	@Param		id	path	int	true	"菜单ID"
//	@Success	200	{object}	map[string]interface{}"菜单信息"
//	@Router		/system/menu/{id} [get]
//	@Security	BearerAuth
func (s *MenuHandler) GetMenu(c *gin.Context) {
	response.SetOperTitle(c, "查询菜单详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysMenuService.SelectMenuById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}
// TreeSelect 查询菜单下拉树结构
//
//	@Summary	查询菜单下拉树结构
//	@Tags		菜单管理
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"菜单树"
//	@Router		/system/menu/treeselect [get]
//	@Security	BearerAuth
func (s *MenuHandler) TreeSelect(c *gin.Context) {
	response.SetOperTitle(c, "查询菜单下拉树结构")
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	data, err := s.Services.SysMenuService.SelectMenuList(loginUser.UserId, map[string]interface{}{})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	dataTree := s.Services.SysMenuService.BuildMenuTreeSelect(data)
	response.Data(c, dataTree)
}
// RoleMenuTreeSelect 根据角色ID查询菜单下拉树结构
//
//	@Summary	根据角色ID查询菜单下拉树结构
//	@Tags		菜单管理
//	@Produce	json
//	@Param		roleId	path	int	true	"角色ID"
//	@Success	200		{object}	map[string]interface{}"菜单树和已勾选的菜单ID"
//	@Router		/system/menu/roleMenuTreeselect/{roleId} [get]
//	@Security	BearerAuth
func (s *MenuHandler) RoleMenuTreeSelect(c *gin.Context) {
	response.SetOperTitle(c, "查询角色菜单下拉树结构")
	req := s.Api.SysMenu.RoleIdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	data, err := s.Services.SysMenuService.SelectMenuList(loginUser.UserId, map[string]interface{}{})
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	menus := s.Services.SysMenuService.BuildMenuTreeSelect(data)
	ids, err := s.Services.SysMenuService.SelectMenuListByRoleId(req.RoleId)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.DataExpand(c, nil, map[string]interface{}{
		"checkedKeys": ids,
		"menus":       menus,
	})

}
// AddMenu 新增菜单
//
//	@Summary	新增菜单
//	@Tags		菜单管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"菜单信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/menu [post]
//	@Security	BearerAuth
func (s *MenuHandler) AddMenu(c *gin.Context) {
	response.SetOperTitle(c, "新增菜单")
	var req model.SysMenu
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysMenuService.CheckMenuNameUnique(req.Id, req.MenuName) {
		response.Error(c, "新增菜单'"+req.MenuName+"'失败，菜单名称已存在")
		return
	} else if "0" == req.IsFrame && !stringutils.IsHttp(req.Path) {
		response.Error(c, "新增菜单'"+req.MenuName+"'失败，地址必须以http(s)://开头")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	req.CreateBy = loginUser.UserName
	err := s.Services.SysMenuService.InsertMenu(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)
}
// UpdateMenu 修改菜单
//
//	@Summary	修改菜单
//	@Tags		菜单管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"菜单信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/menu [put]
//	@Security	BearerAuth
func (s *MenuHandler) UpdateMenu(c *gin.Context) {
	response.SetOperTitle(c, "修改菜单")
	req := s.Api.SysMenu.UpdateMenuRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysMenuService.CheckMenuNameUnique(req.Id, req.MenuName) {
		response.Error(c, "修改菜单'"+req.MenuName+"'失败，菜单名称已存在")
		return
	} else if "0" == req.IsFrame && !stringutils.IsHttp(req.Path) {
		response.Error(c, "修改菜单'"+req.MenuName+"'失败，地址必须以http(s)://开头")
		return
	} else if req.Id == req.ParentId {
		response.Error(c, "修改菜单'"+req.MenuName+"'失败，上级菜单不能选择自己")
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err := s.Services.SysMenuService.UpdateMenu(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// UpdateSort 菜单排序调整
//
//	@Summary	菜单排序调整
//	@Tags		菜单管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含menuIds和orderNums"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/menu/updateSort [put]
//	@Security	BearerAuth
func (s *MenuHandler) UpdateSort(c *gin.Context) {
	response.SetOperTitle(c, "菜单排序调整")
	var req struct {
		MenuIds   string `json:"menuIds"`
		OrderNums string `json:"orderNums"`
	}
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := stringutils.StringSliceToUint64Slice(strings.Split(req.MenuIds, ","))
	orderNums := stringutils.StringSliceToUint64Slice(strings.Split(req.OrderNums, ","))
	if len(ids) != len(orderNums) {
		response.Error(c, "参数错误：menuIds 和 orderNums 长度不一致")
		return
	}
	for i, id := range ids {
		upd := map[string]interface{}{"order_num": orderNums[i]}
		err := s.Services.SysMenuService.UpdateMenu(id, upd)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
	}
	response.Data(c, nil)
}

// DelMenu 删除菜单
//
//	@Summary	删除菜单
//	@Tags		菜单管理
//	@Produce	json
//	@Param		id	path	int	true	"菜单ID"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/menu/{id} [delete]
//	@Security	BearerAuth
func (s *MenuHandler) DelMenu(c *gin.Context) {
	response.SetOperTitle(c, "删除菜单")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	has, err := s.Services.SysMenuService.HasChildByMenuId(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if has {
		response.Error(c, "存在子菜单,不允许删除")
		return
	}

	has, err = s.Services.SysMenuService.CheckMenuExistRole(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if has {
		response.Error(c, "菜单已分配,不允许删除")
		return
	}

	err = s.Services.SysMenuService.DeleteMenuById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
