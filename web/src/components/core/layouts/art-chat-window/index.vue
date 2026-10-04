<!-- 系统聊天窗口 -->
<template>
  <div>
    <ElDrawer
      v-model="isDrawerVisible"
      :size="isMobile ? '100%' : '720px'"
      :with-header="false"
      class="chat-drawer">
      <div class="flex h-full">
        <!-- 左侧：会话列表 + 在线用户 -->
        <div class="chat-sidebar flex flex-col border-r border-g-200 dark:border-g-700">
          <!-- 头部 -->
          <div class="flex-cb px-4 py-3 border-b border-g-200 dark:border-g-700">
            <span class="text-base font-medium text-g-800">消息</span>
            <div class="flex-c gap-2">
              <ArtSvgIcon
                icon="ri:group-add-line"
                class="c-p text-lg text-g-600 hover:text-theme!"
                title="发起群聊"
                @click="openGroupChatDialog" />
              <ElBadge :value="chatStore.totalUnreadCount" :hidden="chatStore.totalUnreadCount === 0" type="primary">
                <ArtSvgIcon
                  icon="ri:group-line"
                  class="c-p text-lg text-g-600"
                  :class="{ 'text-theme!': showOnlineUsers }"
                  @click="showOnlineUsers = !showOnlineUsers" />
              </ElBadge>
              <ArtSvgIcon
                icon="ri:close-line"
                class="c-p text-lg text-g-600"
                @click="closeChat" />
            </div>
          </div>

          <!-- 连接状态 + 在线人数 -->
          <div class="flex-c gap-1.5 px-4 py-2 border-b border-g-200 dark:border-g-700">
            <div
              class="h-2 w-2 rounded-full"
              :class="statusClass"></div>
            <span class="text-xs text-g-500">{{ statusText }}</span>
            <span class="text-xs text-g-400">·</span>
            <span class="text-xs font-medium text-theme">{{ '在线 ' + chatStore.onlineUsers.length + ' 人' }}</span>
          </div>

          <!-- 在线用户列表 -->
          <div v-if="showOnlineUsers" class="flex-1 overflow-y-auto">
            <div class="px-4 py-2">
              <span class="text-xs font-medium text-g-500">在线用户</span>
            </div>
            <ul class="scrollbar-thin">
              <li
                v-for="user in chatStore.onlineUsers"
                :key="user.userId"
                class="flex-c c-p gap-3 px-4 py-2.5 hover:bg-g-100 dark:hover:bg-g-800"
                @click="startChat(user)">
                <ElAvatar :size="36" :src="avatarSrc(user.avatar)" class="shrink-0">
                  {{ user.nickName?.charAt(0) || user.userName?.charAt(0) }}
                </ElAvatar>
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium text-g-800 truncate">
                    {{ user.nickName || user.userName }}
                  </p>
                  <p class="mt-0.5 flex items-center gap-1 text-xs text-g-500 truncate">
                    <ArtSvgIcon v-if="user.postName" icon="ri:user-star-line" class="text-g-400" />
                    <span>{{ user.postName || user.deptName || user.userName }}</span>
                  </p>
                </div>
                <ElBadge
                  :value="chatStore.privateUnreadByUser[user.userId] || 0"
                  :hidden="!chatStore.privateUnreadByUser[user.userId]"
                  type="danger"
                  :offset="[0, 0]">
                  <ArtSvgIcon
                    icon="ri:message-3-line"
                    class="text-g-400" />
                </ElBadge>
              </li>
            </ul>
            <div
              v-if="chatStore.onlineUsers.length === 0"
              class="flex-col-c py-8 text-g-400">
              <ArtSvgIcon icon="ri:user-line" class="text-3xl" />
              <p class="mt-2 text-xs">暂无其他用户在线</p>
            </div>
          </div>

          <!-- 会话列表 -->
          <div v-else class="flex-1 overflow-y-auto">
            <ul class="scrollbar-thin">
              <li
                v-for="session in chatStore.sessions"
                :key="session.id"
                class="flex-c c-p gap-3 px-4 py-3 hover:bg-g-100 dark:hover:bg-g-800"
                :class="{ 'bg-g-100 dark:bg-g-800': session.id === chatStore.activeSessionId }"
                @click="selectSession(session)">
                <ElAvatar
                  :size="40"
                  :src="session.type === 'group' ? '' : avatarSrc(session.targetAvatar)"
                  class="shrink-0">
                  <ArtSvgIcon v-if="session.type === 'group'" icon="ri:team-line" class="text-g-500" />
                  <template v-else>{{ session.targetName?.charAt(0) }}</template>
                </ElAvatar>
                <div class="flex-1 min-w-0">
                  <div class="flex-cb">
                    <p class="text-sm font-medium text-g-800 truncate">
                      {{ session.targetName }}
                    </p>
                    <span class="text-xs text-g-400">{{ formatSessionTime(session.lastTime) }}</span>
                  </div>
                  <div class="flex-cb mt-0.5">
                    <p class="text-xs text-g-500 truncate">{{ session.lastMessage || '暂无消息' }}</p>
                    <ElBadge
                      v-if="session.unreadCount > 0"
                      :value="session.unreadCount"
                      type="primary" />
                  </div>
                </div>
                <ArtSvgIcon
                  icon="ri:close-line"
                  class="c-p shrink-0 text-g-300 hover:text-danger!"
                  title="关闭该会话"
                  @click.stop="chatStore.closeSession(session.id)" />
              </li>
            </ul>
            <div
              v-if="chatStore.sessions.length === 0"
              class="flex-col-c py-12 text-g-400">
              <ArtSvgIcon icon="ri:message-3-line" class="text-4xl" />
              <p class="mt-3 text-xs">暂无聊天消息</p>
              <p class="mt-1 text-xs">点击左侧在线用户开始聊天</p>
            </div>
          </div>
        </div>

        <!-- 右侧：聊天区域 -->
        <div class="flex-1 flex flex-col min-w-0">
          <template v-if="chatStore.activeSession">
            <!-- 聊天头部 -->
            <div class="flex-cb px-4 py-3 border-b border-g-200 dark:border-g-700">
              <div class="flex-c gap-3">
                <ElAvatar
                  :size="32"
                  :src="chatStore.activeSession.type === 'group' ? '' : avatarSrc(chatStore.activeSession.targetAvatar)">
                  <ArtSvgIcon v-if="chatStore.activeSession.type === 'group'" icon="ri:team-line" class="text-g-500" />
                  <template v-else>{{ chatStore.activeSession.targetName?.charAt(0) }}</template>
                </ElAvatar>
                <div>
                  <p class="text-sm font-medium text-g-800">
                    {{ chatStore.activeSession.targetName }}
                  </p>
                  <p class="text-xs text-g-500">
                    {{ chatStore.activeSession.type === 'private' ? '私聊' : '群聊' }}
                  </p>
                </div>
              </div>
              <ArtSvgIcon
                icon="ri:close-line"
                class="c-p text-lg text-g-500"
                title="关闭聊天窗口"
                @click="closeChat" />
            </div>

            <!-- 消息区域 -->
            <div
              ref="messageContainer"
              class="flex-1 overflow-y-auto px-4 py-4 scrollbar-thin">
              <template v-for="message in chatStore.activeSession.messages" :key="message.id">
                <div
                  :class="[
                    'mb-4 flex w-full items-start gap-2',
                    message.isMe ? 'flex-row-reverse' : 'flex-row'
                  ]">
                  <ElAvatar :size="32" :src="avatarSrc(message.fromAvatar)" class="shrink-0">
                    {{ message.fromName?.charAt(0) }}
                  </ElAvatar>
                  <div
                    :class="[
                      'flex max-w-[70%] flex-col',
                      message.isMe ? 'items-end' : 'items-start'
                    ]">
                    <div
                      :class="[
                        'mb-1 flex gap-2 text-xs',
                        message.isMe ? 'flex-row-reverse' : 'flex-row'
                      ]">
                      <span class="font-medium text-g-700">{{ message.fromName }}</span>
                      <span class="text-g-400">{{ message.time }}</span>
                    </div>
                    <!-- 文本消息 -->
                    <div
                      v-if="message.msgType === 'text' || !message.msgType"
                      :class="[
                        'rounded-lg px-3.5 py-2.5 text-sm leading-relaxed text-g-900 break-words',
                        message.isMe
                          ? 'bg-theme/15 rounded-tr-sm'
                          : 'bg-g-100 dark:bg-g-800 rounded-tl-sm'
                      ]"
                      v-html="formatContent(message.content)">
                    </div>
                    <!-- 图片消息 -->
                    <div
                      v-else-if="message.msgType === 'image'"
                      :class="[
                        'rounded-lg overflow-hidden',
                        message.isMe ? 'rounded-tr-sm' : 'rounded-tl-sm'
                      ]">
                      <ElImage
                        :src="message.fileUrl"
                        :preview-src-list="[message.fileUrl!]"
                        fit="cover"
                        class="max-w-[200px] max-h-[200px]" />
                    </div>
                    <!-- 文件消息 -->
                    <div
                      v-else-if="message.msgType === 'file'"
                      :class="[
                        'flex-c gap-2 rounded-lg px-3 py-2',
                        message.isMe ? 'bg-theme/15 rounded-tr-sm' : 'bg-g-100 dark:bg-g-800 rounded-tl-sm'
                      ]">
                      <ArtSvgIcon icon="ri:file-line" class="text-lg text-g-600" />
                      <a
                        :href="message.fileUrl"
                        target="_blank"
                        class="text-sm text-theme hover:underline">
                        {{ message.fileName || '下载文件' }}
                      </a>
                    </div>
                  </div>
                </div>
              </template>

              <!-- 空消息提示 -->
              <div
                v-if="chatStore.activeSession.messages.length === 0"
                class="flex-col-c h-full text-g-400">
                <ArtSvgIcon icon="ri:chat-smile-line" class="text-4xl" />
                <p class="mt-3 text-sm">开始对话吧</p>
              </div>
            </div>

            <!-- 输入区域 -->
            <div class="border-t border-g-200 dark:border-g-700 px-4 py-3">
              <!-- 工具栏 -->
              <div class="flex-c gap-3 mb-2">
                <ArtSvgIcon
                  icon="ri:emotion-happy-line"
                  class="c-p text-lg text-g-500 hover:text-theme!"
                  @click="showEmojiPicker = !showEmojiPicker" />
                <ArtSvgIcon
                  icon="ri:image-line"
                  class="c-p text-lg text-g-500 hover:text-theme!"
                  @click="triggerImageUpload" />
                <ArtSvgIcon
                  icon="ri:attachment-line"
                  class="c-p text-lg text-g-500 hover:text-theme!"
                  @click="triggerFileUpload" />
              </div>

              <!-- 输入框 -->
              <div class="flex gap-2">
                <ElInput
                  v-model="inputMessage"
                  type="textarea"
                  :rows="2"
                  placeholder="输入消息，Enter 发送，Shift+Enter 换行"
                  resize="none"
                  @keydown="handleKeydown" />
                <ElButton
                  type="primary"
                  :disabled="!inputMessage.trim()"
                  @click="sendMessage"
                  class="self-end h-10">
                  发送
                </ElButton>
              </div>

              <!-- 表情选择器 -->
              <div v-if="showEmojiPicker" class="mt-2 p-2 bg-g-50 dark:bg-g-800 rounded-lg">
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="emoji in emojis"
                    :key="emoji"
                    class="c-p text-xl p-1 hover:bg-g-200 dark:hover:bg-g-700 rounded"
                    @click="insertEmoji(emoji)">
                    {{ emoji }}
                  </span>
                </div>
              </div>
            </div>
          </template>

          <!-- 未选择会话提示 -->
          <div
            v-else
            class="flex-col-c h-full text-g-400">
            <ArtSvgIcon icon="ri:message-3-line" class="text-5xl" />
            <p class="mt-4 text-base">选择一个对话开始聊天</p>
            <p class="mt-2 text-sm">从左侧选择在线用户或已有会话</p>
          </div>
        </div>
      </div>

      <!-- 隐藏的文件输入 -->
      <input
        ref="imageInput"
        type="file"
        accept="image/*"
        class="hidden"
        @change="handleImageUpload" />
      <input
        ref="fileInput"
        type="file"
        class="hidden"
        @change="handleFileUpload" />
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
import { mittBus } from '@utils/sys'
import { useChatStore } from '@/store/modules/chat'
import type { OnlineUser, ChatSession } from '@/store/modules/chat'
import request from '@utils/http'
import { ElMessageBox } from 'element-plus'
import defAva from '@imgs/user/avatar.webp'

