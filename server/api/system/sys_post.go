package system

import "go-fin-server/api/common"

type SysPost struct {
	ListPostRequest
	UpdatePostRequest
}

type ListPostRequest struct {
	PostCode string `form:"postCode" query:"postCode:like"`
	PostName string `form:"postName" query:"postName:like"`
	Status   string `form:"status" query:"status:="`
	common.PageReq
}

type UpdatePostRequest struct {
	Id       uint64 `gorm:"column:id;primary_key" json:"id"    description:"岗位ID" binding:"required" err_msg:"岗位ID 必填"`
	PostCode string `gorm:"column:post_code" json:"postCode"  description:"岗位编码" binding:"required" err_msg:"岗位编码 必填"`
	PostName string `gorm:"column:post_name" json:"postName"  description:"岗位名称"`
	PostSort *int   `gorm:"column:post_sort" json:"postSort"  description:"显示顺序" binding:"required" err_msg:"岗位顺序 必填"`
	Status   string `gorm:"column:status" json:"status"    description:"状态（0正常 1停用）"`
	Remark   string `gorm:"column:remark" json:"remark"    description:"备注"`
}
