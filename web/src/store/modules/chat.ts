/**
 * 聊天状态管理模块
 * 
 * 管理聊天相关的状态，包括：
 * - 在线用户列表
 * - 聊天消息
 * - 当前聊天对象
 * - WebSocket 连接状态
 * 
 * @module store/modules/chat
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'

import { fetchOnlineUsers, fetchPrivateHistory, fetchGroupHistory, type ChatHistoryRecord } from '@/api/chat'
import type { ChatMessage, OnlineUser } from '@utils/websocket'
import { getWebSocket, initWebSocket, closeWebSocket } from '@utils/websocket'
import { useUserStore } from './user'

// Re-export OnlineUser type for external use
export type { OnlineUser }

/** 聊天消息（扩展本地显示用） */
export interface ChatMessageLocal extends ChatMessage {
  id: string
  isMe: boolean
  time: string
}

/** 聊天会话 */
export interface ChatSession {
  id: string
  type: 'private' | 'group'
  targetId: number        // 对方用户ID 或 群组ID
  targetName: string
  targetAvatar: string
  lastMessage: string
  lastTime: number
  unreadCount: number
  messages: ChatMessageLocal[]
}

export const useChatStore = defineStore('chatStore', () => {
  // WebSocket 连接状态
  const wsStatus = ref<'connected' | 'disconnected' | 'reconnecting' | 'kicked'>('disconnected')
  
  // 在线用户列表
  const onlineUsers = ref<OnlineUser[]>([])
  
  // 聊天会话列表
  const sessions = ref<ChatSession[]>([])
  
  // 当前活跃的会话ID
  const activeSessionId = ref<string>('')
  
  // 当前用户信息
  const userStore = useUserStore()
  const currentUserId = computed(() => userStore.info.userId || 0)
  const currentUserName = computed(() => userStore.info.nickName || userStore.info.userName || '')
  const currentUserAvatar = computed(() => userStore.info.avatar || '')

  // 当前活跃会话
  const activeSession = computed(() => {
    return sessions.value.find(s => s.id === activeSessionId.value) || null
  })

  // 未读消息总数
  const totalUnreadCount = computed(() => {
    return sessions.value.reduce((sum, s) => sum + s.unreadCount, 0)
  })

  // 各在线用户的私聊未读数（用于在线列表红标：对方发给我的未读消息数）
  const privateUnreadByUser = computed<Record<number, number>>(() => {
    const map: Record<number, number> = {}
    for (const s of sessions.value) {
      if (s.type === 'private' && s.targetId) {
        map[s.targetId] = s.unreadCount
      }
    }
    return map
  })

  /**
   * 初始化 WebSocket 连接
   */
  function connect() {
    if (wsStatus.value === 'connected') return

    const ws = initWebSocket()

    ws.onStatusChange((status) => {
      wsStatus.value = status
      // 后端离线被踢出：直接退出登录
      if (status === 'kicked') {
        handleKicked()
      }
    })

    ws.onMessage(handleMessage)
    // 后端离线/心跳超时被踢出时退出登录
    ws.onKick(handleKicked)

    // 通过 HTTP 接口兜底拉取在线用户（即使 WS 的 online 消息丢失也能展示）
    refreshOnlineUsers()
  }

  /**
   * 被踢出（后端不在线）：退出登录
   */
  function handleKicked() {
    closeWebSocket()
    wsStatus.value = 'disconnected'
    onlineUsers.value = []
    ElMessage.error('连接已断开，请重新登录')
    userStore.logOut()
  }

  /**
   * 从 HTTP 接口刷新在线用户列表（兜底，同时刷新岗位/部门信息）
   */
  async function refreshOnlineUsers() {
    try {
      const data = await fetchOnlineUsers()
      if (Array.isArray(data)) {
        onlineUsers.value = data.filter(u => u.userId !== currentUserId.value)
      }
    } catch (error) {
      // 接口失败静默，等待 WS 在线消息补充
      console.warn('获取在线用户列表失败:', error)
    }
  }

  /**
   * 断开 WebSocket 连接
   */
  function disconnect() {
    closeWebSocket()
    wsStatus.value = 'disconnected'
  }

  /**
   * 处理接收到的 WebSocket 消息
   */
  function handleMessage(message: ChatMessage) {
    // 处理系统消息
    if (message.type === 'system') {
      handleSystemMessage(message)
      return
    }

    // 处理聊天消息
    const localMsg: ChatMessageLocal = {
      ...message,
      id: `${message.timestamp}-${message.fromId}-${Math.random().toString(36).slice(2)}`,
      isMe: message.fromId === currentUserId.value,
      time: formatTime(message.timestamp || Date.now())
    }

    if (message.type === 'private') {
      handlePrivateMessage(localMsg)
    } else if (message.type === 'group') {
      handleGroupMessage(localMsg)
    }
  }

  /**
   * 处理系统消息
   */
  function handleSystemMessage(message: ChatMessage) {
    switch (message.action) {
      case 'online':
        // 更新在线用户列表
        if (message.onlineUsers) {
          onlineUsers.value = message.onlineUsers.filter(
            u => u.userId !== currentUserId.value
          )
        }
        break
      case 'join':
        // 用户上线
        if (message.fromId && message.fromId !== currentUserId.value) {
          const exists = onlineUsers.value.some(u => u.userId === message.fromId)
          if (!exists && message.fromName) {
            onlineUsers.value.push({
              userId: message.fromId,
              userName: message.fromName,
              nickName: message.fromName,
              avatar: message.fromAvatar || ''
            })
          }
        }
        break
      case 'leave':
        // 用户下线
        if (message.fromId) {
          onlineUsers.value = onlineUsers.value.filter(u => u.userId !== message.fromId)
        }
        break
    }
  }

  /**
   * 处理私聊消息
   */
  function handlePrivateMessage(msg: ChatMessageLocal) {
    const otherUserId = msg.isMe ? msg.toId! : msg.fromId!
    const sessionId = `private-${otherUserId}`
    
    // 收到他人消息时，把发送者补进在线用户列表（保证在线列表能看到该用户及其未读数）
    if (!msg.isMe && msg.fromId) {
      const exists = onlineUsers.value.some(u => u.userId === msg.fromId)
      if (!exists) {
        onlineUsers.value.unshift({
          userId: msg.fromId,
          userName: msg.fromName || '',
          nickName: msg.fromName || '',
          avatar: msg.fromAvatar || ''
        })
      }
    }
    
    let session = sessions.value.find(s => s.id === sessionId)
    
    if (!session) {
      // 创建新会话
      session = {
        id: sessionId,
        type: 'private',
        targetId: otherUserId,
        targetName: msg.isMe ? (msg.toId?.toString() || '') : (msg.fromName || ''),
        targetAvatar: msg.isMe ? '' : (msg.fromAvatar || ''),
        lastMessage: msg.content || '',
        lastTime: msg.timestamp || Date.now(),
        unreadCount: msg.isMe ? 0 : 1,
        messages: [msg]
      }
      sessions.value.unshift(session)
    } else {
      session.messages.push(msg)
      session.lastMessage = msg.content || ''
      session.lastTime = msg.timestamp || Date.now()
      if (!msg.isMe && session.id !== activeSessionId.value) {
        session.unreadCount++
      }
    }
  }

  /**
   * 处理群聊消息
   */
  function handleGroupMessage(msg: ChatMessageLocal) {
    const groupName = msg.groupName || ''
    const sessionId = `group-${groupName}`
    
    let session = sessions.value.find(s => s.id === sessionId)
    
    if (!session) {
      session = {
        id: sessionId,
        type: 'group',
        targetId: 0,
        targetName: groupName,
        targetAvatar: '',
        lastMessage: msg.content || '',
        lastTime: msg.timestamp || Date.now(),
        unreadCount: msg.isMe ? 0 : 1,
        messages: [msg]
      }
      sessions.value.unshift(session)
    } else {
      session.messages.push(msg)
      session.lastMessage = msg.content || ''
      session.lastTime = msg.timestamp || Date.now()
      if (!msg.isMe && session.id !== activeSessionId.value) {
        session.unreadCount++
      }
    }
  }

  /**
   * 打开与指定用户的私聊
   */
  function openPrivateChat(user: OnlineUser) {
    const sessionId = `private-${user.userId}`
    let session = sessions.value.find(s => s.id === sessionId)
    
    if (!session) {
      session = {
        id: sessionId,
        type: 'private',
        targetId: user.userId,
        targetName: user.nickName || user.userName,
        targetAvatar: user.avatar || '',
        lastMessage: '',
        lastTime: Date.now(),
        unreadCount: 0,
        messages: []
      }
      sessions.value.unshift(session)
    }
    
    // 清除未读
    session.unreadCount = 0
    activeSessionId.value = sessionId

    // 加载持久化历史消息
    loadHistory(sessionId, 'private', user.userId, '')
  }

  /**
   * 打开群组聊天
   */
  function openGroupChat(groupName: string) {
    const sessionId = `group-${groupName}`
    let session = sessions.value.find(s => s.id === sessionId)
    
    if (!session) {
      session = {
        id: sessionId,
        type: 'group',
        targetId: 0,
        targetName: groupName,
        targetAvatar: '',
        lastMessage: '',
        lastTime: Date.now(),
        unreadCount: 0,
        messages: []
      }
      sessions.value.unshift(session)
      
      // 加入群组
      const ws = getWebSocket()
      ws?.joinGroup(groupName)
    }
    
    session.unreadCount = 0
    activeSessionId.value = sessionId

    // 加载持久化历史消息
    loadHistory(sessionId, 'group', 0, groupName)
  }

  /**
   * 将持久化的历史记录转换为本地消息格式
   */
  function historyToLocal(record: ChatHistoryRecord): ChatMessageLocal {
    return {
      id: `hist-${record.id}`,
      isMe: record.fromUserId === currentUserId.value,
      type: record.toUserId ? 'private' : 'group',
      action: 'message',
      fromId: record.fromUserId,
      fromName: record.fromName,
      fromAvatar: record.fromAvatar,
      toId: record.toUserId,
      groupName: record.groupName || '',
      content: record.content,
      msgType: record.msgType,
      fileUrl: record.fileUrl || '',
      fileName: record.fileName || '',
      timestamp: record.timestamp,
      time: formatTime(record.timestamp)
    }
  }

  /**
   * 加载会话历史消息（仅首次拉取，历史在前、实时消息在后）
   */
  async function loadHistory(sessionId: string, type: 'private' | 'group', targetId: number, groupName: string) {
    const session = sessions.value.find(s => s.id === sessionId)
    if (!session) return
    // 已加载过历史则跳过
    if (session.messages.some(m => m.id.startsWith('hist-'))) return

    try {
      const records = type === 'private'
        ? await fetchPrivateHistory(targetId, 100)
        : await fetchGroupHistory(groupName, 100)
      const locals = (records || []).map(historyToLocal)
      // 去重：以「发送者+时间戳+内容」为指纹，过滤掉已在会话中的消息，
      // 覆盖「实时收到/本地乐观发送的消息」与「历史加载」重叠导致的重复显示（如离线/抖动后重连）
      const existingFingerprints = new Set<string>()
      for (const m of session.messages) {
        if (m.fromId && m.timestamp && m.content) {
          existingFingerprints.add(`${m.fromId}:${m.timestamp}:${m.content}`)
        }
      }
      const filtered = locals.filter((m) => {
        if (m.fromId && m.timestamp && m.content) {
          const fp = `${m.fromId}:${m.timestamp}:${m.content}`
          if (existingFingerprints.has(fp)) return false
          existingFingerprints.add(fp)
        }
        return true
      })
      session.messages = [...filtered, ...session.messages]
      if (filtered.length > 0) {
        const last = session.messages[session.messages.length - 1]
        session.lastMessage = last.content || session.lastMessage
        session.lastTime = last.timestamp || session.lastTime
      }
    } catch (error) {
      console.warn(`加载 ${sessionId} 聊天历史失败:`, error)
    }
  }

  /**
   * 发送消息
   */
  function sendMessage(content: string, msgType: ChatMessage['msgType'] = 'text') {
    const session = activeSession.value
    if (!session) return
    
    const ws = getWebSocket()
    if (!ws) return

    // 统一使用同一个时间戳：既用于 WS 发送（后端落库），也用于本地乐观消息，
    // 保证历史加载去重（发送者+时间戳+内容）能命中自己发送的消息
    const now = Date.now()

    if (session.type === 'private') {
      ws.sendPrivateMessage(session.targetId, content, msgType, now)
    } else if (session.type === 'group') {
      ws.sendGroupMessage(session.targetName, content, msgType, now)
    }
    
    // 添加本地消息到会话
    const localMsg: ChatMessageLocal = {
      type: session.type,
      action: 'message',
      fromId: currentUserId.value,
      fromName: currentUserName.value,
      fromAvatar: currentUserAvatar.value,
      toId: session.type === 'private' ? session.targetId : undefined,
      groupName: session.type === 'group' ? session.targetName : undefined,
      content,
      msgType,
      timestamp: now,
      id: `${now}-${currentUserId.value}-${Math.random().toString(36).slice(2)}`,
      isMe: true,
      time: formatTime(now)
    }
    
    session.messages.push(localMsg)
    session.lastMessage = content
    session.lastTime = now
  }

  /**
   * 关闭会话
   */
  function closeSession(sessionId: string) {
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index > -1) {
      const session = sessions.value[index]
      if (session.type === 'group') {
        const ws = getWebSocket()
        ws?.leaveGroup(session.targetName)
      }
      sessions.value.splice(index, 1)
    }
    
    if (activeSessionId.value === sessionId) {
      activeSessionId.value = sessions.value.length > 0 ? sessions.value[0].id : ''
    }
  }

  /**
   * 格式化时间
   */
  function formatTime(timestamp: number): string {
    const date = new Date(timestamp)
    const now = new Date()
    const isToday = date.toDateString() === now.toDateString()
    
    if (isToday) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    
    const yesterday = new Date(now)
    yesterday.setDate(yesterday.getDate() - 1)
    if (date.toDateString() === yesterday.toDateString()) {
      return '昨天 ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    
    return date.toLocaleDateString([], { month: '2-digit', day: '2-digit' }) + 
           ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  return {
    wsStatus,
    onlineUsers,
    sessions,
    activeSessionId,
    activeSession,
    totalUnreadCount,
    privateUnreadByUser,
    currentUserId,
    currentUserName,
    currentUserAvatar,
    connect,
    disconnect,
    refreshOnlineUsers,
    handleMessage,
    openPrivateChat,
    openGroupChat,
    sendMessage,
    closeSession
  }
}, {
  persist: {
    key: 'chat',
    storage: sessionStorage,
    pick: ['sessions', 'activeSessionId']
  }
})
