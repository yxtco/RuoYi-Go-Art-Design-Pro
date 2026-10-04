package common

import (
	"go-fin-server/internal/model"
	"go-fin-server/internal/service/system"
)

type PermissionService struct {
	SysRoleService system.SysRoleService
}

// GetRolePermission 获取角色数据权限
func (c PermissionService) GetRolePermission(user model.SysUser) ([]string, error) {
	roles := make([]string, 0) // set
	// 管理员拥有所有权限
	if user.IsAdmin(user.Id) {
		roles = append(roles, "admin")
	}
	perm, err := c.SysRoleService.SelectRolePermissionByUserId(user.Id)
	if err != nil {
		return roles, err
	}
	roles = append(roles, perm...)
	return roles, nil
}

// GetMenuPermission 获取菜单数据权限
func (c PermissionService) GetMenuPermission(user model.SysUser) ([]string, error) {
	permissions := make([]string, 0) // set
	// 管理员拥有所有权限
	if user.IsAdmin(user.Id) {
		permissions = append(permissions, "*:*:*")
	} else {
		var userRole model.SysUserRole
		userRoles, err := userRole.GetRolesByUserId(user.Id)
		if err != nil {
			return permissions, err
		}
		var menu model.SysMenu
		if len(userRoles) > 0 {
			userRoleIds := make([]uint64, 0)
			for _, item := range userRoles {
				userRoleIds = append(userRoleIds, item.RoleId)
			}
			// 多角色设置permissions属性，以便数据权限匹配权限
			rolePerms, err := menu.GetMenuPermsByRoleIds(userRoleIds)
			if err != nil {
				return permissions, err
			}
			permissions = append(permissions, rolePerms...)
		} else {
			userPerms, err := menu.GetMenuPermsByUserId(user.Id)
			if err != nil {
				return permissions, err
			}
			permissions = append(permissions, userPerms...)
		}
	}
	return permissions, nil
}
