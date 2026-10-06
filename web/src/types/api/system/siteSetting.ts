// 网站设置（基础设置 + 登录页设置），持久化到后端 sys_config
export interface SiteSetting {
  /** 网站名称 */
  siteName?: string
  /** 网站 Logo 图片地址 */
  siteLogo?: string
  /** 网站 favicon 图片地址 */
  siteFavicon?: string
  /** ICP 备案号 */
  siteRecordNo?: string
  /** 底部版权信息 */
  siteCopyright?: string
  /** 网站描述 */
  siteDescription?: string
  /** 登录页左侧标题 */
  loginTitle?: string
  /** 登录页左侧副标题 */
  loginSubtitle?: string
  /** 登录页背景图地址 */
  loginBackground?: string
  /** 登录页版权文字 */
  loginCopyright?: string
  /** 登录验证码开关（"true"/"false"，对应 sys.account.captchaEnabled） */
  captchaEnabled?: string
  /** 是否开放注册（"true"/"false"，对应 sys.account.registerUser） */
  registerUser?: string
  /** 登录IP黑名单（;分隔，对应 sys.login.blackIPList） */
  blackIPList?: string
  /** 日志级别：quiet(安静) / standard(标准) / detailed(详细)，对应 sys.log.level */
  logLevel?: string
  /** 日志格式：json / console，对应 sys.log.format */
  logFormat?: string
}

// 通用文件上传结果（/common/upload）
export interface UploadResult {
  fileName?: string
  newFileName?: string
  url?: string
  originalFilename?: string
}
