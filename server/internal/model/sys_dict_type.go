package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysDictType struct {
	Id       uint64 `gorm:"column:id;primary_key" json:"id"    description:"ID"`
	DictName string `gorm:"column:dict_name" json:"dictName"  description:"字典名称" binding:"required" err_msg:"字典名称 必填"`
	DictType string `gorm:"column:dict_type" json:"dictType"  description:"字典类型" binding:"required" err_msg:"字典类型 必填"`
	Status   string `gorm:"column:status;default:'0'" json:"status"    description:"状态（0正常 1停用）"`
	Remark   string `gorm:"column:remark" json:"remark"    description:"备注"`
	BaseNoDelModel
}

func (c SysDictType) TableName() string {
	return "sys_dict_type"
}

func (c SysDictType) Get(query string, args ...interface{}) (SysDictType, error) {
	var data SysDictType
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysDictType) Create(data *SysDictType) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysDictType) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(&SysDictType{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysDictType) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysDictType{})
	return d.Error
}

func (c SysDictType) GetAll(query string, args ...interface{}) ([]SysDictType, error) {
	data := make([]SysDictType, 0)
	d := db.DBConnections["master"].Model(&SysDictType{}).Where(query, args...).Find(&data)
	return data, d.Error
}

func (c SysDictType) SelectDictTypeList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysDictType, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysDictType{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	var count int64
	data := make([]SysDictType, 0)
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysDictType) SelectDictTypeAllList(conditions map[string]interface{}) ([]SysDictType, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysDictType{})
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysDictType, 0)
	d := query.Order("create_time desc").Find(&data)
	return data, d.Error
}
