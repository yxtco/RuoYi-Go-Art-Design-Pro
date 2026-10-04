package model

type SysCache struct {
	CacheName  string `json:"cacheName"    description:"缓存名称"`
	CacheKey   string `json:"cacheKey"    description:"缓存键名"`
	CacheValue string `json:"cacheValue"    description:"缓存内容"`
	Remark     string `json:"remark"    description:"备注"`
}
