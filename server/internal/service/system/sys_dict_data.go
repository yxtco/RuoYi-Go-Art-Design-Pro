package system

import (
	"go-fin-server/internal/constant"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"

	"gorm.io/gorm"
)

type SysDictDataService struct {
}

// SelectDictDataList 根据条件分页查询字典数据
func (c SysDictDataService) SelectDictDataList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysDictData, int64, error) {
	data, count, err := model.SysDictData{}.SelectDictDataList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// SelectDictDataAllList 根据条件查询字典数据
func (c SysDictDataService) SelectDictDataAllList(conditions map[string]interface{}) ([]model.SysDictData, error) {
	data, err := model.SysDictData{}.SelectDictDataAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectDictDataById 根据字典类型ID查询信息
func (c SysDictDataService) SelectDictDataById(id uint64) (model.SysDictData, error) {
	data, err := model.SysDictData{}.Get("dict_code = ?", id)
	if err != nil {
		return data, err
	}
	return data, nil
}

// CheckDictValueUnique 校验字典键值是否唯一
func (c SysDictDataService) CheckDictValueUnique(id uint64, dictValue string) bool {
	var data model.SysDictData
	dataObj, err := data.Get("dict_value = ?", dictValue)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if dataObj.DictCode != id {
		return false
	}
	return true
}

// InsertDictData 新增保存字典数据信息
func (c SysDictDataService) InsertDictData(data *model.SysDictData) error {
	err := data.Create(data)
	if err != nil {
		return err
	}
	dictDatas, err := model.SysDictData{}.GetAll("dict_type = ?", data.DictType)
	if err != nil {
		return err
	}
	err = SysDictTypeService{}.SetDictCache(c.GetCacheKey(data.DictType), dictDatas)
	if err != nil {
		return err
	}
	return nil
}

// UpdateDictData 新增保存字典数据信息
func (c SysDictDataService) UpdateDictData(dictCode uint64, dictType string, upd map[string]interface{}) error {
	err := model.SysDictData{}.UpdateMap(upd, "dict_code = ?", dictCode)
	if err != nil {
		return err
	}
	dictDatas, err := model.SysDictData{}.GetAll("dict_type = ?", dictType)
	if err != nil {
		return err
	}
	err = SysDictTypeService{}.SetDictCache(c.GetCacheKey(dictType), dictDatas)
	if err != nil {
		return err
	}
	return nil
}

// DeleteDictDataByIds 批量删除字典数据信息
func (c SysDictDataService) DeleteDictDataByIds(ids []uint64) error {
	for _, id := range ids {
		data, err := model.SysDictData{}.Get("dict_code = ?", id)
		if err != nil {
			return err
		}
		err = model.SysDictData{}.Delete("dict_code = ?", id)
		if err != nil {
			return err
		}
		dictDatas, err := model.SysDictData{}.GetAll("dict_type = ?", data.DictType)
		if err != nil {
			return err
		}
		err = SysDictTypeService{}.SetDictCache(c.GetCacheKey(data.DictType), dictDatas)
		if err != nil {
			return err
		}
		return nil

	}
	return nil
}

// GetCacheKey 设置cache key
func (c SysDictDataService) GetCacheKey(key string) string {
	return constant.CACHE_SYS_DICT_KEY + key
}