defineOptions({ name: 'ArtChatWindow' })

// 常量
const MOBILE_BREAKPOINT = 640

// 常用表情
const emojis = [
  '😀', '😃', '😄', '😁', '😆', '😅', '🤣', '😂',
  '🙂', '🙃', '😉', '😊', '😇', '🥰', '😍', '🤩',
  '😘', '😗', '😚', '😙', '🥲', '😋', '😛', '😜',
  '🤪', '😝', '🤑', '🤗', '🤭', '🤫', '🤔', '😐',
  '👍', '👎', '👏', '🙌', '🤝', '✌️', '🤞', '🤟',
  '❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '🤍'
]

// 响应式布局
const { width } = useWindowSize()
const isMobile = computed(() => width.value < MOBILE_BREAKPOINT)

// Store
const chatStore = useChatStore()

// 组件状态
const isDrawerVisible = ref(false)
const showOnlineUsers = ref(true)
const showEmojiPicker = ref(false)
const inputMessage = ref('')
const messageContainer = ref<HTMLElement | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
let onlineRefreshTimer: ReturnType<typeof setInterval> | null = null

// 头像加载失败记录（key 为 userId 或 sessionId），失败后回退为默认头像
const brokenAvatar = ref<Record<string, boolean>>({})
function markAvatarBroken(key: string | number) {
  brokenAvatar.value[String(key)] = true
}
function avatarOk(key: string | number, url?: string): boolean {
  return !!url && !brokenAvatar.value[String(key)]
}

