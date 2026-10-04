package system

import (
	"errors"
	"fmt"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"

	"gorm.io/gorm"
)

type SysRoleService struct {
}

// SelectRoleList 根据条件分页查询角色数据
func (c SysRoleService) SelectRoleList(conditions map[string]interface{}, pageNum, pageSize int) ([]model.SysRole, int64, error) {
	var role model.SysRole
	data, total, err := role.SelectRoleList(conditions, pageNum, pageSize)
	if err != nil {
		return data, total, err
	}
	return data, total, nil
}

// SelectDataScopeRoleList 根据条件分页查询角色数据
func (c SysRoleService) SelectDataScopeRoleList(conditions map[string]interface{}, pageNum, pageSize int, user model.SysUser) ([]model.SysRole, int64, error) {
	var role model.SysRole
	data, total, err := role.SelectDataScopeRoleList(conditions, pageNum, pageSize, user)
	if err != nil {
		return data, total, err
	}
	return data, total, nil
}

// SelectRoleAllList 根据条件查询角色数据
func (c SysRoleService) SelectRoleAllList(conditions map[string]interface{}) ([]model.SysRole, error) {
	var role model.SysRole
	data, err := role.SelectRoleAllList(conditions)
	if err != nil {
		return data, err
	}
	return data, nil
}

// SelectRolesByUserId 获取所有角色并获取用户角色
func (c SysRoleService) SelectRolesByUserId(userId uint64) ([]model.SysRole, error) {
	userRoles, err := model.SysUserRole{}.GetRolesByUserId(userId)
	if err != nil {
		return nil, err
	}
	// 获取所有的角色
	roles, err := model.SysRole{}.GetAllList("status = ?", "0")
	if err != nil {
		return nil, err
	}
	for i, role := range roles {
		for _, userRole := range userRoles {
			if role.Id == userRole.SysRole.Id {
				roles[i].Flag = true
			}
		}
	}
	return roles, nil
}

