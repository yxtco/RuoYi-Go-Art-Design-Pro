package system

import "go-fin-server/api/common"

type SysRole struct {
	GetRoleListRequest
	GetAllocatedListRequest
	AuthUserCancelRequest
	AuthUserCancelAllRequest
	UpdateRoleRequest
}

type GetRoleListRequest struct {
	RoleName  string `form:"roleName" query:"role_name:like"`
	RoleKey   string `form:"roleKey" query:"role_key:like"`
	Status    string `form:"status" query:"status:="`
	BeginTime string `form:"beginTime" query:"date(create_time):>="`
	EndTime   string `form:"endTime" query:"date(create_time):<="`
	common.PageReq
}

type GetAllocatedListRequest struct {
	RoleId      uint64 `form:"roleId" binding:"required" err_msg:"roleId 必填"`
	UserName    string `form:"userName"`
	PhoneNumber string `form:"phonenumber"`
	common.PageReq
}

type AuthUserCancelRequest struct {
	RoleId uint64 `form:"roleId" binding:"required" err_msg:"roleId 必填"`
	UserId uint64 `form:"userId" binding:"required" err_msg:"UserId 必填"`
}

type AuthUserCancelAllRequest struct {
	RoleId  uint64 `form:"roleId" binding:"required" err_msg:"roleId 必填"`
	UserIds string `form:"userIds" binding:"required" err_msg:"UserId 必填"`
}

type UpdateRoleRequest struct {
	Id                uint64   `gorm:"column:id;primary_key;" json:"id"        description:"" binding:"required" err_msg:"角色ID 必填"`
	RoleName          string   `gorm:"column:role_name" json:"roleName"        description:"角色名称" binding:"required" err_msg:"角色名称 必填"`
	RoleKey           string   `gorm:"column:role_key" json:"roleKey"        description:"角色权限字符串" binding:"required" err_msg:"权限字符 必填"`
	RoleSort          *uint    `gorm:"column:role_sort" json:"roleSort"        description:"显示顺序" binding:"required" err_msg:"角色顺序 必填"`
	DataScope         string   `gorm:"column:data_scope" json:"dataScope" description:"数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限）"`
	MenuCheckStrictly bool     `gorm:"column:menu_check_strictly" json:"menuCheckStrictly"        description:"菜单树选择项是否关联显示"`
	DeptCheckStrictly bool     `gorm:"column:dept_check_strictly" json:"deptCheckStrictly"        description:"部门树选择项是否关联显示"`
	Status            string   `gorm:"column:status" json:"status"    description:"角色状态（0正常 1停用）"`
	Remark            string   `gorm:"column:remark" json:"remark"        description:"备注"`
	Flag              bool     `gorm:"-" json:"flag"        description:"是否已拥有的角色"`
	MenuIds           []uint64 `gorm:"-" json:"menuIds"`
	DeptIds           []uint64 `gorm:"-" json:"deptIds"`
}
