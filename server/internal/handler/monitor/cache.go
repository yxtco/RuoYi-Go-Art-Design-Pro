package monitor

import (
	"go-fin-server/internal/constant"
	"go-fin-server/internal/db"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/pkg"
	"go-fin-server/pkg/redistool"
	"go-fin-server/pkg/utils/stringutils"
	"strings"

	"github.com/gin-gonic/gin"
)

type CacheHandler struct {
}

func NewCacheHandler() *CacheHandler {
	return &CacheHandler{}
}

// GetCache 查询缓存详细
//
//	@Summary	查询缓存详细
//	@Tags		缓存监控
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"缓存信息"
//	@Router		/monitor/cache [get]
//	@Security	BearerAuth
func (s *CacheHandler) GetCache(c *gin.Context) {
	response.SetOperTitle(c, "查询缓存信息")
	client := db.RedisConnections["master"]
	info, err := redistool.Info(client)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	commandStats, err := redistool.CommandStats(client)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	dbSize, err := redistool.DBSize(client)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, map[string]interface{}{
		"info":         parseInfo(info),
		"dbSize":       dbSize,
		"commandStats": parseCommandStats(commandStats),
	})
}
// ListCacheName 查询缓存名称列表
//
//	@Summary	查询缓存名称列表
//	@Tags		缓存监控
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"缓存名称列表"
//	@Router		/monitor/cache/getNames [get]
//	@Security	BearerAuth
func (s *CacheHandler) ListCacheName(c *gin.Context) {
	response.SetOperTitle(c, "查询缓存名称列表")
	caches := []model.SysCache{
		{
			CacheName: constant.CACHE_LOGIN_TOKEN_KEY,
			Remark:    "用户信息",
		},
		{
			CacheName: constant.CACHE_SYS_CONFIG_KEY,
			Remark:    "配置信息",
		},
		{
			CacheName: constant.CACHE_SYS_DICT_KEY,
			Remark:    "数据字典",
		},
		{
			CacheName: constant.CACHE_CAPTCHA_CODE_KEY,
			Remark:    "验证码",
		},
		{
			CacheName: constant.CACHE_REPEAT_SUBMIT_KEY,
			Remark:    "防重提交",
		},
		{
			CacheName: constant.CACHE_RATE_LIMIT_KEY,
			Remark:    "限流处理",
		},
		{
			CacheName: constant.CACHE_PWD_ERR_CNT_KEY,
			Remark:    "密码错误次数",
		},
	}
	response.Data(c, caches)
}
// GetCacheKeys 查询缓存键名列表
//
//	@Summary	查询缓存键名列表
//	@Tags		缓存监控
//	@Produce	json
//	@Param		cacheName	path	string	true	"缓存名称"
//	@Success	200		{object}	map[string]interface{}"缓存键名列表"
//	@Router		/monitor/cache/getKeys/{cacheName} [get]
//	@Security	BearerAuth
func (s *CacheHandler) GetCacheKeys(c *gin.Context) {
	response.SetOperTitle(c, "查询缓存键名列表")
	cacheName := c.Param("cacheName")
	keys, err := redistool.Keys(db.RedisConnections["master"], cacheName+"*")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, keys)
}
// GetCacheValue 查询缓存内容
//
//	@Summary	查询缓存内容
//	@Tags		缓存监控
//	@Produce	json
//	@Param		cacheName	path	string	true	"缓存名称"
//	@Param		cacheKey	path	string	true	"缓存键名"
//	@Success	200		{object}	map[string]interface{}"缓存内容"
//	@Router		/monitor/cache/getValue/{cacheName}/{cacheKey} [get]
//	@Security	BearerAuth
func (s *CacheHandler) GetCacheValue(c *gin.Context) {
	response.SetOperTitle(c, "查询缓存内容")
	cacheName := c.Param("cacheName")
	cacheKey := c.Param("cacheKey")
	cacheValue, err := redistool.Get(db.RedisConnections["master"], cacheKey)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	sysCache := model.SysCache{
		CacheName:  strings.Replace(cacheName, ":", "", -1),
		CacheKey:   strings.Replace(cacheKey, cacheName, "", -1),
		CacheValue: cacheValue,
	}
	response.Data(c, sysCache)
}
// ClearCacheName 清理指定名称缓存
//
//	@Summary	清理指定名称缓存
//	@Tags		缓存监控
//	@Produce	json
//	@Param		cacheName	path	string	true	"缓存名称"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/cache/clearCacheName/{cacheName} [delete]
//	@Security	BearerAuth
func (s *CacheHandler) ClearCacheName(c *gin.Context) {
	response.SetOperTitle(c, "清理指定名称缓存")
	cacheName := c.Param("cacheName")
	keys, err := redistool.Keys(db.RedisConnections["master"], cacheName+"*")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = redistool.Del(db.RedisConnections["master"], keys...)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// ClearCacheKey 清理指定键名缓存
//
//	@Summary	清理指定键名缓存
//	@Tags		缓存监控
//	@Produce	json
//	@Param		cacheKey	path	string	true	"缓存键名"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/cache/clearCacheKey/{cacheKey} [delete]
//	@Security	BearerAuth
func (s *CacheHandler) ClearCacheKey(c *gin.Context) {
	response.SetOperTitle(c, "清理指定键名缓存")
	cacheKey := c.Param("cacheKey")
	err := redistool.Del(db.RedisConnections["master"], cacheKey)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}
// ClearCacheAll 清理全部缓存
//
//	@Summary	清理全部缓存
//	@Tags		缓存监控
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"操作结果"
//	@Router		/monitor/cache/clearCacheAll [delete]
//	@Security	BearerAuth
func (s *CacheHandler) ClearCacheAll(c *gin.Context) {
	response.SetOperTitle(c, "清理全部缓存")
	keys, err := redistool.Keys(db.RedisConnections["master"], "*")
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	err = redistool.Del(db.RedisConnections["master"], keys...)
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, nil)
}

func parseInfo(res string) map[string]string {
	info := make(map[string]string)
	lines := strings.Split(res, "\r\n")
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info[parts[0]] = parts[1]
			}
		}
	}
	return info
}

// parseCommandStats 解析命令统计信息响应，返回包含命令名称和调用次数的映射切片。
func parseCommandStats(res string) []map[string]string {
	pieList := make([]map[string]string, 0)
	lines := strings.Split(res, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cmdstat_") {
			name := stringutils.RemoveStart(line, "cmdstat_")
			name = strings.SplitN(name, ":", 2)[0] // 使用 SplitN 确保只分割第一个冒号
			value := stringutils.SubstringBetween(line, "calls=", ",")
			if value != "" {
				data := map[string]string{"name": name, "value": value}
				pieList = append(pieList, data)
			}
		}
	}
	return pieList
}
