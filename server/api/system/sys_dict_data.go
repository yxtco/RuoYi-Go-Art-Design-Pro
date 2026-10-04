package system

import "go-fin-server/api/common"

type SysDictData struct {
	ListDictDataRequest
	UpdateDictDataRequest
}

type ListDictDataRequest struct {
	DictType  string `form:"dictType" query:"dict_type:like"`
	DictLabel string `form:"dictLabel" query:"dict_label:like"`
	Status    string `form:"status" query:"status:="`
	Sort      string `form:"sort,default=dict_sort:asc"`
	common.PageReq
}

type UpdateDictDataRequest struct {
	DictCode  uint64 `gorm:"column:dict_code;primary_key;" json:"dictCode"        description:"字典编码" binding:"required" err_msg:"字典编码 必填"`
	DictSort  *uint  `gorm:"column:dict_sort" json:"dictSort"  description:"字典排序" err_msg:"显示排序 必填"`
	DictLabel string `gorm:"column:dict_label" json:"dictLabel" description:"字典标签" binding:"required" err_msg:"数据标签 必填"`
	DictValue string `gorm:"column:dict_value" json:"dictValue" description:"字典键值" binding:"required" err_msg:"字典键值 必填"`
	DictType  string `gorm:"-" json:"dictType"  description:"字典类型" binding:"required" err_msg:"字典类型 必填"`
	CssClass  string `gorm:"column:css_class" json:"cssClass"  description:"样式属性（其他样式扩展）"`
	ListClass string `gorm:"column:list_class" json:"listClass"  description:"表格回显样式"`
	//IsDefault string `gorm:"column:is_default" json:"isDefault"  description:"是否默认（Y是 N否）"`
	Status string `gorm:"column:status" json:"status"    description:"状态（0正常 1停用）"`
	Remark string `gorm:"column:remark" json:"remark"    description:"备注"`
}
