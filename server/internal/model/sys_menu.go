package model

import (
	"fmt"
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
	"go-fin-server/pkg/utils/stringutils"
	"sort"
)

type SysMenu struct {
	Id        uint64 `gorm:"column:id;primary_key" json:"id"            description:""`
	MenuName  string `gorm:"column:menu_name" json:"menuName"            description:"菜单名称" binding:"required" err_msg:"菜单名称 必填"`
	ParentId  uint64 `gorm:"column:parent_id" json:"parentId"            description:"父菜单ID"`
	OrderNum  *int   `gorm:"column:order_num" json:"orderNum"            description:"显示顺序" binding:"required" err_msg:"显示排序 必填"`
	Path      string `gorm:"column:path" json:"path"            description:"路由地址"`
	Component string `gorm:"column:component" json:"component"            description:"组件路径"`
	Query     string `gorm:"column:query" json:"query"            description:"路由参数"`
	IsFrame   string `gorm:"column:is_frame;default:'1'" json:"isFrame"            description:"是否为外链（0是 1否）"`
	IsCache   string `gorm:"column:is_cache;default:'0'" json:"isCache"            description:"是否缓存（0缓存 1不缓存）"`
	MenuType  string `gorm:"column:menu_type" json:"menuType"            description:"菜单类型（M目录 C菜单 F按钮）"`
	Visible   string `gorm:"column:visible" json:"visible"            description:"菜单状态（0显示 1隐藏）"`
	Status    string `gorm:"column:status" json:"status"            description:"菜单状态（0正常 1停用）"`
	Perms     string `gorm:"column:perms" json:"perms"            description:"权限标识"`
	Icon      string `gorm:"column:icon;default:'#'" json:"icon"            description:"菜单图标"`
	Remark    string `gorm:"column:remark" json:"remark"        description:"备注"`
	BaseNoDelModel
}

func (c SysMenu) TableName() string {
	return "sys_menu"
}

func (c SysMenu) Get(query string, args ...interface{}) (SysMenu, error) {
	var data SysMenu
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysMenu) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(&SysMenu{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysMenu) Create(data *SysMenu) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysMenu) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysMenu{})
	return d.Error
}

func (c SysMenu) HasChildByMenuId(query string, args ...interface{}) (int64, error) {
	var count int64
	if err := db.DBConnections["master"].Model(&SysMenu{}).Where(query, args...).Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}

func (c SysMenu) GetMenuPermsByRoleIds(roleIds []uint64) ([]string, error) {
	// 如果 roleIds 为空，直接返回空结果，避免生成无效的 SQL (IN ())
	if len(roleIds) == 0 {
		return make([]string, 0), nil
	}
	sql := `
		select distinct m.perms
		from sys_menu m
			 left join sys_role_menu rm on m.id = rm.menu_id
		where m.status = '0' and rm.role_id in (?)
	`
	data := make([]string, 0)
	d := db.DBConnections["master"].Raw(sql, roleIds).Scan(&data)
	return data, d.Error
}

func (c SysMenu) GetMenuPermsByUserId(userId uint64) ([]string, error) {
	sql := `
		select distinct m.perms
		from sys_menu m
			 left join sys_role_menu rm on m.id = rm.menu_id
			 left join sys_user_role ur on rm.role_id = ur.role_id
			 left join sys_role r on r.id = ur.role_id
		where m.status = '0' and r.status = '0' and ur.user_id = ?
	`
	data := make([]string, 0)
	d := db.DBConnections["master"].Raw(sql, userId).Scan(&data)
	return data, d.Error
}

func (c SysMenu) SelectMenuList(conditions map[string]interface{}) ([]SysMenu, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysMenu{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysMenu, 0)
	d := query.Order("parent_id, order_num").Find(&data)

	return data, d.Error
}

func (c SysMenu) SelectMenuListByUserId(userId uint64, conditions map[string]interface{}) ([]SysMenu, error) {
	args := make([]interface{}, 0)
	args = append(args, userId)
	sql := `
		select distinct m.id, m.parent_id, m.menu_name, m.path, m.component, m.query, m.visible, m.status, ifnull(m.perms,'') as perms, m.is_frame, m.is_cache, m.menu_type, m.icon, m.order_num, m.create_time
		from sys_menu m
		left join sys_role_menu rm on m.id = rm.menu_id
		left join sys_user_role ur on rm.role_id = ur.role_id
		left join sys_role ro on ur.role_id = ro.id
		where ur.user_id = ?
	`
	// 添加条件
	for key, value := range conditions {
		sql += fmt.Sprintf(` AND %s`, key)
		args = append(args, value)
	}
	sql += ` order by m.parent_id, m.order_num`
	data := make([]SysMenu, 0)
	d := db.DBConnections["master"].Raw(sql, args...).Scan(&data)
	return data, d.Error
}

