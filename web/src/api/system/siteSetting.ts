import request from '@utils/http'
import type { SiteSetting, UploadResult } from '@/types/api/system/siteSetting'

// 查询网站设置（需权限 system:site:query）
export function fetchSiteSetting() {
  return request<SiteSetting>({
    url: '/system/siteSetting',
    method: 'get'
  })
}

// 公开查询网站设置（登录页未登录时读取）
export function fetchPublicSiteSetting() {
  return request<SiteSetting>({
    url: '/system/siteSetting/public',
    method: 'get'
  })
}

// 保存网站设置（需权限 system:site:edit）
export function updateSiteSetting(data: SiteSetting) {
  return request<boolean>({
    url: '/system/siteSetting',
    method: 'put',
    data,
    showSuccessMessage: true
  })
}

// 上传网站 Logo / 背景图（通用文件上传）
export function uploadSiteFile(file: File) {
  const form = new FormData()
  form.append('file', file)
  return request<UploadResult>({
    url: '/common/upload',
    method: 'post',
    headers: { 'Content-Type': 'multipart/form-data' },
    data: form
  })
}