/**
 * 头像兜底：头像为空（未设置）时显示内置默认头像，与顶部导航栏/用户菜单保持一致
 */
function avatarSrc(url?: string): string {
  return url || defAva
}

// 计算属性
const statusClass = computed(() => {
  switch (chatStore.wsStatus) {
    case 'connected': return 'bg-success'
    case 'reconnecting': return 'bg-warning'
    default: return 'bg-danger'
  }
})

const statusText = computed(() => {
  switch (chatStore.wsStatus) {
    case 'connected': return '已连接'
    case 'reconnecting': return '重连中...'
    case 'kicked': return '连接已断开'
    default: return '未连接'
  }
})

// 方法
function openChat() {
  isDrawerVisible.value = true
  
  // 连接 WebSocket（如果未连接）
  if (chatStore.wsStatus !== 'connected') {
    chatStore.connect()
  }
  
  // 打开时刷新在线用户列表
  chatStore.refreshOnlineUsers()
  
  scrollToBottom()
}

function closeChat() {
  isDrawerVisible.value = false
  showEmojiPicker.value = false
}

function startChat(user: OnlineUser) {
  chatStore.openPrivateChat(user)
  // 保持在线用户列表可见，便于看到其他用户及其未读数
  scrollToBottom()
}

function openGroupChatDialog() {
  ElMessageBox.prompt('请输入群聊名称，其他在线用户加入即可多人沟通', '发起群聊', {
    confirmButtonText: '创建',
    cancelButtonText: '取消',
    inputPlaceholder: '例如：项目讨论组',
    inputValidator: (v) => !!v?.trim() || '群聊名称不能为空'
  })
    .then(({ value }) => {
      const name = value?.trim()
      if (!name) return
      chatStore.openGroupChat(name)
      scrollToBottom()
    })
    .catch(() => {})
}

