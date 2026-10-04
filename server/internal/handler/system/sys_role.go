package system

import (
	"fmt"
	"go-fin-server/api"
	"go-fin-server/internal/config"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/exceltool"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	Services service.Services
	Api      api.Api
}

func NewRoleHandler() *RoleHandler {
	return &RoleHandler{}
}

// ListRole 查询角色列表
//
//	@Summary	查询角色列表
//	@Tags		角色管理
//	@Produce	json
//	@Param		roleName	query	string	false	"角色名称"
//	@Param		roleKey		query	string	false	"权限字符"
//	@Param		status		query	string	false	"状态"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}"角色列表"
//	@Router		/system/role/list [get]
//	@Security	BearerAuth
func (s *RoleHandler) ListRole(c *gin.Context) {
	response.SetOperTitle(c, "查询角色列表")
	req := s.Api.SysRole.GetRoleListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	data, total, err := s.Services.SysRoleService.SelectDataScopeRoleList(conditions, req.PageNum, req.PageSize, loginUser.User)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}
// GetRole 查询角色详细
//
//	@Summary	查询角色详细
//	@Tags		角色管理
//	@Produce	json
//	@Param		id	path	int	true	"角色ID"
//	@Success	200	{object}	map[string]interface{}"角色信息"
//	@Router		/system/role/{id} [get]
//	@Security	BearerAuth
func (s *RoleHandler) GetRole(c *gin.Context) {
	response.SetOperTitle(c, "查询角色详细信息")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)

	err := s.Services.SysRoleService.CheckRoleDataScope(loginUser.UserId, req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, err := s.Services.SysRoleService.SelectRoleById(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)

}
// AddRole 新增角色
//
//	@Summary	新增角色
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"角色信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role [post]
//	@Security	BearerAuth
func (s *RoleHandler) AddRole(c *gin.Context) {
	response.SetOperTitle(c, "新增角色")
	var req model.SysRoleAll
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	if !s.Services.SysRoleService.CheckRoleNameUnique(req.Id, req.RoleName) {
		response.Error(c, "新增角色'"+req.RoleName+"'失败，角色名称已存在")
		return
	} else if !s.Services.SysRoleService.CheckRoleKeyUnique(req.Id, req.RoleKey) {
		response.Error(c, "新增角色'"+req.RoleKey+"'失败，角色权限已存在")
		return
	}
	req.CreateBy = loginUser.UserName
	err := s.Services.SysRoleService.InsertRole(&req)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, req)

}
// UpdateRole 修改角色
//
//	@Summary	修改角色
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"角色信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role [put]
//	@Security	BearerAuth
func (s *RoleHandler) UpdateRole(c *gin.Context) {
	response.SetOperTitle(c, "修改角色")
	req := s.Api.SysRole.UpdateRoleRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err := s.Services.SysRoleService.CheckRoleAllowed(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysRoleService.CheckRoleDataScope(loginUser.UserId, req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	if !s.Services.SysRoleService.CheckRoleNameUnique(req.Id, req.RoleName) {
		response.Error(c, "修改角色'"+req.RoleName+"'失败，角色名称已存在")
		return
	} else if !s.Services.SysRoleService.CheckRoleKeyUnique(req.Id, req.RoleKey) {
		response.Error(c, "修改角色'"+req.RoleKey+"'失败，角色权限已存在")
		return
	}
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err = s.Services.SysRoleService.UpdateRole(req.Id, req.MenuIds, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	// 更新缓存用户权限
	var user model.SysUser
	if !user.IsAdmin(loginUser.UserId) {
		permissions, err := s.Services.PermissionService.GetMenuPermission(loginUser.User)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		loginUser.Permissions = permissions
		userObj, err := model.SysUser{}.Get("id = ?", loginUser.UserId)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
		loginUser.User = userObj
		tokenService := s.Services.TokenService.New(c, db.RedisConnections["master"], config.GlobalConfig.Jwt.Secret, config.GlobalConfig.Jwt.ExpirationTime)
		err = tokenService.SetLoginUser(loginUser)
		if err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
	}
	response.Data(c, nil)
}
// DataScope 角色数据权限
//
//	@Summary	角色数据权限
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"数据权限信息"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role/dataScope [put]
//	@Security	BearerAuth
func (s *RoleHandler) DataScope(c *gin.Context) {
	response.SetOperTitle(c, "设置角色数据权限")
	req := s.Api.SysRole.UpdateRoleRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err := s.Services.SysRoleService.CheckRoleAllowed(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = s.Services.SysRoleService.CheckRoleDataScope(loginUser.UserId, req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	upd := utils.StructToMapWithGormColumn(req)
	upd["update_by"] = loginUser.UserName
	err = s.Services.SysRoleService.AuthDataScope(req.Id, req.DeptIds, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// ChangeRoleStatus 角色状态修改
//
//	@Summary	角色状态修改
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含id和status"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role/changeStatus [put]
//	@Security	BearerAuth
func (s *RoleHandler) ChangeRoleStatus(c *gin.Context) {
	response.SetOperTitle(c, "修改角色状态")
	req := s.Api.Common.ChangeStatusRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err := s.Services.SysRoleService.CheckRoleAllowed(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err = s.Services.SysRoleService.CheckRoleDataScope(loginUser.UserId, req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	upd := map[string]interface{}{
		"status":    req.Status,
		"update_by": loginUser.UserName,
	}
	err = s.Services.SysRoleService.UpdateRoleStatus(req.Id, upd)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)

}
// DelRole 删除角色
//
//	@Summary	删除角色
//	@Tags		角色管理
//	@Produce	json
//	@Param		ids	path	string	true	"角色ID，多个用逗号分隔"
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/system/role/{ids} [delete]
//	@Security	BearerAuth
func (s *RoleHandler) DelRole(c *gin.Context) {
	response.SetOperTitle(c, "删除角色")
	req := s.Api.Common.IdsUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	ids := strings.Split(req.Ids, ",")
	loginUser := c.MustGet("loginUser").(*model.LoginUser)

	err := s.Services.SysRoleService.DeleteRoleByIds(loginUser.UserId, stringutils.StringSliceToUint64Slice(ids))
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, "删除失败服务错误")
		return
	}

	response.Data(c, nil)

}
// AllocatedUserList 查询角色已授权用户列表
//
//	@Summary	查询角色已授权用户列表
//	@Tags		角色管理
//	@Produce	json
//	@Param		roleId	query	int		true	"角色ID"
//	@Param		userName	query	string	false	"用户名称"
//	@Param		phonenumber	query	string	false	"手机号码"
//	@Success	200		{object}	map[string]interface{}"用户列表"
//	@Router		/system/role/authUser/allocatedList [get]
//	@Security	BearerAuth
func (s *RoleHandler) AllocatedUserList(c *gin.Context) {
	response.SetOperTitle(c, "查询已分配用户列表")
	req := s.Api.SysRole.GetAllocatedListRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, total, err := s.Services.SysUserService.SelectAllocatedList(req.RoleId, model.SysUser{UserName: req.UserName, Phonenumber: req.PhoneNumber}, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)

}
// UnallocatedUserList 查询角色未授权用户列表
//
//	@Summary	查询角色未授权用户列表
//	@Tags		角色管理
//	@Produce	json
//	@Param		roleId	query	int		true	"角色ID"
//	@Param		userName	query	string	false	"用户名称"
//	@Param		phonenumber	query	string	false	"手机号码"
//	@Success	200		{object}	map[string]interface{}"用户列表"
//	@Router		/system/role/authUser/unallocatedList [get]
//	@Security	BearerAuth
func (s *RoleHandler) UnallocatedUserList(c *gin.Context) {
	response.SetOperTitle(c, "查询未分配用户列表")
	req := s.Api.SysRole.GetAllocatedListRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	data, total, err := s.Services.SysUserService.SelectUnallocatedList(req.RoleId, model.SysUser{UserName: req.UserName, Phonenumber: req.PhoneNumber}, req.PageNum, req.PageSize)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.PageData(c, data, total)
}
// AuthUserCancel 取消用户授权角色
//
//	@Summary	取消用户授权角色
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含roleId和userId"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role/authUser/cancel [put]
//	@Security	BearerAuth
func (s *RoleHandler) AuthUserCancel(c *gin.Context) {
	response.SetOperTitle(c, "取消用户授权")
	req := s.Api.SysRole.AuthUserCancelRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err := s.Services.SysRoleService.DeleteAuthUser(req.RoleId, req.UserId)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)

}
// AuthUserCancelAll 批量取消用户授权角色
//
//	@Summary	批量取消用户授权角色
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含roleId和userIds"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role/authUser/cancelAll [put]
//	@Security	BearerAuth
func (s *RoleHandler) AuthUserCancelAll(c *gin.Context) {
	response.SetOperTitle(c, "批量取消用户授权")
	req := s.Api.SysRole.AuthUserCancelAllRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	userIds := make([]uint64, 0)
	for _, item := range strings.Split(req.UserIds, ",") {
		userId, _ := strconv.ParseUint(item, 10, 64)
		userIds = append(userIds, userId)
	}
	err := s.Services.SysRoleService.DeleteAuthUsers(req.RoleId, userIds)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// AuthUserSelectAll 授权用户选择
//
//	@Summary	授权用户选择
//	@Tags		角色管理
//	@Accept		json
//	@Produce	json
//	@Param		body	body		map[string]interface{}	true	"包含roleId和userIds"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/role/authUser/selectAll [put]
//	@Security	BearerAuth
func (s *RoleHandler) AuthUserSelectAll(c *gin.Context) {
	response.SetOperTitle(c, "批量授权用户")
	req := s.Api.SysRole.AuthUserCancelAllRequest
	if err := api.ShouldBindError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	loginUser := c.MustGet("loginUser").(*model.LoginUser)
	err := s.Services.SysRoleService.CheckRoleDataScope(loginUser.UserId, req.RoleId)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}

	userIds := make([]uint64, 0)
	for _, item := range strings.Split(req.UserIds, ",") {
		userId, _ := strconv.ParseUint(item, 10, 64)
		userIds = append(userIds, userId)
	}
	err = s.Services.SysRoleService.InsertAuthUsers(req.RoleId, userIds)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)

}
// DeptTreeSelect 根据角色ID查询部门树结构
//
//	@Summary	根据角色ID查询部门树结构
//	@Tags		角色管理
//	@Produce	json
//	@Param		id	path	int	true	"角色ID"
//	@Success	200	{object}	map[string]interface{}"部门树"
//	@Router		/system/role/deptTree/{id} [get]
//	@Security	BearerAuth
func (s *RoleHandler) DeptTreeSelect(c *gin.Context) {
	response.SetOperTitle(c, "查询部门下拉树")
	req := s.Api.Common.IdUriRequest
	if err := api.ShouldBindUriError(c, &req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	checkedKeys, err := s.Services.SysDeptService.SelectDeptListByRoleId(req.Id)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	depts, err := s.Services.SysDeptService.SelectDeptTreeList()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
	}
	response.DataExpand(c, nil, map[string]interface{}{
		"checkedKeys": checkedKeys,
		"depts":       depts,
	})
}

// Export 导出角色数据
//
//	@Summary	导出角色数据
//	@Tags		角色管理
//	@Produce	application/octet-stream
//	@Success	200	{file}	binary"Excel文件"
//	@Router		/system/role/export [post]
//	@Security	BearerAuth
func (s *RoleHandler) Export(c *gin.Context) {
	response.SetOperTitle(c, "导出角色数据")
	req := s.Api.SysRole.GetRoleListRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, err.Error())
		return
	}
	conditions := utils.BuildConditions(req)
	roleList, err := s.Services.SysRoleService.SelectRoleAllList(conditions)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	headers := []string{"角色编号", "角色名称", "权限字符", "显示顺序", "状态", "创建时间"}
	var data [][]string
	for _, role := range roleList {
		data = append(data, []string{
			fmt.Sprint(role.Id),
			role.RoleName,
			role.RoleKey,
			fmt.Sprint(role.RoleSort),
			role.Status,
			fmt.Sprint(role.CreatedAt),
		})
	}
	// 调用封装的函数直接将 Excel 数据写入响应
	file, err := exceltool.CreateExcelFile(headers, data, "角色数据")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.ExportExcel(c, "exported_data.xlsx", file)
}
