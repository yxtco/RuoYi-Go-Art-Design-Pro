package system

import (
	"errors"
	"fmt"
	"go-fin-server/internal/constant"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/pkg"
	"go-fin-server/pkg/redistool"

	"gorm.io/gorm"
)

type SysConfigService struct {
}

// SelectConfigList 查询参数配置列表
func (c SysConfigService) SelectConfigList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.SysConfig, int64, error) {
	data, count, err := model.SysConfig{}.SelectConfigList(conditions, sortConditions, pageNum, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return data, count, nil
}

// SelectConfigAllList 查询参数配置列表
func (c SysConfigService) SelectConfigAllList(conditions map[string]interface{}) ([]model.SysConfig, error) {
	data, err := model.SysConfig{}.SelectConfigAllList(conditions)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SelectConfigById 查询参数配置信息
func (c SysConfigService) SelectConfigById(configId uint64) (model.SysConfig, error) {
	data, err := model.SysConfig{}.Get("id = ?", configId)
	if err != nil {
		return data, err
	}
	return data, nil
}

// SelectConfigByKey 根据键名查询参数配置信息
func (c SysConfigService) SelectConfigByKey(configKey string) (string, error) {
	client := db.RedisConnections["master"]
	configValue, err := redistool.Get(client, c.GetCacheKey(configKey))
	if err == nil && configValue != "" {
		return configValue, nil
	}

	data, err := model.SysConfig{}.Get("config_key = ?", configKey)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return "", nil
		}
		return "", err
	}

	_ = redistool.Set(client, c.GetCacheKey(configKey), data.ConfigValue, 0)
	return data.ConfigValue, nil
}

// CheckConfigKeyUnique 校验参数键名是否唯一
func (c SysConfigService) CheckConfigKeyUnique(configId uint64, configKey string) bool {
	var Config model.SysConfig
	configObj, err := Config.Get("config_key = ?", configKey)
	if err != nil {
		if err.Error() != gorm.ErrRecordNotFound.Error() {
			pkg.Logger.Error(err.Error())
			return false
		}
		return true
	}
	if configObj.Id != configId {
		return false
	}
	return true
}

// InsertConfig 新增参数配置
func (c SysConfigService) InsertConfig(data *model.SysConfig) error {
	err := data.Create(data)
	if err != nil {
		return err
	}
	err = redistool.Set(db.RedisConnections["master"], c.GetCacheKey(data.ConfigKey), data.ConfigValue, 0)
	if err != nil {
		return err
	}
	return nil
}

// UpdateConfig 修改参数配置
func (c SysConfigService) UpdateConfig(configId uint64, ConfigKey, ConfigValue string, upd map[string]interface{}) error {
	client := db.RedisConnections["master"]
	temp, err := model.SysConfig{}.Get("id = ?", configId)
	if err != nil {
		return err
	}
	if temp.ConfigKey != ConfigKey {
		err = redistool.Del(client, c.GetCacheKey(temp.ConfigKey))
		if err != nil {
			return err
		}
	}
	err = model.SysConfig{}.UpdateMap(upd, "id = ?", configId)
	if err != nil {
		return err
	}
	err = redistool.Set(client, c.GetCacheKey(ConfigKey), ConfigValue, 0)
	if err != nil {
		return err
	}
	return nil
}

// DeleteConfigByIds 批量删除参数信息
func (c SysConfigService) DeleteConfigByIds(ids []uint64) error {
	for _, id := range ids {
		data, err := model.SysConfig{}.Get("id = ?", id)
		if err != nil {
			return err
		}
		if "Y" == data.ConfigType {
			return errors.New(fmt.Sprintf("内置参数【%s】,不能删除", data.ConfigKey))
		}
		err = model.SysConfig{}.Delete("id = ？", id)
		if err != nil {
			return err
		}
		err = redistool.Del(db.RedisConnections["master"], c.GetCacheKey(data.ConfigKey))
		if err != nil {
			return err
		}
	}

	return nil
}

// ResetConfigCache 重置参数缓存数据
func (c SysConfigService) ResetConfigCache() error {
	err := c.ClearConfigCache()
	if err != nil {
		return err
	}
	err = c.LoadingConfigCache()
	if err != nil {
		return err
	}
	return nil
}

// ClearConfigCache 清空参数缓存数据
func (c SysConfigService) ClearConfigCache() error {
	keys, err := redistool.Keys(db.RedisConnections["master"], c.GetCacheKey("*"))
	if err != nil {
		return err
	}
	err = redistool.Del(db.RedisConnections["master"], keys...)
	if err != nil {
		return err
	}
	return nil
}

// LoadingConfigCache 加载参数缓存数据
func (c SysConfigService) LoadingConfigCache() error {
	data, err := model.SysConfig{}.GetAll("")
	if err != nil {
		return err
	}
	for _, item := range data {
		err := redistool.Set(db.RedisConnections["master"], c.GetCacheKey(item.ConfigKey), item.ConfigValue, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCacheKey 设置cache key
func (c SysConfigService) GetCacheKey(key string) string {
	return constant.CACHE_SYS_CONFIG_KEY + key
}
