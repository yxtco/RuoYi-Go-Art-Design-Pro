package system

type SysDept struct {
	ListDeptRequest
	UpdateDeptRequest
}

type ListDeptRequest struct {
	DeptName string `form:"deptName" query:"dept_name:like"`
	Status   string `form:"status" query:"status:="`
	Sort     string `form:"sort,default=parent_id:asc,order_num:asc"`
}

type UpdateDeptRequest struct {
	Id        uint64  `gorm:"column:id;primary_key;" json:"id"        description:"部门id" binding:"required" err_msg:"部门id 必填"`
	ParentId  *uint64 `gorm:"column:parent_id" json:"parentId"  description:"父部门id" binding:"required" err_msg:"上级部门 必填"`
	Ancestors string  `gorm:"column:ancestors" json:"ancestors" description:"祖级列表"`
	DeptName  string  `gorm:"column:dept_name" json:"deptName"  description:"部门名称" binding:"required" err_msg:"部门名称 必填"`
	OrderNum  *int    `gorm:"column:order_num" json:"orderNum"  description:"显示顺序" binding:"required" err_msg:"显示顺序 必填"`
	Leader    string  `gorm:"column:leader" json:"leader"    description:"负责人"`
	Phone     string  `gorm:"column:phone" json:"phone"     description:"联系电话"`
	Email     string  `gorm:"column:email" json:"email"     description:"邮箱"`
	Status    string  `gorm:"column:status" json:"status"    description:"部门状态（0正常 1停用）"`
}
