package system

import "go-fin-server/api/common"

type SysDictType struct {
	ListDictTypeRequest
	UpdateDictTypeRequest
}

type ListDictTypeRequest struct {
	DictName  string `form:"dictName" query:"dictName:like"`
	DictType  string `form:"dictType" query:"dictType:like"`
	Status    string `form:"status" query:"status:="`
	BeginTime string `form:"beginTime" query:"date(create_time):>="`
	EndTime   string `form:"endTime" query:"date(create_time):<="`
	Sort      string `form:"sort,default=create_time:desc"`
	common.PageReq
}

type UpdateDictTypeRequest struct {
	Id       uint64 `gorm:"column:id;primary_key" json:"id"    description:"ID" binding:"required" err_msg:"id 必填"`
	DictName string `gorm:"column:dict_name" json:"dictName"  description:"字典名称" binding:"required" err_msg:"字典名称 必填"`
	DictType string `gorm:"column:dict_type" json:"dictType"  description:"字典类型" binding:"required" err_msg:"字典类型 必填"`
	Status   string `gorm:"column:status" json:"status"    description:"状态（0正常 1停用）"`
	Remark   string `gorm:"column:remark" json:"remark"    description:"备注"`
}
