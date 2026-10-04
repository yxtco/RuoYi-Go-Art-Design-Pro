import type {
  OperlogQueryParams,
  SysOperLog
} from '@/types/api/monitor/operlog'
import type { TableDataInfo } from '@/types/api/system/common'

import response from '@utils/http'

// 查询操作日志列表
export function list(query: OperlogQueryParams) {
  return response<TableDataInfo<SysOperLog>>({
    url: '/monitor/operLog/list',
    method: 'get',
    params: query
  })
}

// 删除操作日志
export function delOperlog(operId: number | number[]) {
  return response({
    url: `/monitor/operLog/${operId}`,
    method: 'delete'
  })
}

// 清空操作日志
export function cleanOperlog() {
  return response({
    url: '/monitor/operLog/clean',
    method: 'delete'
  })
}
