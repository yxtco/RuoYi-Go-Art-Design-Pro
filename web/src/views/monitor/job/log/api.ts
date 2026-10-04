import type { JobLogQueryParams, SysJobLog } from '@/types/api/monitor/jobLog'
import type { TableDataInfo } from '@/types/api/system/common'

import response from '@utils/http'

// 查询调度日志列表
export function listJobLog(query: JobLogQueryParams & Record<string, any>) {
  return response<TableDataInfo<SysJobLog>>({
    url: '/monitor/jobLog/list',
    method: 'get',
    params: query
  })
}

// 删除调度日志
export function delJobLog(jobLogId: number | number[]) {
  return response({
    url: `/monitor/jobLog/${jobLogId}`,
    method: 'delete'
  })
}

// 清空调度日志
export function cleanJobLog() {
  return response({
    url: '/monitor/jobLog/clean',
    method: 'delete'
  })
}