function selectSession(session: ChatSession) {
  chatStore.activeSessionId = session.id
  scrollToBottom()
}

function sendMessage() {
  const text = inputMessage.value.trim()
  if (!text) return
  
  chatStore.sendMessage(text)
  inputMessage.value = ''
  showEmojiPicker.value = false
  scrollToBottom()
}

function handleKeydown(e: Event) {
  const keyEvent = e as KeyboardEvent
  if (keyEvent.key === 'Enter' && !keyEvent.shiftKey) {
    keyEvent.preventDefault()
    sendMessage()
  }
}

function insertEmoji(emoji: string) {
  inputMessage.value += emoji
}

function triggerImageUpload() {
  imageInput.value?.click()
}

function triggerFileUpload() {
  fileInput.value?.click()
}

async function handleImageUpload(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  
  try {
    const formData = new FormData()
    formData.append('file', file)
    
    const res = await request<{ url: string }>({
      url: '/common/upload',
      method: 'post',
      data: formData,
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    
    if (res?.url) {
      chatStore.sendMessage(res.url, 'image')
    }
  } catch (error) {
    console.error('图片上传失败:', error)
  }
  
  target.value = ''
}

async function handleFileUpload(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  
  try {
    const formData = new FormData()
    formData.append('file', file)
    
    const res = await request<{ url: string, fileName: string }>({
      url: '/common/upload',
      method: 'post',
      data: formData,
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    
    if (res?.url) {
      // 通过 WebSocket 发送文件消息
      const ws = (await import('@utils/websocket')).getWebSocket()
      const session = chatStore.activeSession
      if (ws && session) {
        ws.send({
          type: session.type,
          action: 'message',
          toId: session.type === 'private' ? session.targetId : undefined,
          groupName: session.type === 'group' ? session.targetName : undefined,
          content: '',
          msgType: 'file',
          fileUrl: res.url,
          fileName: res.fileName || file.name,
          timestamp: Date.now()
        })
      }
    }
  } catch (error) {
    console.error('文件上传失败:', error)
  }
  
  target.value = ''
}

function formatContent(content: string | undefined): string {
  if (!content) return ''
  // 简单的链接识别
  return content.replace(
    /(https?:\/\/[^\s]+)/g,
    '<a href="$1" target="_blank" class="text-theme hover:underline">$1</a>'
  )
}

function formatSessionTime(timestamp: number): string {
  const date = new Date(timestamp)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()
  
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (date.toDateString() === yesterday.toDateString()) {
    return '昨天'
  }
  
  return date.toLocaleDateString([], { month: '2-digit', day: '2-digit' })
}

function scrollToBottom() {
  nextTick(() => {
    setTimeout(() => {
      if (messageContainer.value) {
        messageContainer.value.scrollTop = messageContainer.value.scrollHeight
      }
    }, 100)
  })
}

// 监听消息变化，自动滚动
watch(
  () => chatStore.activeSession?.messages.length,
  () => scrollToBottom()
)

// 生命周期
onMounted(() => {
  mittBus.on('openChat', openChat)
  // 周期性刷新在线用户列表，保证列表完整、不缩减
  onlineRefreshTimer = setInterval(() => {
    if (chatStore.wsStatus === 'connected') {
      chatStore.refreshOnlineUsers()
    }
  }, 30000)
})

onUnmounted(() => {
  mittBus.off('openChat', openChat)
  if (onlineRefreshTimer) {
    clearInterval(onlineRefreshTimer)
    onlineRefreshTimer = null
  }
})
</script>

<style lang="scss" scoped>
@reference '@styles/core/tailwind.css';

.chat-drawer {
  :deep(.el-drawer__body) {
    padding: 0;
    height: 100%;
  }
}

.chat-sidebar {
  width: 200px;
  
  @media screen and (width <= 640px) {
    width: 100%;
  }
}

.scrollbar-thin {
  &::-webkit-scrollbar {
    width: 4px;
  }
  
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  
  &::-webkit-scrollbar-thumb {
    background-color: var(--el-border-color-lighter);
    border-radius: 2px;
  }
}

:deep(.el-textarea__inner) {
  box-shadow: none;
  border: 1px solid var(--el-border-color-lighter);
  
  &:focus {
    border-color: var(--el-color-primary);
  }
}
</style>
