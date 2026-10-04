/**
 * 聊天相关 API 接口
 */

import request from '@utils/http'
import type { OnlineUser } from '@utils/websocket'

/** 历史聊天消息（数据库持久化记录） */
export interface ChatHistoryRecord {
  id: number
  fromUserId: number
  fromName: string
  fromAvatar: string
  toUserId: number
  groupName?: string
  content: string
  msgType: 'text' | 'image' | 'file' | 'emoji'
  fileUrl?: string
  fileName?: string
  timestamp: number
}

/**
 * 获取在线用户列表
 */
export function fetchOnlineUsers() {
  return request<OnlineUser[]>({
    url: '/chat/online',
    method: 'get'
  })
}

/**
 * 获取在线用户数量
 */
export function fetchOnlineCount() {
  return request<number>({
    url: '/chat/online/count',
    method: 'get'
  })
}

/**
 * 获取私聊历史消息
 * @param otherId 对方用户ID
 */
export function fetchPrivateHistory(otherId: number, limit = 100) {
  return request<ChatHistoryRecord[]>({
    url: '/chat/history',
    method: 'get',
    params: { otherId, limit }
  })
}

/**
 * 获取群聊历史消息
 * @param groupName 群聊名称
 */
export function fetchGroupHistory(groupName: string, limit = 100) {
  return request<ChatHistoryRecord[]>({
    url: '/chat/history',
    method: 'get',
    params: { groupName, limit }
  })
}
