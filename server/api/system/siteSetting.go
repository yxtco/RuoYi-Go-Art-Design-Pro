package system

// SiteSetting 网站设置（基础设置 + 登录页设置 + 登录安全）
//
// 所有字段均持久化到 sys_config（key-value），键前缀 site.* / login.* / sys.account.* / sys.login.*。
type SiteSetting struct {
	SiteName        string `json:"siteName"`
	SiteLogo        string `json:"siteLogo"`
	SiteFavicon     string `json:"siteFavicon"`
	SiteRecordNo    string `json:"siteRecordNo"`
	SiteCopyright   string `json:"siteCopyright"`
	SiteDescription string `json:"siteDescription"`
	LoginTitle      string `json:"loginTitle"`
	LoginSubtitle   string `json:"loginSubtitle"`
	LoginBackground string `json:"loginBackground"`
	LoginCopyright  string `json:"loginCopyright"`
	// 登录安全（对应 sys_config 参数，值 "true"/"false" 或文本）
	CaptchaEnabled string `json:"captchaEnabled"` // sys.account.captchaEnabled 登录验证码开关
	RegisterUser   string `json:"registerUser"`   // sys.account.registerUser 是否开放注册
	BlackIPList    string `json:"blackIPList"`    // sys.login.blackIPList 登录IP黑名单
}