func (c SysMenu) GetMenuByUserId(userId uint64) ([]*SysMenuTreeNode, error) {
	sql := `
		select distinct m.id, m.parent_id, m.menu_name, m.path, m.component, m.query, m.visible, m.status, ifnull(m.perms,'') as perms, m.is_frame, m.is_cache, m.menu_type, m.icon, m.order_num, m.create_time
		from sys_menu m
			 left join sys_role_menu rm on m.id = rm.menu_id
			 left join sys_user_role ur on rm.role_id = ur.role_id
			 left join sys_role ro on ur.role_id = ro.id
			 left join sys_user u on ur.user_id = u.id
		where u.id = ?  and m.status = 0  AND ro.status = 0 and m.menu_type not in ('F')
		order by m.parent_id, m.order_num
	`
	data := make([]*SysMenuTreeNode, 0)
	d := db.DBConnections["master"].Raw(sql, userId).Scan(&data)
	return data, d.Error
}
func (c SysMenu) SelectMenuListByRoleId(roleId uint64, isMenuCheckStrictly bool) ([]uint64, error) {
	args := make([]interface{}, 0)
	args = append(args, roleId)
	sql := `
		select m.id
		from sys_menu m
		left join sys_role_menu rm on m.id = rm.menu_id
		where rm.role_id = ?
	`
	if isMenuCheckStrictly {
		sql += ` and m.id not in (select m.parent_id from sys_menu m inner join sys_role_menu rm on m.id = rm.menu_id and rm.role_id = ?) `
		args = append(args, roleId)
	}
	sql += ` order by m.parent_id, m.order_num`
	data := make([]uint64, 0)
	d := db.DBConnections["master"].Raw(sql, args...).Scan(&data)
	return data, d.Error
}

func (c SysMenu) GetMenuAll() ([]*SysMenuTreeNode, error) {
	sql := `
		select distinct m.id, m.parent_id, m.menu_name, m.path, m.component, m.query, m.visible, m.status, ifnull(m.perms,'') as perms, m.is_frame, m.is_cache, m.menu_type, m.icon, m.order_num, m.create_time
		from sys_menu m where m.status = 0 and m.menu_type not in ('F')
		order by m.parent_id, m.order_num
	`
	data := make([]*SysMenuTreeNode, 0)
	d := db.DBConnections["master"].Raw(sql).Scan(&data)
	return data, d.Error
}

type SysMenuTreeNode struct {
	Id       uint64 `gorm:"column:id;primary_key" json:"-"            description:""`
	MenuName string `gorm:"column:menu_name" json:"-"            description:"菜单名称"`
	ParentId uint64 `gorm:"column:parent_id" json:"-"            description:"父菜单ID"`
	OrderNum int    `gorm:"column:order_num" json:"-"            description:"显示顺序"`

	Query      string             `gorm:"column:query" json:"-"            description:"路由参数"`
	IsFrame    string             `gorm:"column:is_frame" json:"-"            description:"是否为外链（0是 1否）"`
	IsCache    string             `gorm:"column:is_cache" json:"-"            description:"是否缓存（0缓存 1不缓存）"`
	MenuType   string             `gorm:"column:menu_type" json:"-"            description:"菜单类型（M目录 C菜单 F按钮）"`
	Visible    string             `gorm:"column:visible" json:"-"            description:"菜单状态（0显示 1隐藏）"`
	Status     string             `gorm:"column:status" json:"-"            description:"菜单状态（0正常 1停用）"`
	Perms      string             `gorm:"column:perms" json:"-"            description:"权限标识"`
	Icon       string             `gorm:"column:icon" json:"-"            description:"菜单图标"`
	Remark     string             `gorm:"column:remark" json:"-"        description:"备注"`
	Name       string             `gorm:"-" json:"name"`
	Path       string             `gorm:"column:path" json:"path"            description:"路由地址"`
	Hidden     bool               `gorm:"-" json:"hidden"`
	Redirect   *string            `gorm:"-" json:"redirect,omitempty"`
	Component  string             `gorm:"column:component" json:"component"            description:"组件路径"`
	AlwaysShow *bool              `gorm:"-" json:"alwaysShow,omitempty"`
	Meta       *MenuTreeMeta      `gorm:"-" json:"meta"`
	Children   []*SysMenuTreeNode `gorm:"-" json:"children,omitempty"`
}

type MenuTreeMeta struct {
	Title   string `json:"title"`
	Icon    string `json:"icon"`
	NoCache bool   `json:"noCache"`
	Link    string `json:"link"`
}

type SysMenuNode struct {
	Id       uint64         `json:"id"`
	Label    string         `json:"label"`
	ParentId uint64         `json:"-" `
	OrderNum int            `json:"-"`
	Children []*SysMenuNode `json:"children,omitempty"`
}

