package system

import (
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"sort"

	"gorm.io/gorm"
)

type SysMenuService struct {
}

// SelectMenuList 根据用户查询系统菜单列表
func (c SysMenuService) SelectMenuList(userId uint64, conditions map[string]interface{}) ([]model.SysMenu, error) {
	var user model.SysUser
	// 管理员显示所有菜单信息
	if user.IsAdmin(userId) {
		data, err := model.SysMenu{}.SelectMenuList(conditions)
		if err != nil {
			return data, err
		}
		return data, nil
	} else {
		data, err := model.SysMenu{}.SelectMenuListByUserId(userId, conditions)
		if err != nil {
			return data, err
		}
		return data, nil
	}
}

func (c SysMenuService) BuildMenuTreeSelect(menuItems []model.SysMenu) []*model.SysMenuNode {
	menuMap := make(map[uint64]*model.SysMenuNode)

	// 构建菜单映射
	for _, menuItem := range menuItems {
		menuMap[menuItem.Id] = &model.SysMenuNode{
			Id:       menuItem.Id,
			Label:    menuItem.MenuName,
			ParentId: menuItem.ParentId,
			OrderNum: *menuItem.OrderNum,
		}
	}

	// 将菜单项连接起来
	var tree []*model.SysMenuNode
	for _, menuItem := range menuMap {
		parentID := menuItem.ParentId
		if parentID == 0 {
			// 如果没有父菜单项，则将自身作为根节点添加
			tree = append(tree, menuItem)
		} else {
			// 将自身作为子菜单项添加到父菜单项的 Children 中
			parent, exists := menuMap[parentID]
			if exists {
				parent.Children = append(parent.Children, menuItem)
			}
		}
	}

	// 对最外层排序
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].OrderNum < tree[j].OrderNum
	})
	// 对子排序
	for _, d := range tree {
		d.SortChildren()
	}
	return tree
}

// SelectMenuListByRoleId 根据角色ID查询菜单树信息
func (c SysMenuService) SelectMenuListByRoleId(roleId uint64) ([]uint64, error) {
	role, err := model.SysRole{}.Get("id = ?", roleId)
	if err != nil {
		return nil, err
	}
	ids, err := model.SysMenu{}.SelectMenuListByRoleId(roleId, role.MenuCheckStrictly)
	if err != nil {
		return ids, err
	}
	return ids, nil
}

// SelectMenuById 根据菜单ID查询信息
func (c SysMenuService) SelectMenuById(menuId uint64) (model.SysMenu, error) {
	data, err := model.SysMenu{}.Get("id = ?", menuId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CheckMenuNameUnique 校验菜单名称是否唯一
func (c SysMenuService) CheckMenuNameUnique(menuId uint64, menuName string) bool {
	var menu model.SysMenu
	menuObj, err := menu.Get("menu_name = ?", menuName)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if menuObj.Id != menuId {
		return false
	}
	return true
}

// InsertMenu 新增保存菜单信息
func (c SysMenuService) InsertMenu(menu *model.SysMenu) error {
	err := menu.Create(menu)
	if err != nil {
		return err
	}
	return nil
}

// UpdateMenu 修改保存菜单信息
func (c SysMenuService) UpdateMenu(menuId uint64, upd map[string]interface{}) error {
	err := model.SysMenu{}.UpdateMap(upd, "id = ?", menuId)
	if err != nil {
		return err
	}
	return nil
}

// HasChildByMenuId 是否存在菜单子节点
func (c SysMenuService) HasChildByMenuId(menuId uint64) (bool, error) {
	count, err := model.SysMenu{}.HasChildByMenuId("parent_id = ?", menuId)
	if err != nil {
		return count > 0, err
	}
	return count > 0, nil
}

// CheckMenuExistRole 查询菜单使用数量
func (c SysMenuService) CheckMenuExistRole(menuId uint64) (bool, error) {
	count, err := model.SysRoleMenu{}.CheckMenuExistRole("menu_id = ?", menuId)
	if err != nil {
		return count > 0, err
	}
	return count > 0, nil
}

// DeleteMenuById 删除菜单管理信息
func (c SysMenuService) DeleteMenuById(menuId uint64) error {
	err := model.SysMenu{}.Delete("id = ?", menuId)
	if err != nil {
		return err
	}
	return nil
}