// SelectRoleById 根据角色编号获取详细信息
func (c SysRoleService) SelectRoleById(roleId uint64) (model.SysRole, error) {
	data, err := model.SysRole{}.Get("id = ?", roleId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CheckRoleDataScope 校验角色是否有数据权限
func (c SysRoleService) CheckRoleDataScope(userId, roleId uint64) error {
	var user model.SysUser
	if !user.IsAdmin(userId) {
		roles, err := model.SysRole{}.GetAllList("id = ?", roleId)
		if err != nil {
			return err
		}
		if len(roles) <= 0 {
			return errors.New("没有权限访问角色数据！")
		}
	}
	return nil
}

// CheckRoleNameUnique 校验角色名称是否唯一
func (c SysRoleService) CheckRoleNameUnique(roleId uint64, roleName string) bool {
	data, err := model.SysRole{}.Get("role_name = ?", roleName)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if data.Id != roleId {
		return false
	}
	return true
}

// CheckRoleKeyUnique 校验角色名称是否唯一
func (c SysRoleService) CheckRoleKeyUnique(roleId uint64, roleKey string) bool {
	data, err := model.SysRole{}.Get("role_key = ?", roleKey)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if data.Id != roleId {
		return false
	}
	return true
}

// insertRole 新增保存角色信息
func (c SysRoleService) InsertRole(role *model.SysRoleAll) error {
	// 新增角色信息
	err := role.Create(&role.SysRole)
	if err != nil {
		return err
	}
	err = c.InsertRoleMenu(role.Id, role.MenuIds)
	if err != nil {
		return err
	}
	return nil
}

// InsertRoleMenu 新增角色菜单信息
func (c SysRoleService) InsertRoleMenu(roleId uint64, menuIds []uint64) error {
	var roleMenu model.SysRoleMenu
	roleMenuList := make([]model.SysRoleMenu, 0)
	for _, item := range menuIds {
		roleMenuList = append(roleMenuList, model.SysRoleMenu{
			RoleId: roleId,
			MenuId: item,
		})
	}
	err := roleMenu.BatchRoleMenu(roleMenuList)
	if err != nil {
		return err
	}
	return nil

}

// DeleteRoleById 通过角色ID删除角色
func (c SysRoleService) DeleteRoleById(roleId uint64) error {
	// 删除角色与菜单关联
	err := model.SysRoleMenu{}.Delete("role_id = ?", roleId)
	if err != nil {
		return err
	}
	// 删除角色与部门关联
	err = model.SysRoleDept{}.Delete("role_id = ?", roleId)
	if err != nil {
		return err
	}
	// 删除角色
	err = model.SysRole{}.Delete("id = ", roleId)
	if err != nil {
		return err
	}
	return nil
}

// DeleteRoleByIds 批量删除角色信息
func (c SysRoleService) DeleteRoleByIds(userId uint64, roleIds []uint64) error {

	for _, roleId := range roleIds {
		err := c.CheckRoleAllowed(roleId)
		if err != nil {
			return err
		}
		err = c.CheckRoleDataScope(userId, roleId)
		if err != nil {
			return err
		}
		roleObj, err := model.SysRole{}.Get("id = ?", roleId)
		if err != nil {
			return err
		}
		count, err := model.SysUserRole{}.CountUserRoleByRoleId(roleId)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New(fmt.Sprintf("%s已分配,不能删除", roleObj.RoleName))
		}
	}
	// 删除角色与菜单关联
	err := model.SysRoleMenu{}.Delete("role_id in (?)", roleIds)
	if err != nil {
		return err
	}
	// 删除角色与部门关联
	err = model.SysRoleDept{}.Delete("role_id in (?)", roleIds)
	if err != nil {
		return err
	}
	// 删除角色
	err = model.SysRole{}.Delete("id in (?)", roleIds)
	if err != nil {
		return err
	}
	return nil
}

// CheckRoleAllowed 校验角色是否允许操作
func (c SysRoleService) CheckRoleAllowed(roleId uint64) error {
	var role model.SysRole
	if role.IsAdmin(roleId) {
		return errors.New("不允许操作超级管理员用户")
	}
	return nil
}

// UpdateRole 修改保存角色信息
func (c SysRoleService) UpdateRole(roleId uint64, menuIds []uint64, upd map[string]interface{}) error {
	// 修改角色信息
	err := model.SysRole{}.UpdateMap(upd, "id = ?", roleId)
	if err != nil {
		return err
	}
	// 删除角色与菜单关联
	err = model.SysRoleMenu{}.Delete("role_id = ?", roleId)
	if err != nil {
		return err
	}
	err = c.InsertRoleMenu(roleId, menuIds)
	if err != nil {
		return err
	}
	return nil
}

// UpdateRoleStatus 修改保存角色信息
func (c SysRoleService) UpdateRoleStatus(roleId uint64, upd map[string]interface{}) error {
	err := model.SysRole{}.UpdateMap(upd, "id = ?", roleId)
	if err != nil {
		return err
	}
	return nil
}

// SelectRolePermissionByUserId 根据用户ID查询权限
func (c SysRoleService) SelectRolePermissionByUserId(userId uint64) ([]string, error) {
	permissions := make([]string, 0) // set

	perms, err := model.SysRole{}.SelectRolePermissionByUserId(userId)
	if err != nil {
		return permissions, err
	}

	roleKeys := make([]string, 0)
	for _, item := range perms {
		roleKeys = append(roleKeys, item.RoleKey)
	}
	permissions = append(permissions, roleKeys...)

	return permissions, nil
}

// DeleteAuthUser 取消授权用户
func (c SysRoleService) DeleteAuthUser(roleId, userId uint64) error {
	err := model.SysUserRole{}.Delete("role_id = ? and user_id = ?", roleId, userId)
	if err != nil {
		return err
	}
	return nil
}

// DeleteAuthUsers 批量取消授权用户 取消授权用户
func (c SysRoleService) DeleteAuthUsers(roleId uint64, userIds []uint64) error {
	err := model.SysUserRole{}.Delete("role_id = ? and user_id in (?)", roleId, userIds)
	if err != nil {
		return err
	}
	return nil
}

// InsertAuthUsers 批量选择用户授权
func (c SysRoleService) InsertAuthUsers(roleId uint64, userIds []uint64) error {
	var userRole model.SysUserRole
	roleList := make([]model.SysUserRole, 0)
	for _, item := range userIds {
		roleList = append(roleList, model.SysUserRole{
			RoleId: roleId,
			UserId: item,
		})
	}
	err := userRole.BatchUserRole(roleList)
	if err != nil {
		return err
	}
	return nil
}

// AuthDataScope 修改保存数据权限
func (c SysRoleService) AuthDataScope(roleId uint64, deptIds []uint64, upd map[string]interface{}) error {
	// 修改角色信息
	err := model.SysRole{}.UpdateMap(upd, "id = ?", roleId)
	if err != nil {
		return err
	}
	// 删除角色与部门关联
	err = model.SysRoleDept{}.Delete("role_id in (?)", roleId)
	if err != nil {
		return err
	}
	// 新增角色和部门信息（数据权限）
	err = c.InsertRoleDept(roleId, deptIds)
	if err != nil {
		return err
	}
	return nil
}

// InsertRoleDept 新增角色部门信息(数据权限)
func (c SysRoleService) InsertRoleDept(roleId uint64, deptIds []uint64) error {
	// 新增角色与部门（数据权限）管理
	data := make([]model.SysRoleDept, 0)
	for _, deptId := range deptIds {
		data = append(data, model.SysRoleDept{
			RoleId: roleId,
			DeptId: deptId,
		})
	}
	err := model.SysRoleDept{}.BatchRoleDept(data)
	if err != nil {
		return err
	}
	return nil
}