func (c SysMenuNode) SortChildren() {
	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].OrderNum < c.Children[j].OrderNum
	})

	// Recursively sort children's children
	for _, child := range c.Children {
		child.SortChildren()
	}
}

// BuildMenuTree 将扁平的菜单数据转换为树形结构
func (c SysMenu) BuildMenuTree(menuItems []*SysMenuTreeNode) []*SysMenuTreeNode {
	menuMap := make(map[uint64]*SysMenuTreeNode)

	// 构建菜单映射
	for _, menuItem := range menuItems {
		menuMap[menuItem.Id] = menuItem
	}

	// 将菜单项连接起来
	var tree []*SysMenuTreeNode
	for _, menuItem := range menuMap {
		parentID := menuItem.ParentId
		menuItem.Name = menuItem.getRouteName()
		menuItem.Path = menuItem.getRouterPath()
		menuItem.Component = menuItem.getComponent()
		menuItem.Hidden = false // 菜单状态（0显示 1隐藏）
		if menuItem.Visible == "1" {
			menuItem.Hidden = true
		}
		menuItem.Query = menuItem.Query
		// setMeta
		menuItem.Meta = &MenuTreeMeta{
			Title:   menuItem.MenuName,
			Icon:    menuItem.Icon,
			NoCache: true,
			Link:    "",
		}

		if stringutils.StartsWithAny(menuItem.Path, "http://", "https://") {
			menuItem.Meta.Link = menuItem.Path
		}
		if menuItem.IsCache == "0" {
			menuItem.Meta.NoCache = false
		}
		if parentID == 0 {
			// 如果没有父菜单项，则将自身作为根节点添加
			tree = append(tree, menuItem)
		} else {
			// 将自身作为子菜单项添加到父菜单项的 Children 中
			parent, exists := menuMap[parentID]
			if exists {
				if parent.MenuType == "M" {
					// 当你一个路由下面的 children 声明的路由大于1个时，自动会变成嵌套的模式--如组件页面
					alwaysShow := true
					parent.AlwaysShow = &alwaysShow
					redirect := "noRedirect"
					parent.Redirect = &redirect
				}
				parent.Children = append(parent.Children, menuItem)
			} else if menuItem.isMenuFrame() {
				parent.Meta = nil
				menuItem.Path = menuItem.Path
				menuItem.Component = menuItem.Component
				menuItem.Name = stringutils.Capitalize(menuItem.Path)
			} else if menuItem.ParentId == 0 && menuItem.isInnerLink() {
				parent.Path = "/"
				routerPath := menuItem.innerLinkReplaceEach()
				menuItem.Path = routerPath
				menuItem.Component = "InnerLink"
				menuItem.Name = stringutils.Capitalize(routerPath)
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

func (c SysMenuTreeNode) isInnerLink() bool {
	return c.IsFrame == "1" && stringutils.StartsWithAny(c.Path, "http://", "https://")
}

func (c SysMenuTreeNode) innerLinkReplaceEach() string {
	return stringutils.ReplaceEach(c.Path, []string{"http://", "https://", "www.", ".", ":"}, []string{"", "", "", "/", "/"}, false, 0)
}

func (c SysMenuTreeNode) isMenuFrame() bool {
	return c.ParentId == 0 && c.MenuType == "C" && c.IsFrame == "1"
}

func (c SysMenuTreeNode) isParentView() bool {
	return c.ParentId != 0 && c.MenuType == "M"
}

func (c SysMenuTreeNode) getRouteName() string {
	routerName := stringutils.Capitalize(c.Path)
	// 非外链并且是一级目录（类型为目录）
	if c.isMenuFrame() {
		routerName = ""
	}
	return routerName
}

func (c SysMenuTreeNode) getRouterPath() string {
	routerPath := c.Path
	// 内链打开外网方式
	if c.ParentId != 0 && c.isInnerLink() {
		routerPath = c.innerLinkReplaceEach()
	}
	// 非外链并且是一级目录（类型为目录）
	if c.ParentId == 0 && c.MenuType == "M" && c.IsFrame == "1" {
		routerPath = "/" + c.Path
	} else if c.isMenuFrame() { // 非外链并且是一级目录（类型为菜单）
		routerPath = "/"
	}
	return routerPath
}

func (c SysMenuTreeNode) getComponent() string {
	component := "Layout"
	if c.Component != "" && !c.isMenuFrame() {
		component = c.Component
	} else if c.Component == "" && c.ParentId != 0 && c.isInnerLink() {
		component = "InnerLink"
	} else if c.Component == "" && c.isParentView() {
		component = "ParentView"
	}
	return component
}

func (c SysMenuTreeNode) SortChildren() {
	sort.Slice(c.Children, func(i, j int) bool {
		return c.Children[i].OrderNum < c.Children[j].OrderNum
	})

	// Recursively sort children's children
	for _, child := range c.Children {
		child.SortChildren()
	}
}
