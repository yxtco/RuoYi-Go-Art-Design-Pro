/**
 * HTTP 请求封装模块
 * 基于 Axios 封装的 HTTP 请求工具，提供统一的请求/响应处理
 *
 * ## 主要功能
 *
 * - 请求/响应拦截器（自动添加 Token、统一错误处理）
 * - 401 未授权自动登出（带防抖机制）
 * - 请求失败自动重试（可配置）
 * - 统一的成功/错误消息提示
 * - 支持 GET/POST/PUT/DELETE 等常用方法
 *
 * @module utils/http
 * @author Art Design Pro Team
 */

import axios, {
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig
} from 'axios'

import { $t } from '@/locales'
import { useUserStore } from '@/store/modules/user'
import { BaseResponse } from '@/types'

import { HttpError, handleError, showError, showSuccess } from './error'
import { ApiStatus } from './status'

/** 请求配置常量 */
const REQUEST_TIMEOUT = 15000
const LOGOUT_DELAY = 500
const MAX_RETRIES = 0
const RETRY_DELAY = 1000
const UNAUTHORIZED_DEBOUNCE_TIME = 3000

/** 401防抖状态 */
let isUnauthorizedErrorShown = false
let unauthorizedTimer: NodeJS.Timeout | null = null

/** 扩展 AxiosRequestConfig */
interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
  showErrorMessage?: boolean
  showSuccessMessage?: boolean
}

const { VITE_API_URL, VITE_WITH_CREDENTIALS } = import.meta.env

/** Axios实例 */
export const axiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT,
  baseURL: VITE_API_URL,
  withCredentials: VITE_WITH_CREDENTIALS === 'true',
  validateStatus: (status) => status >= 200 && status < 300,
  transformResponse: [
    (data, headers) => {
      const contentType = headers['content-type']
      if (contentType?.includes('application/json')) {
        try {
          return JSON.parse(data)
        } catch {
          return data
        }
      }
      return data
    }
  ]
})

/** 请求拦截器 */
axiosInstance.interceptors.request.use(
  (request: InternalAxiosRequestConfig) => {
    const { accessToken } = useUserStore()
    if (accessToken) request.headers.set('Authorization', `Bearer ${accessToken}`)

    if (
      request.data &&
      !(request.data instanceof FormData) &&
      !request.headers['Content-Type']
    ) {
      request.headers.set('Content-Type', 'application/json')
      request.data = JSON.stringify(request.data)
    }

    return request
  },
  (error) => {
    showError(
      createHttpError($t('httpMsg.requestConfigError'), ApiStatus.error)
    )
    return Promise.reject(error)
  }
)

/** 响应拦截器 */
axiosInstance.interceptors.response.use(
  (response: AxiosResponse<BaseResponse>) => {
    const { code, msg } = response.data
    if (
      response.request.responseType === 'blob' ||
      response.request.responseType === 'arraybuffer'
    ) {
      return response
    }
    if (code === ApiStatus.success) return response
    if (code === ApiStatus.unauthorized) handleUnauthorizedError(msg)
    throw createHttpError(msg || $t('httpMsg.requestFailed'), code)
  },
  (error) => {
    if (error.response?.status === ApiStatus.unauthorized)
      handleUnauthorizedError()
    return Promise.reject(handleError(error))
  }
)

/** 统一创建HttpError */
function createHttpError(message: string, code: number) {
  return new HttpError(message, code)
}

/** 处理401错误（带防抖） */
function handleUnauthorizedError(message?: string): never {
  const error = createHttpError(
    message || $t('httpMsg.unauthorized'),
    ApiStatus.unauthorized
  )

  if (!isUnauthorizedErrorShown) {
    isUnauthorizedErrorShown = true
    logOut()

    unauthorizedTimer = setTimeout(
      resetUnauthorizedError,
      UNAUTHORIZED_DEBOUNCE_TIME
    )

    showError(error, true)
    throw error
  }

  throw error
}

/** 重置401防抖状态 */
function resetUnauthorizedError() {
  isUnauthorizedErrorShown = false
  if (unauthorizedTimer) clearTimeout(unauthorizedTimer)
  unauthorizedTimer = null
}

/** 退出登录函数 */
function logOut() {
  setTimeout(() => {
    useUserStore().logOut()
  }, LOGOUT_DELAY)
}

/** 是否需要重试 */
function shouldRetry(statusCode: number) {
  return [
    ApiStatus.requestTimeout,
    ApiStatus.internalServerError,
    ApiStatus.badGateway,
    ApiStatus.serviceUnavailable,
    ApiStatus.gatewayTimeout
  ].includes(statusCode)
}

