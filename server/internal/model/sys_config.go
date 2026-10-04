package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysConfig struct {
	Id          uint64 `gorm:"column:id;primary_key" json:"id"    description:"参数主键"`
	ConfigName  string `gorm:"column:config_name" json:"configName"  description:"参数名称" binding:"required" err_msg:"参数名称 必填"`
	ConfigKey   string `gorm:"column:config_key" json:"configKey"  description:"参数键名" binding:"required" err_msg:"参数键名 必填"`
	ConfigValue string `gorm:"column:config_value" json:"configValue"  description:"参数键值" binding:"required" err_msg:"参数键值 必填"`
	ConfigType  string `gorm:"column:config_type;default:'N'" json:"configType"  description:"系统内置（Y是 N否）"`
	Remark      string `gorm:"column:remark" json:"remark"    description:"备注"`
	BaseNoDelModel
}

func (c SysConfig) TableName() string {
	return "sys_config"
}

func (c SysConfig) Get(query string, args ...interface{}) (SysConfig, error) {
	var data SysConfig
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysConfig) Create(data *SysConfig) error {
	return db.DBConnections["master"].Create(&data).Error
}
func (c SysConfig) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(&SysConfig{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysConfig) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysConfig{})
	return d.Error
}

func (c SysConfig) GetAll(query string, args ...interface{}) ([]SysConfig, error) {
	data := make([]SysConfig, 0)
	d := db.DBConnections["master"].Model(&SysConfig{}).Where(query, args...).Find(&data)
	return data, d.Error
}

func (c SysConfig) SelectConfigList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysConfig, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysConfig{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	var count int64
	data := make([]SysConfig, 0)
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysConfig) SelectConfigAllList(conditions map[string]interface{}) ([]SysConfig, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysConfig{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysConfig, 0)
	d := query.Order("create_time desc").Find(&data)
	return data, d.Error
}
