package system

import "go-fin-server/api/common"

type SysConfig struct {
	ListConfigRequest
	ConfigKeyUriRequest
	UpdateConfigRequest
}

type ListConfigRequest struct {
	ConfigName string `form:"configName" query:"config_name:like"`
	ConfigKey  string `form:"configKey" query:"config_key:like"`
	ConfigType string `form:"configType" query:"config_type:="`
	BeginTime  string `form:"beginTime" query:"date(create_time):>="`
	EndTime    string `form:"endTime" query:"date(create_time):<="`
	Sort       string `form:"sort,default=create_time:desc"`
	common.PageReq
}

type ConfigKeyUriRequest struct {
	ConfigKey string `uri:"configKey" binding:"required" err_msg:"configKey 必填"`
}

type UpdateConfigRequest struct {
	Id          uint64 `gorm:"column:id;primary_key" json:"id"    description:"参数主键" binding:"required" err_msg:"参数主键 必填"`
	ConfigName  string `gorm:"column:config_name" json:"configName"  description:"参数名称" binding:"required" err_msg:"参数名称 必填"`
	ConfigKey   string `gorm:"column:config_key" json:"configKey"  description:"参数键名" binding:"required" err_msg:"参数键名 必填"`
	ConfigValue string `gorm:"column:config_value" json:"configValue"  description:"参数键值" binding:"required" err_msg:"参数键值 必填"`
	ConfigType  string `gorm:"column:config_type" json:"configType"  description:"系统内置（Y是 N否）"`
	Remark      string `gorm:"column:remark" json:"remark"    description:"备注"`
}
