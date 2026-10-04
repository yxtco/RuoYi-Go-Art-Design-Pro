package system

type SysMenu struct {
	RoleIdUriRequest
	ListMenuRequest
	UpdateMenuRequest
}

type RoleIdUriRequest struct {
	RoleId uint64 `uri:"roleId" binding:"required" err_msg:"角色Id 必填"`
}

type ListMenuRequest struct {
	MenuName string `form:"menuName" query:"m.menu_name:like"`
	Status   string `form:"status" query:"m.status:="`
}
type UpdateMenuRequest struct {
	Id        uint64 `gorm:"column:id;primary_key" json:"id"            description:"" binding:"required" err_msg:"id 必填"`
	MenuName  string `gorm:"column:menu_name" json:"menuName"            description:"菜单名称" binding:"required" err_msg:"菜单名称 必填"`
	ParentId  uint64 `gorm:"column:parent_id" json:"parentId"            description:"父菜单ID"`
	OrderNum  *int   `gorm:"column:order_num" json:"orderNum"            description:"显示顺序" binding:"required" err_msg:"显示排序 必填"`
	Path      string `gorm:"column:path" json:"path"            description:"路由地址"`
	Component string `gorm:"column:component" json:"component"            description:"组件路径"`
	Query     string `gorm:"column:query" json:"query"            description:"路由参数"`
	IsFrame   string `gorm:"column:is_frame" json:"isFrame"            description:"是否为外链（0是 1否）"`
	IsCache   string `gorm:"column:is_cache" json:"isCache"            description:"是否缓存（0缓存 1不缓存）"`
	MenuType  string `gorm:"column:menu_type" json:"menuType"            description:"菜单类型（M目录 C菜单 F按钮）"`
	Visible   string `gorm:"column:visible" json:"visible"            description:"菜单状态（0显示 1隐藏）"`
	Status    string `gorm:"column:status" json:"status"            description:"菜单状态（0正常 1停用）"`
	Perms     string `gorm:"column:perms" json:"perms"            description:"权限标识"`
	Icon      string `gorm:"column:icon" json:"icon"            description:"菜单图标"`
	Remark    string `gorm:"column:remark" json:"remark"        description:"备注"`
}
