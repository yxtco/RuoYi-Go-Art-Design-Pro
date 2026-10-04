package model

import (
	"go-fin-server/internal/db"
	"go-fin-server/pkg/utils"
)

type SysDictData struct {
	DictCode  uint64 `gorm:"column:dict_code;primary_key;" json:"dictCode"        description:"字典编码"`
	DictSort  *uint  `gorm:"column:dict_sort" json:"dictSort"  description:"字典排序" binding:"required" err_msg:"显示排序 必填"`
	DictLabel string `gorm:"column:dict_label" json:"dictLabel" description:"字典标签" binding:"required" err_msg:"数据标签 必填"`
	DictValue string `gorm:"column:dict_value" json:"dictValue" description:"字典键值" binding:"required" err_msg:"数据键值 必填"`
	DictType  string `gorm:"column:dict_type" json:"dictType"  description:"字典类型" binding:"required" err_msg:"字典类型 必填"`
	CssClass  string `gorm:"column:css_class" json:"cssClass"  description:"样式属性（其他样式扩展）"`
	ListClass string `gorm:"column:list_class" json:"listClass"  description:"表格回显样式"`
	IsDefault string `gorm:"column:is_default;default:'N'" json:"isDefault"  description:"是否默认（Y是 N否）"`
	Status    string `gorm:"column:status;default:'0'" json:"status"    description:"状态（0正常 1停用）"`
	Remark    string `gorm:"column:remark" json:"remark"    description:"备注"`
	BaseNoDelModel
}

func (c SysDictData) TableName() string {
	return "sys_dict_data"
}

func (c SysDictData) Get(query string, args ...interface{}) (SysDictData, error) {
	var data SysDictData
	d := db.DBConnections["master"].Where(query, args...).First(&data)
	return data, d.Error
}

func (c SysDictData) Create(data *SysDictData) error {
	return db.DBConnections["master"].Create(&data).Error
}

func (c SysDictData) UpdateMap(upd map[string]interface{}, query string, args ...interface{}) error {
	upd["update_time"] = utils.GetCurrentDateTime()
	d := db.DBConnections["master"].Model(SysDictData{}).Where(query, args...).Updates(upd)
	return d.Error
}

func (c SysDictData) Delete(query string, args ...interface{}) error {
	d := db.DBConnections["master"].Where(query, args...).Delete(&SysDictData{})
	return d.Error
}

func (c SysDictData) GetAll(query string, args ...interface{}) ([]SysDictData, error) {
	data := make([]SysDictData, 0)
	d := db.DBConnections["master"].Model(&SysDictData{}).Where(query, args...).Find(&data)
	return data, d.Error
}

func (c SysDictData) SelectDictDataList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]SysDictData, int64, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysDictData{})
	offset, limit := utils.PageParse(pageNum, pageSize)
	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	var count int64
	data := make([]SysDictData, 0)
	if err := query.Count(&count).Error; err != nil {
		return data, count, err
	}
	for _, condition := range sortConditions {
		query = query.Order(condition)
	}
	d := query.Offset(offset).Limit(limit).Find(&data)
	return data, count, d.Error
}

func (c SysDictData) SelectDictDataAllList(conditions map[string]interface{}) ([]SysDictData, error) {
	// 构建查询
	query := db.DBConnections["master"].Model(&SysDictData{})

	// 添加条件
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	data := make([]SysDictData, 0)
	d := query.Order("dict_sort").Find(&data)
	return data, d.Error
}

func (c SysDictData) CountDictDataByType(dictType string) (int64, error) {
	sql := `
		        select count(1) from sys_dict_data where dict_type= ?  
	`
	var count int64
	d := db.DBConnections["master"].Raw(sql, dictType).Scan(&count)
	return count, d.Error
}

type SysDictDataSimplify struct {
	DictLabel string `gorm:"column:dict_label" json:"label" description:"字典标签"`
	DictValue string `gorm:"column:dict_value" json:"value" description:"字典键值"`
	ListClass string `gorm:"column:list_class" json:"listClass"  description:"表格回显样式"`
	CssClass  string `gorm:"column:css_class" json:"cssClass"  description:"样式属性（其他样式扩展）"`
}

func (c SysDictDataSimplify) TableName() string {
	return "sys_dict_data"
}

func (c SysDictDataSimplify) GetSimplify(query string, args ...interface{}) ([]SysDictDataSimplify, error) {
	data := make([]SysDictDataSimplify, 0)
	d := db.DBConnections["master"].Select("dict_label", "dict_value", "list_class").Model(&SysDictDataSimplify{}).Order("dict_sort asc").Where(query, args...).Find(&data)
	return data, d.Error
}
