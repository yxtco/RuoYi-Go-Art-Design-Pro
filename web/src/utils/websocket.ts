/**
 * WebSocket 客户端工具类
 * 
 * 提供 WebSocket 连接的封装，支持自动重连、心跳检测
 * 
 * ## 主要功能
 * 
 * - WebSocket 连接管理
 * - 自动重连机制
 * - 心跳检测（Ping/Pong）
 * - 消息发送和接收
 * - 连接状态管理
 * 
 * @module utils/websocket
 */

import { useUserStore } from '@/store/modules/user'

/** WebSocket 消息类型 */
export interface ChatMessage {
  type: 'private' | 'group' | 'system'
  action: 'message' | 'join' | 'leave' | 'online' | 'typing' | 'joinGroup' | 'leaveGroup' | 'ping' | 'pong'
  fromId?: number
  fromName?: string
  fromAvatar?: string
  toId?: number
  groupName?: string
  content?: string
  msgType?: 'text' | 'image' | 'file' | 'emoji'
  fileUrl?: string
  fileName?: string
  timestamp?: number
  onlineUsers?: OnlineUser[]
}

/** 在线用户信息 */
export interface OnlineUser {
  userId: number
  userName: string
  nickName: string
  avatar: string
  deptName?: string
  postName?: string
}

/** WebSocket 配置 */
export interface WebSocketConfig {
  url: string
  token: string
  reconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeat?: boolean
  heartbeatInterval?: number
}

/** 消息回调 */
export type MessageCallback = (message: ChatMessage) => void
/** WebSocket 连接状态 */
export type WsStatus = 'connected' | 'disconnected' | 'reconnecting' | 'kicked'
export type StatusCallback = (status: WsStatus) => void

/**
 * WebSocket 客户端类
 */
export class WebSocketClient {
  private ws: WebSocket | null = null
  private config: Required<WebSocketConfig>
  private reconnectAttempts = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private messageCallbacks: MessageCallback[] = []
  private statusCallbacks: StatusCallback[] = []
  private kickCallbacks: (() => void)[] = []
  private isManualClose = false
  private lastPongAt = 0
  private kicked = false

  constructor(config: WebSocketConfig) {
    this.config = {
      reconnect: true,
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      heartbeat: true,
      heartbeatInterval: 10000,
      ...config
    }
  }

