import { defineStore } from 'pinia'

import {
  fetchPublicSiteSetting,
  updateSiteSetting as apiUpdateSiteSetting
} from '@/api/system/siteSetting'
import type { SiteSetting } from '@/types/api/system/siteSetting'

/**
 * 网站设置 Store
 * - 网站名称 / Logo 用于登录页、侧边栏、顶部、路由标题、水印
 * - fetchSite() 在应用启动时调用（登录页未登录也可通过公开接口读取）
 */
export const useSiteStore = defineStore('site', {
  state: () => ({
    name: 'Art Design Pro',
    logo: '',
    favicon: '',
    recordNo: '',
    copyright: '',
    description: '',
    loginTitle: '',
    loginSubtitle: '',
    loginBackground: '',
    loginCopyright: '',
    captchaEnabled: 'true',
    registerUser: 'false',
    blackIPList: '',
    logLevel: 'standard',
    logFormat: 'json',
    loaded: false
  }),
  actions: {
    /** 应用启动时读取公开配置（登录页未登录也可用） */
    async fetchSite() {
      if (this.loaded) return
      try {
        const data = await fetchPublicSiteSetting()
        if (data) this.apply(data)
      } catch {
        // 网络失败时使用默认值，不阻塞页面
      }
    },
    /** 应用本地站点配置 */
    apply(data: SiteSetting) {
      if (data.siteName !== undefined) this.name = data.siteName
      if (data.siteLogo !== undefined) this.logo = data.siteLogo
      if (data.siteFavicon !== undefined) this.favicon = data.siteFavicon
      if (data.siteRecordNo !== undefined) this.recordNo = data.siteRecordNo
      if (data.siteCopyright !== undefined) this.copyright = data.siteCopyright
      if (data.siteDescription !== undefined) this.description = data.siteDescription
      if (data.loginTitle !== undefined) this.loginTitle = data.loginTitle
      if (data.loginSubtitle !== undefined) this.loginSubtitle = data.loginSubtitle
      if (data.loginBackground !== undefined) this.loginBackground = data.loginBackground
      if (data.loginCopyright !== undefined) this.loginCopyright = data.loginCopyright
      if (data.captchaEnabled !== undefined) this.captchaEnabled = data.captchaEnabled
      if (data.registerUser !== undefined) this.registerUser = data.registerUser
      if (data.blackIPList !== undefined) this.blackIPList = data.blackIPList
      if (data.logLevel !== undefined) this.logLevel = data.logLevel
      if (data.logFormat !== undefined) this.logFormat = data.logFormat
      this.loaded = true
    },
    /** 保存网站设置（管理页） */
    async saveSite(data: SiteSetting) {
      await apiUpdateSiteSetting(data)
      this.apply(data)
    }
  }
})
