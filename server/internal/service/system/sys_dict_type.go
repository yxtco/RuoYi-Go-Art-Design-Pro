package system

import (
	"errors"
	"fmt"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/redistool"

	"gorm.io/gorm"
)

type SysDictTypeService struct {
}

// SelectDictTypeList 根据条件分页查询字典类型
func (c SysDictTypeService) SelectDictTypeList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysDictType, int64, error) {
	data, count, err := model.SysDictType{}.SelectDictTypeList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// SelectDictTypeAllList 根据条件查询字典类型
func (c SysDictTypeService) SelectDictTypeAllList(conditions map[string]interface{}) ([]model.SysDictType, error) {
	data, err := model.SysDictType{}.SelectDictTypeAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectDictTypeById 根据字典类型ID查询信息
func (c SysDictTypeService) SelectDictTypeById(id uint64) (model.SysDictType, error) {
	data, err := model.SysDictType{}.Get("id = ?", id)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CheckDictTypeUnique 校验字典类型称是否唯一
func (c SysDictTypeService) CheckDictTypeUnique(id uint64, dictType string) bool {
	var data model.SysDictType
	dataObj, err := data.Get("dict_type = ?", dictType)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if dataObj.Id != id {
		return false
	}
	return true
}

// InsertDictType 新增保存字典类型信息
func (c SysDictTypeService) InsertDictType(data *model.SysDictType) error {
	err := data.Create(data)
	if err != nil {
		return err
	}
	err = c.SetDictCache(fmt.Sprintf("sys_dict:%s", data.DictType), make([]model.SysDictData, 0))
	if err != nil {
		return err
	}
	return nil
}

// UpdateDictType 修改保存字典类型信息
func (c SysDictTypeService) UpdateDictType(dictTypeId uint64, dictType string, upd map[string]interface{}) error {
	oldDict, err := c.SelectDictTypeById(dictTypeId)
	if err != nil {
		return err
	}
	err = model.SysDictData{}.UpdateMap(map[string]interface{}{
		"dict_type": dictType,
	}, "dict_type = ?", oldDict.DictType)
	if err != nil {
		return err
	}
	err = model.SysDictType{}.UpdateMap(upd, "id = ?", dictTypeId)
	if err != nil {
		return err
	}

	dictDatas, err := model.SysDictData{}.GetAll("dict_type = ?", dictType)
	if err != nil {
		return err
	}
	err = c.SetDictCache(fmt.Sprintf("sys_dict:%s", dictType), dictDatas)
	if err != nil {
		return err
	}
	return nil
}

// DeleteDictTypeByIds 批量删除字典类型信息
func (c SysDictTypeService) DeleteDictTypeByIds(ids []uint64) error {
	for _, id := range ids {
		dictType, err := model.SysDictType{}.Get("id = ?", id)
		if err != nil {
			return err
		}
		count, err := model.SysDictData{}.CountDictDataByType(dictType.DictType)
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New(fmt.Sprintf("%s已分配,不能删除", dictType.DictName))
		}
		err = model.SysDictType{}.Delete("id = ?", id)
		if err != nil {
			return err
		}
		err = c.RemoveDictCache(fmt.Sprintf("sys_dict:%s", dictType.DictType))
		if err != nil {
			return err
		}

	}
	return nil
}

// SelectDictTypeAll 根据所有字典类型
func (c SysDictTypeService) SelectDictTypeAll() ([]model.SysDictType, error) {
	data, err := model.SysDictType{}.GetAll("")
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ResetDictCache 重置字典缓存数据
func (c SysDictTypeService) ResetDictCache() error {
	err := c.ClearDictCache()
	if err != nil {
		return err
	}
	err = c.LoadingDictCache()
	if err != nil {
		return err
	}
	return nil
}

// SetDictCache 设置字典缓存
func (c SysDictTypeService) SetDictCache(key string, dictDatas []model.SysDictData) error {
	err := redistool.SetCacheObject(db.RedisConnections["master"], key, dictDatas, 0)
	if err != nil {
		return err
	}
	return nil
}

// GetDictCache 获取字典缓存
func (c SysDictTypeService) GetDictCache(key string) ([]model.SysDictData, error) {
	data := make([]model.SysDictData, 0)
	err := redistool.GetCacheObject(db.RedisConnections["master"], key, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// RemoveDictCache 删除指定字典缓存
func (c SysDictTypeService) RemoveDictCache(key string) error {
	err := redistool.Del(db.RedisConnections["master"], key)
	if err != nil {
		return err
	}
	return nil
}

// ClearDictCache 清空字典缓存数据
func (c SysDictTypeService) ClearDictCache() error {
	keys, err := redistool.Keys(db.RedisConnections["master"], "sys_dict:*")
	if err != nil {
		return err
	}
	err = redistool.Del(db.RedisConnections["master"], keys...)
	if err != nil {
		return err
	}
	return nil
}

// LoadingDictCache 加载字典缓存数据
func (c SysDictTypeService) LoadingDictCache() error {
	data, err := model.SysDictData{}.GetAll("status = ?", "0")
	if err != nil {
		return err
	}
	dictDataMap := make(map[string][]model.SysDictData)

	for _, item := range data {
		if _, ok := dictDataMap[item.DictType]; !ok {
			dictDataMap[item.DictType] = make([]model.SysDictData, 0)
		}
		dictDataMap[item.DictType] = append(dictDataMap[item.DictType], item)
	}

	for key, list := range dictDataMap {
		err := c.SetDictCache(fmt.Sprintf("sys_dict:%s", key), list)
		if err != nil {
			return err
		}
	}
	return nil
}