  /**
   * 连接 WebSocket
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return
    }

    this.isManualClose = false
    this.kicked = false
    this.notifyStatus('reconnecting')

    try {
      // 构建 WebSocket URL：使用相对路径（页面为 https 时浏览器自动使用 wss），token 需编码
      // 与 HTTP 请求保持一致：使用 VITE_API_URL 作为基础地址（开发环境为 /dev-api 经 Vite 代理转发，生产环境为后端完整地址）
      const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')
      const wsUrl = baseUrl.replace(/^http/, 'ws') + `/ws/chat?token=${encodeURIComponent(this.config.token)}`

      this.ws = new WebSocket(wsUrl)
      this.ws.onopen = this.handleOpen.bind(this)
      this.ws.onmessage = this.handleMessage.bind(this)
      this.ws.onclose = this.handleClose.bind(this)
      this.ws.onerror = this.handleError.bind(this)
    } catch (error) {
      console.error('WebSocket 连接失败:', error)
      this.scheduleReconnect()
    }
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    this.isManualClose = true
    this.clearTimers()
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    
    this.reconnectAttempts = 0
    this.notifyStatus('disconnected')
  }

  /**
   * 发送消息
   */
  send(message: ChatMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket 未连接，消息发送失败')
    }
  }

  /**
   * 发送私聊消息
   */
  sendPrivateMessage(toId: number, content: string, msgType: ChatMessage['msgType'] = 'text', timestamp?: number): void {
    this.send({
      type: 'private',
      action: 'message',
      toId,
      content,
      msgType,
      timestamp: timestamp ?? Date.now()
    })
  }

  /**
   * 发送群聊消息
   */
  sendGroupMessage(groupName: string, content: string, msgType: ChatMessage['msgType'] = 'text', timestamp?: number): void {
    this.send({
      type: 'group',
      action: 'message',
      groupName,
      content,
      msgType,
      timestamp: timestamp ?? Date.now()
    })
  }

  /**
   * 发送正在输入提示
   */
  sendTyping(toId: number): void {
    this.send({
      type: 'system',
      action: 'typing',
      toId
    })
  }

  /**
   * 加入群组
   */
  joinGroup(groupName: string): void {
    this.send({
      type: 'system',
      action: 'joinGroup',
      groupName
    })
  }

  /**
   * 离开群组
   */
  leaveGroup(groupName: string): void {
    this.send({
      type: 'system',
      action: 'leaveGroup',
      groupName
    })
  }

  /**
   * 注册消息回调
   */
  onMessage(callback: MessageCallback): void {
    this.messageCallbacks.push(callback)
  }

  /**
   * 移除消息回调
   */
  offMessage(callback: MessageCallback): void {
    const index = this.messageCallbacks.indexOf(callback)
    if (index > -1) {
      this.messageCallbacks.splice(index, 1)
    }
  }

  /**
   * 注册状态回调
   */
  onStatusChange(callback: StatusCallback): void {
    this.statusCallbacks.push(callback)
  }

  /**
   * 注册踢出回调（后端离线 / 心跳超时导致被踢出登录时触发）
   */
  onKick(callback: () => void): void {
    this.kickCallbacks.push(callback)
  }

  /**
   * 获取连接状态
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  // 私有方法

  private handleOpen(): void {
    console.log('WebSocket 连接成功')
    this.reconnectAttempts = 0
    this.lastPongAt = Date.now()
    this.kicked = false
    this.notifyStatus('connected')
    this.startHeartbeat()
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message: ChatMessage = JSON.parse(event.data)
      // 收到 pong 心跳回复，记录时间
      if (message.type === 'system' && message.action === 'pong') {
        this.lastPongAt = Date.now()
      }
      this.messageCallbacks.forEach(callback => callback(message))
    } catch (error) {
      console.error('解析 WebSocket 消息失败:', error)
    }
  }

  private handleClose(event: CloseEvent): void {
    console.log('WebSocket 连接关闭:', event.code, event.reason)
    this.stopHeartbeat()
    this.notifyStatus('disconnected')

    if (!this.isManualClose && this.config.reconnect) {
      this.scheduleReconnect()
    }
  }

  private handleError(error: Event): void {
    console.error('WebSocket 错误:', error)
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.warn('达到最大重连次数，判定后端离线，执行踢出')
      this.kick()
      return
    }

    this.reconnectAttempts++
    this.notifyStatus('reconnecting')

    this.reconnectTimer = setTimeout(() => {
      console.log(`尝试第 ${this.reconnectAttempts} 次重连...`)
      this.connect()
    }, this.config.reconnectInterval)
  }

  /**
   * 心跳：定时发送 ping，后端响应 pong；若超过阈值未收到 pong 判定离线并踢出
   */
  private startHeartbeat(): void {
    if (!this.config.heartbeat) return

    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        // 应用层心跳 ping
        this.send({ type: 'system', action: 'ping' })
        // 超过 2 个周期未收到 pong，判定后端不在线
        if (Date.now() - this.lastPongAt > this.config.heartbeatInterval * 2) {
          console.warn('心跳 pong 超时，判定后端离线，执行踢出')
          this.kick()
        }
      }
    }, this.config.heartbeatInterval)
  }

  /**
   * 执行踢出：强制关闭连接并通知（供上层调用退出登录）
   */
  private kick(): void {
    if (this.kicked) return
    this.kicked = true
    this.stopHeartbeat()
    this.isManualClose = true
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.reconnectAttempts = 0
    this.notifyStatus('kicked')
    this.kickCallbacks.forEach(cb => cb())
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private clearTimers(): void {
    this.stopHeartbeat()
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }

  private notifyStatus(status: WsStatus): void {
    this.statusCallbacks.forEach(callback => callback(status))
  }
}

/** 全局 WebSocket 实例 */
let wsInstance: WebSocketClient | null = null

/**
 * 获取 WebSocket 实例
 */
export function getWebSocket(): WebSocketClient | null {
  return wsInstance
}

/**
 * 初始化 WebSocket 连接
 */
export function initWebSocket(): WebSocketClient {
  const userStore = useUserStore()
  
  if (wsInstance) {
    wsInstance.disconnect()
  }

  const baseUrl = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')
  const wsUrl = baseUrl.replace(/^http/, 'ws') + '/ws/chat'

  wsInstance = new WebSocketClient({
    url: wsUrl,
    token: userStore.accessToken
  })

  wsInstance.connect()
  return wsInstance
}

/**
 * 关闭 WebSocket 连接
 */
export function closeWebSocket(): void {
  if (wsInstance) {
    wsInstance.disconnect()
    wsInstance = null
  }
}
