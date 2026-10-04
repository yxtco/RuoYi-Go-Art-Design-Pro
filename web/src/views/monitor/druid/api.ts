import response from '@utils/http'

/** SQL 监控统计数据 */
export interface SqlStats {
  totalCount: number
  totalDuration: number
  avgDuration: number
  slowCount: number
  selectCount: number
  insertCount: number
  updateCount: number
  deleteCount: number
  selectDuration: number
  insertDuration: number
  updateDuration: number
  deleteDuration: number
  recent: SqlRecord[]
  slow: SqlRecord[]
}

/** SQL 执行记录 */
export interface SqlRecord {
  time: string
  type: string
  duration: number
  rows: number
  sql: string
}

/** 获取 SQL 监控统计数据 */
export function getSQLStats() {
  return response<SqlStats>({
    url: '/monitor/druid/sqlStats',
    method: 'get'
  })
}

/** 清空 SQL 监控统计数据 */
export function clearSQLStats() {
  return response({
    url: '/monitor/druid/sqlStats',
    method: 'delete'
  })
}