/** 请求重试逻辑 */
async function retryRequest<T>(
  config: ExtendedAxiosRequestConfig,
  retries: number = MAX_RETRIES
): Promise<T> {
  try {
    return await request<T>(config)
  } catch (error) {
    if (retries > 0 && error instanceof HttpError && shouldRetry(error.code)) {
      await delay(RETRY_DELAY)
      return retryRequest<T>(config, retries - 1)
    }
    throw error
  }
}

/** 延迟函数 */
function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * 后端所有模型统一使用 `id` 作为主键 JSON 字段名，
 * 但前端各实体类型使用不同的 ID 字段名（如 userId、roleId 等）。
 * 此函数根据请求 URL 自动将 `id` 复制为对应实体字段名，以兼容前端代码。
 */
function adaptIdField(data: any, url: string): any {
  if (!data || typeof data !== 'object') return data

  const idField = getIdFieldByUrl(url)
  if (!idField) return data

  if (Array.isArray(data)) {
    return data.map((item) => adaptIdField(item, url))
  }

  if (data.id !== undefined) {
    data[idField] = data.id
  }

  return data
}

function getIdFieldByUrl(url: string): string | null {
  if (!url) return null
  const u = url.toLowerCase()
  if (u.includes('/user')) return 'userId'
  if (u.includes('/role')) return 'roleId'
  if (u.includes('/dept')) return 'deptId'
  if (u.includes('/post')) return 'postId'
  if (u.includes('/menu')) return 'menuId'
  if (u.includes('/config')) return 'configId'
  if (u.includes('/notice')) return 'noticeId'
  if (u.includes('/dict/data')) return 'dictCode'
  if (u.includes('/dict/type')) return 'dictId'
  if (u.includes('/joblog') || u.includes('/job/log')) return 'jobLogId'
  if (u.includes('/job')) return 'jobId'
  if (u.includes('/logininfor') || u.includes('/logininfo')) return 'infoId'
  if (u.includes('/operlog') || u.includes('/operLog')) return 'operId'
  if (u.includes('/online')) return 'tokenId'
  return null
}

/** 请求函数 */
async function request<T = any>(
  config: ExtendedAxiosRequestConfig
): Promise<T> {
  try {
    const res = await axiosInstance.request<BaseResponse<T>>(config)

    // 显示成功消息
    if (config.showSuccessMessage && res.data.msg) {
      showSuccess(res.data.msg)
    }

    // ruoyi没有采用严格的{code, msg, data}模式
    const { code: _code, msg, data, ...other } = res.data as any

    // 还有直接返回msg的情况
    if (Object.keys(res.data).length === 2) {
      return msg as T
    }

    // 只包含data的情况
    if (data && Object.keys(res.data).length === 3) {
      return adaptIdField(data, config.url || '') as T
    }
    // 包含data 还有其它参数（如分页的 total、rows）
    if (data) {
      return adaptIdField({ data, ...other }, config.url || '') as T
    }
    // 需要考虑data为null的情况(比如查询为空)
    if (data !== undefined) {
      // data为null但存在额外字段时(如roleMenuTreeSelect返回的checkedKeys/menus)，优先返回额外字段
      if (data === null && Object.keys(other).length > 0) {
        return other as T
      }
      return adaptIdField(data, config.url || '') as T
    } else {
      // 分页数据在 other 中（rows + total）
      const result = other as T
      if (other.rows && Array.isArray(other.rows)) {
        other.rows = other.rows.map((item: any) =>
          adaptIdField(item, config.url || '')
        )
      }
      return result
    }
  } catch (error) {
    if (error instanceof HttpError && error.code !== ApiStatus.unauthorized) {
      const showMsg = config.showErrorMessage !== false
      showError(error, showMsg)
    }
    return Promise.reject(error)
  }
}

/** API方法集合 */
// const api = {
//   get<T>(config: ExtendedAxiosRequestConfig) {
//     return retryRequest<T>({ ...config, method: 'GET' })
//   },
//   post<T>(config: ExtendedAxiosRequestConfig) {
//     return retryRequest<T>({ ...config, method: 'POST' })
//   },
//   put<T>(config: ExtendedAxiosRequestConfig) {
//     return retryRequest<T>({ ...config, method: 'PUT' })
//   },
//   del<T>(config: ExtendedAxiosRequestConfig) {
//     return retryRequest<T>({ ...config, method: 'DELETE' })
//   },
//   request<T>(config: ExtendedAxiosRequestConfig) {
//     return retryRequest<T>(config)
//   }
// }

export default retryRequest
