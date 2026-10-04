import type { SysNotice } from '@/types/api/system/notice'
import type { TableDataInfo } from '@/types/api/system/common'

import response from '@utils/http'

// 首页顶部公告列表（复用 list 接口，取最新 5 条正常状态公告）
export async function listNoticeTop(): Promise<{ data: SysNotice[]; unreadCount: number }> {
  const [res, unreadRes] = await Promise.all([
    response<TableDataInfo<SysNotice>>({
      url: '/system/notice/list',
      method: 'get',
      params: { status: '0', pageNum: 1, pageSize: 5 }
    }),
    response<{ unreadCount: number }>({
      url: '/system/notice/unreadCount',
      method: 'get'
    })
  ])
  const list = res.rows || []
  return {
    data: list,
    unreadCount: unreadRes.unreadCount ?? 0
  }
}
