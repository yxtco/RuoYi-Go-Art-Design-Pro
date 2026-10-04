package system

import (
	"errors"
	"fmt"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/utils/stringutils"
	"strings"

	"gorm.io/gorm"
)

type SysDeptService struct {
}

// SelectDeptList 查询部门管理数据
func (c SysDeptService) SelectDeptList(conditions map[string]interface{}, sortConditions []string) ([]model.SysDept, error) {
	data, err := model.SysDept{}.SelectDeptList(conditions, sortConditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectDataScopeDeptList 查询部门管理数据
func (c SysDeptService) SelectDataScopeDeptList(conditions map[string]interface{}, sortConditions []string, user model.SysUser) ([]model.SysDept, error) {
	data, err := model.SysDept{}.SelectDataScopeDeptList(conditions, sortConditions, user)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectDeptById 根据部门ID查询信息
func (c SysDeptService) SelectDeptById(deptId uint64) (model.SysDept, error) {
	data, err := model.SysDept{}.Get("id = ?", deptId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// SelectDeptListByRoleId 根据角色ID查询部门树信息
func (c SysDeptService) SelectDeptListByRoleId(roleId uint64) ([]uint64, error) {
	role, err := model.SysRole{}.Get("id = ?", roleId)
	if err != nil {
		return nil, err
	}
	data, err := model.SysRole{}.SelectDeptListByRoleId(roleId, role.DeptCheckStrictly)
	if err != nil {
		return data, err
	}
	return data, nil
}

// SelectDeptTreeList 查询部门树结构信息
func (c SysDeptService) SelectDeptTreeList() ([]*model.SysDeptTreeNode, error) {
	var dept model.SysDeptTreeNode
	data, err := dept.GetDeptTreeList()
	if err != nil {
		return nil, nil
	}
	dataTree := dept.BuildSysDeptTree(data)
	return dataTree, err
}

// HasChildByDeptId 是否存在子节点
func (c SysDeptService) HasChildByDeptId(deptId uint64) (bool, error) {
	count, err := model.SysDept{}.HasChildByDeptId("parent_id = ?", deptId)
	if err != nil {
		return count > 0, err
	}
	return count > 0, nil
}

// CheckDeptExistUser 查询部门是否存在用户
func (c SysDeptService) CheckDeptExistUser(deptId uint64) (bool, error) {
	count, err := model.SysDept{}.CheckDeptExistUser("dept_id = ?", deptId)
	if err != nil {
		return count > 0, err
	}
	return count > 0, nil
}

// DeleteDeptById 删除部门管理信息
func (c SysDeptService) DeleteDeptById(deptId uint64) error {
	err := model.SysDept{}.Delete("id = ?", deptId)
	if err != nil {
		return err
	}
	return nil
}

// CheckDeptDataScope 校验部门是否有数据权限
func (c SysDeptService) CheckDeptDataScope(userId, deptId uint64) error {
	var user model.SysUser
	if !user.IsAdmin(userId) {
		data, err := c.SelectDeptList(map[string]interface{}{
			"id = ?": deptId,
		}, []string{"parent_id", "order_num"})
		if err != nil {
			return err
		}
		if len(data) <= 0 {
			return errors.New("没有权限访问部门数据！")
		}
	}
	return nil
}

// CheckDeptNameUnique 校验部门名称是否唯一
func (c SysDeptService) CheckDeptNameUnique(deptId uint64, deptName string) bool {
	var dept model.SysDept
	deptObj, err := dept.Get("dept_name = ?", deptName)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if deptObj.Id != deptId {
		return false
	}
	return true
}

// InsertDept 新增保存部门信息
func (c SysDeptService) InsertDept(data *model.SysDept) error {
	info, err := model.SysDept{}.Get("id = ?", data.ParentId)
	if err != nil {
		return err
	}
	// 如果父节点不为正常状态,则不允许新增子节点
	if !("0" == info.Status) {
		return errors.New("部门停用，不允许新增")
	}
	data.Ancestors = fmt.Sprintf("%s,%d", info.Ancestors, data.ParentId)
	err = data.Create(data)
	if err != nil {
		return err
	}
	return nil
}

// UpdateDept 修改保存部门信息
func (c SysDeptService) UpdateDept(deptId, ParentId uint64, upd map[string]interface{}) error {
	newParentDept, err := model.SysDept{}.Get("id = ?", ParentId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	oldDept, err := model.SysDept{}.Get("id = ?", deptId)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if newParentDept.Id != 0 && oldDept.Id != 0 {
		newAncestors := fmt.Sprintf("%s,%d", newParentDept.Ancestors, newParentDept.Id)
		oldAncestors := oldDept.Ancestors
		upd["ancestors"] = newAncestors
		err = c.UpdateDeptChildren(deptId, newAncestors, oldAncestors)
		if err != nil {
			return err
		}
	}
	err = model.SysDept{}.UpdateMap(upd, "id = ?", deptId)
	if err != nil {
		return err
	}
	if "0" == upd["status"].(string) && upd["ancestors"].(string) != "" && !("0" == upd["ancestors"].(string)) {
		// 如果该部门是启用状态，则启用该部门的所有上级部门
		err = c.UpdateParentDeptStatusNormal(upd)
		if err != nil {
			return err
		}
	}
	return nil
}

// SelectNormalChildrenDeptById 根据ID查询所有子部门（正常状态）
func (c SysDeptService) SelectNormalChildrenDeptById(deptId uint64) (int64, error) {
	count, err := model.SysDept{}.SelectNormalChildrenDeptById(deptId)
	if err != nil {
		return count, err
	}
	return count, nil
}

// UpdateDeptChildren 修改子元素关系
func (c SysDeptService) UpdateDeptChildren(deptId uint64, newAncestors, oldAncestors string) error {
	children, err := model.SysDept{}.SelectChildrenDeptById(deptId)
	if err != nil {
		return err
	}
	for _, child := range children {
		child.Ancestors = strings.Replace(child.Ancestors, oldAncestors, newAncestors, 1)
	}
	if len(children) > 0 {
		err := model.SysDept{}.UpdateDeptChildren(children)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateParentDeptStatusNormal 修改该部门的父级部门状态
func (c SysDeptService) UpdateParentDeptStatusNormal(upd map[string]interface{}) error {
	deptIds := stringutils.StringSliceToUint64Slice(strings.Split(upd["ancestors"].(string), ","))
	err := model.SysDept{}.UpdateMap(map[string]interface{}{
		"status": "0",
	}, "id in (?)", deptIds)
	if err != nil {
		return err
	}
	return nil
}
