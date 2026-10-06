package system

import (
	apisystem "go-fin-server/api/system"
	"go-fin-server/internal/model"
	"go-fin-server/internal/response"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SiteSettingHandler struct {
	Services service.Services
}

func NewSiteSettingHandler() *SiteSettingHandler {
	return &SiteSettingHandler{}
}

// siteSettingKeys 网站设置全部键（存储于 sys_config）
var siteSettingKeys = []struct{ Key, Name string }{
	{"site.name", "网站名称"},
	{"site.logo", "网站Logo"},
	{"site.favicon", "网站favicon"},
	{"site.recordNo", "备案号"},
	{"site.copyright", "版权信息"},
	{"site.description", "网站描述"},
	{"login.title", "登录页标题"},
	{"login.subtitle", "登录页副标题"},
	{"login.background", "登录页背景图"},
	{"login.copyright", "登录页版权"},
	{"sys.account.captchaEnabled", "登录验证码"},
	{"sys.account.registerUser", "开放注册"},
	{"sys.login.blackIPList", "登录IP黑名单"},
	{"sys.log.level", "日志级别"},
	{"sys.log.format", "日志格式"},
}

// loadSiteSetting 从 sys_config 读取全部网站设置
func (s *SiteSettingHandler) loadSiteSetting() (apisystem.SiteSetting, error) {
	var out apisystem.SiteSetting
	vals := make(map[string]string, len(siteSettingKeys))
	for _, k := range siteSettingKeys {
		v, err := s.Services.SysConfigService.SelectConfigByKey(k.Key)
		if err != nil {
			return out, err
		}
		vals[k.Key] = v
	}
	out.SiteName = vals["site.name"]
	out.SiteLogo = vals["site.logo"]
	out.SiteFavicon = vals["site.favicon"]
	out.SiteRecordNo = vals["site.recordNo"]
	out.SiteCopyright = vals["site.copyright"]
	out.SiteDescription = vals["site.description"]
	out.LoginTitle = vals["login.title"]
	out.LoginSubtitle = vals["login.subtitle"]
	out.LoginBackground = vals["login.background"]
	out.LoginCopyright = vals["login.copyright"]
	out.CaptchaEnabled = vals["sys.account.captchaEnabled"]
	out.RegisterUser = vals["sys.account.registerUser"]
	out.BlackIPList = vals["sys.login.blackIPList"]
	out.LogLevel = vals["sys.log.level"]
	out.LogFormat = vals["sys.log.format"]
	return out, nil
}

// saveSetting 保存单个配置键（存在则更新，不存在则新增）
func (s *SiteSettingHandler) saveSetting(key, name, value string) error {
	cfg, err := model.SysConfig{}.Get("config_key = ?", key)
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return s.Services.SysConfigService.InsertConfig(&model.SysConfig{
				ConfigName:  name,
				ConfigKey:   key,
				ConfigValue: value,
				ConfigType:  "N",
			})
		}
		return err
	}
	upd := map[string]interface{}{"config_value": value}
	return s.Services.SysConfigService.UpdateConfig(cfg.Id, key, value, upd)
}

// GetSiteSetting 查询网站设置
//
//	@Summary	查询网站设置
//	@Tags		网站设置
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"网站设置"
//	@Router		/system/siteSetting [get]
//	@Security	BearerAuth
func (s *SiteSettingHandler) GetSiteSetting(c *gin.Context) {
	response.SetOperTitle(c, "查询网站设置")
	data, err := s.loadSiteSetting()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}

// GetPublicSiteSetting 公开查询网站设置（登录页免鉴权读取）
//
//	@Summary	公开查询网站设置
//	@Tags		网站设置
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}"网站设置"
//	@Router		/system/siteSetting/public [get]
func (s *SiteSettingHandler) GetPublicSiteSetting(c *gin.Context) {
	data, err := s.loadSiteSetting()
	if err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	response.Data(c, data)
}

// UpdateSiteSetting 保存网站设置
//
//	@Summary	保存网站设置
//	@Tags		网站设置
//	@Accept		json
//	@Produce	json
//	@Param		body	body	apisystem.SiteSetting	true	"网站设置"
//	@Success	200		{object}	map[string]interface{}"操作结果"
//	@Router		/system/siteSetting [put]
//	@Security	BearerAuth
func (s *SiteSettingHandler) UpdateSiteSetting(c *gin.Context) {
	response.SetOperTitle(c, "保存网站设置")
	var req apisystem.SiteSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Logger.Error(err)
		response.Error(c, err.Error())
		return
	}
	pairs := []struct{ Key, Name, Value string }{
		{"site.name", "网站名称", req.SiteName},
		{"site.logo", "网站Logo", req.SiteLogo},
		{"site.favicon", "网站favicon", req.SiteFavicon},
		{"site.recordNo", "备案号", req.SiteRecordNo},
		{"site.copyright", "版权信息", req.SiteCopyright},
		{"site.description", "网站描述", req.SiteDescription},
		{"login.title", "登录页标题", req.LoginTitle},
		{"login.subtitle", "登录页副标题", req.LoginSubtitle},
		{"login.background", "登录页背景图", req.LoginBackground},
		{"login.copyright", "登录页版权", req.LoginCopyright},
		{"sys.account.captchaEnabled", "登录验证码", req.CaptchaEnabled},
		{"sys.account.registerUser", "开放注册", req.RegisterUser},
		{"sys.login.blackIPList", "登录IP黑名单", req.BlackIPList},
		{"sys.log.level", "日志级别", req.LogLevel},
		{"sys.log.format", "日志格式", req.LogFormat},
	}
	for _, p := range pairs {
		if err := s.saveSetting(p.Key, p.Name, p.Value); err != nil {
			pkg.Logger.Error(err)
			response.Error(c, err.Error())
			return
		}
	}
	// 日志级别变更后立即动态生效（无需重启服务）
	if req.LogLevel != "" {
		pkg.SetLogLevel(req.LogLevel)
		pkg.Logger.Infof("日志级别已动态调整为: %s", req.LogLevel)
	}
	response.DataMsg(c, true, "保存成功")
}
