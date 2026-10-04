<template>
  <div class="workbench-container p-4">
    <!-- 欢迎区域 -->
    <div class="welcome-banner art-card mb-5">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <el-avatar :size="56" :src="userStore.info.avatar" class="welcome-avatar">
            {{ userStore.info.nickName?.charAt(0) || 'U' }}
          </el-avatar>
          <div>
            <h2 class="welcome-title">
              {{ greeting }}，{{ userStore.info.nickName || userStore.info.userName }}
            </h2>
            <p class="welcome-subtitle">
              {{ todayStr }} · {{ weekDayStr }}
            </p>
          </div>
        </div>
        <div class="flex gap-3">
          <el-tag type="primary" effect="plain" size="large">
            <ArtSvgIcon icon="ri:shield-user-line" class="mr-1 text-sm" />
            {{ roleLabel }}
          </el-tag>
          <el-tag type="success" effect="plain" size="large">
            <ArtSvgIcon icon="ri:time-line" class="mr-1 text-sm" />
            {{ currentTime }}
          </el-tag>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="16" class="mb-5">
      <el-col :xs="12" :sm="12" :md="6" v-for="item in statCards" :key="item.title">
        <div class="art-card stat-card mb-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="stat-label text-sm text-g-500">{{ item.title }}</p>
              <p class="stat-value text-2xl font-bold mt-2 text-g-900">
                {{ item.value }}
              </p>
              <p class="stat-footer mt-2 text-xs" :class="item.footerClass">
                {{ item.footer }}
              </p>
            </div>
            <div
              class="stat-icon flex items-center justify-center rounded-xl"
              :style="{ background: item.iconBg }">
              <component :is="item.icon" :style="{ color: item.iconColor, fontSize: '24px' }" />
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 中部区域：快捷操作 + 公告 -->
    <el-row :gutter="16" class="mb-5">
      <el-col :xs="24" :sm="24" :md="14">
        <div class="art-card p-5 h-full">
          <div class="flex items-center justify-between mb-4">
            <h4 class="section-title text-base font-semibold text-g-800">快捷操作</h4>
          </div>
          <el-row :gutter="12">
            <el-col :span="6" v-for="action in quickActions" :key="action.label">
              <div
                class="quick-action-item flex flex-col items-center gap-2 p-4 rounded-xl cursor-pointer transition-all hover:shadow-md"
                @click="handleQuickAction(action)">
                <div
                  class="quick-action-icon w-11 h-11 rounded-xl flex items-center justify-center"
                  :style="{ background: action.bgColor }">
                  <component :is="action.icon" :style="{ color: action.color, fontSize: '20px' }" />
                </div>
                <span class="text-xs text-g-600 text-center">{{ action.label }}</span>
              </div>
            </el-col>
          </el-row>
        </div>
      </el-col>

      <el-col :xs="24" :sm="24" :md="10">
        <div class="art-card p-5 h-full">
          <div class="flex items-center justify-between mb-4">
            <h4 class="section-title text-base font-semibold text-g-800">通知公告</h4>
            <el-button text type="primary" size="small" @click="goNotice">
              更多
              <ArtSvgIcon icon="ri:arrow-right-s-line" class="ml-0.5 text-sm" />
            </el-button>
          </div>
          <div v-if="notices.length === 0" class="empty-notice flex flex-col items-center justify-center py-8 text-g-400">
            <ArtSvgIcon icon="ri:notification-line" class="text-4xl mb-2" />
            <span class="text-sm">暂无公告</span>
          </div>
          <div v-else class="notice-list space-y-3">
            <div
              v-for="notice in notices"
              :key="notice.noticeId"
              class="notice-item flex items-start gap-3 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer transition-colors"
              @click="handleNoticeClick(notice)">
              <el-tag :type="getNoticeType(notice.noticeType)" size="small" effect="light" class="mt-0.5">
                {{ getNoticeLabel(notice.noticeType) }}
              </el-tag>
              <div class="flex-1 min-w-0">
                <p class="notice-title text-sm text-g-700 truncate">{{ notice.title }}</p>
                <p class="notice-time text-xs text-g-400 mt-1">{{ notice.createTime }}</p>
              </div>
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="16" class="mb-5">
      <el-col :xs="24" :sm="24" :md="12">
        <div class="art-card p-5">
          <div class="flex items-center justify-between mb-4">
            <h4 class="section-title text-base font-semibold text-g-800">访问量趋势</h4>
            <el-radio-group v-model="chartPeriod" size="small">
              <el-radio-button value="week">近7天</el-radio-button>
              <el-radio-button value="month">近30天</el-radio-button>
            </el-radio-group>
          </div>
          <div ref="lineChartRef" :style="{ height: '280px' }"></div>
        </div>
      </el-col>
      <el-col :xs="24" :sm="24" :md="12">
        <div class="art-card p-5">
          <div class="flex items-center justify-between mb-4">
            <h4 class="section-title text-base font-semibold text-g-800">系统资源监控</h4>
          </div>
          <div ref="gaugeChartRef" :style="{ height: '280px' }"></div>
        </div>
      </el-col>
    </el-row>

    <!-- 底部：技术栈信息 -->
    <el-row :gutter="16">
      <el-col :xs="24" :sm="24" :md="12">
        <div class="art-card p-5">
          <h4 class="section-title text-base font-semibold text-g-800 mb-4">系统信息</h4>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="系统名称">{{ systemName }}</el-descriptions-item>
            <el-descriptions-item label="当前版本">{{ version }}</el-descriptions-item>
            <el-descriptions-item label="后端框架">Go + Gin + GORM</el-descriptions-item>
            <el-descriptions-item label="前端框架">Vue 3 + Element Plus</el-descriptions-item>
            <el-descriptions-item label="Node 版本">
              <el-tag size="small" type="info">{{ nodeVersion }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="运行环境">{{ runEnv }}</el-descriptions-item>
          </el-descriptions>
        </div>
      </el-col>
      <el-col :xs="24" :sm="24" :md="12">
        <div class="art-card p-5">
          <h4 class="section-title text-base font-semibold text-g-800 mb-4">最近登录</h4>
          <el-timeline>
            <el-timeline-item
              v-for="(log, index) in recentLogs"
              :key="index"
              :timestamp="log.time"
              placement="top"
              size="normal">
              <div class="text-sm text-g-600">
                <span>{{ log.action }}</span>
                <el-tag v-if="log.status === 'success'" type="success" size="small" class="ml-2">成功</el-tag>
                <el-tag v-else type="danger" size="small" class="ml-2">失败</el-tag>
              </div>
              <p class="text-xs text-g-400 mt-1">{{ log.ip }} · {{ log.location }}</p>
            </el-timeline-item>
          </el-timeline>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick, markRaw, type Component } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/store/modules/user'
import { useSettingStore } from '@/store/modules/setting'
import { echarts, type EChartsOption } from '@plugins/echarts'
import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
import {
  User,
  Avatar,
  OfficeBuilding,
  List,
  SetUp,
  Bell,
  Grid,
  Ticket,
  Collection
} from '@element-plus/icons-vue'

defineOptions({ name: 'Index' })

const router = useRouter()
const userStore = useUserStore()
const settingStore = useSettingStore()

// ==================== 时间相关 ====================
const currentTime = ref('')
const chartPeriod = ref('week')

/** 问候语 */
const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return '夜深了'
  if (hour < 9) return '早上好'
  if (hour < 12) return '上午好'
  if (hour < 14) return '中午好'
  if (hour < 17) return '下午好'
  if (hour < 19) return '傍晚好'
  return '晚上好'
})

/** 今天日期字符串 */
const todayStr = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
})

/** 星期几 */
const weekDayStr = computed(() => {
  const days = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']
  return days[new Date().getDay()]
})

/** 角色标签 */
const roleLabel = computed(() => {
  if (userStore.roles.includes('admin')) return '管理员'
  if (userStore.roles.length > 0) return userStore.roles.join(', ')
  return '普通用户'
})

/** 系统名称 */
const systemName = 'RuoYi-Go 管理系统'
const version = 'v1.0.0'
const nodeVersion = 'v23.11.0'
const runEnv = import.meta.env.DEV ? '开发环境' : '生产环境'

// ==================== 统计卡片 ====================
interface StatCard {
  title: string
  value: string | number
  footer: string
  footerClass: string
  icon: Component
  iconBg: string
  iconColor: string
}

const statCards = computed<StatCard[]>(() => [
  {
    title: '用户数量',
    value: '128',
    footer: '较昨日 +12',
    footerClass: 'text-green-500',
    icon: markRaw(User),
    iconBg: 'rgba(64, 158, 255, 0.1)',
    iconColor: '#409EFF'
  },
  {
    title: '角色数量',
    value: '5',
    footer: '系统预设角色',
    footerClass: 'text-g-400',
    icon: markRaw(Avatar),
    iconBg: 'rgba(103, 194, 58, 0.1)',
    iconColor: '#67C23A'
  },
  {
    title: '部门数量',
    value: '12',
    footer: '组织架构部门',
    footerClass: 'text-g-400',
    icon: markRaw(OfficeBuilding),
    iconBg: 'rgba(230, 162, 60, 0.1)',
    iconColor: '#E6A23C'
  },
  {
    title: '岗位数量',
    value: '8',
    footer: '在招岗位',
    footerClass: 'text-g-400',
    icon: markRaw(List),
    iconBg: 'rgba(245, 108, 108, 0.1)',
    iconColor: '#F56C6C'
  }
])

// ==================== 快捷操作 ====================
interface QuickAction {
  label: string
  icon: Component
  color: string
  bgColor: string
  route?: string
}

const quickActions: QuickAction[] = [
  { label: '用户管理', icon: markRaw(User), color: '#409EFF', bgColor: 'rgba(64, 158, 255, 0.1)', route: '/system/user' },
  { label: '角色管理', icon: markRaw(Avatar), color: '#67C23A', bgColor: 'rgba(103, 194, 58, 0.1)', route: '/system/role' },
  { label: '菜单管理', icon: markRaw(Grid), color: '#E6A23C', bgColor: 'rgba(230, 162, 60, 0.1)', route: '/system/menu' },
  { label: '部门管理', icon: markRaw(OfficeBuilding), color: '#F56C6C', bgColor: 'rgba(245, 108, 108, 0.1)', route: '/system/dept' },
  { label: '岗位管理', icon: markRaw(Ticket), color: '#909399', bgColor: 'rgba(144, 147, 153, 0.1)', route: '/system/post' },
  { label: '字典管理', icon: markRaw(Collection), color: '#9B59B6', bgColor: 'rgba(155, 89, 182, 0.1)', route: '/system/dict' },
  { label: '参数设置', icon: markRaw(SetUp), color: '#1ABC9C', bgColor: 'rgba(26, 188, 156, 0.1)', route: '/system/config' },
  { label: '通知公告', icon: markRaw(Bell), color: '#E74C3C', bgColor: 'rgba(231, 76, 60, 0.1)', route: '/system/notice' }
]

/** 处理快捷操作点击 */
function handleQuickAction(action: QuickAction) {
  if (action.route) {
    router.push(action.route)
  }
}

// ==================== 通知公告 ====================
interface NoticeItem {
  noticeId: number
  title: string
  noticeType: string
  createTime: string
}

const notices = ref<NoticeItem[]>([])

function getNoticeType(type: string) {
  const map: Record<string, string> = { '1': 'success', '2': 'warning' }
  return (map[type] || 'info') as any
}

function getNoticeLabel(type: string) {
  const map: Record<string, string> = { '1': '通知', '2': '公告' }
  return map[type] || '通知'
}

function handleNoticeClick(notice: NoticeItem) {
  console.log('查看公告:', notice.noticeId)
}

function goNotice() {
  router.push('/system/notice')
}

// ==================== 最近登录 ====================
const recentLogs = ref([
  { action: '登录成功', status: 'success', time: '2026-10-04 09:15:23', ip: '192.168.1.100', location: '广东省深圳市' },
  { action: '登录成功', status: 'success', time: '2026-10-03 08:42:11', ip: '192.168.1.100', location: '广东省深圳市' },
  { action: '登录成功', status: 'success', time: '2026-10-02 09:03:45', ip: '192.168.1.101', location: '广东省广州市' },
  { action: '密码错误', status: 'fail', time: '2026-10-01 14:22:08', ip: '10.0.0.55', location: '内网IP' }
])

// ==================== 图表 ====================
const lineChartRef = ref<HTMLElement>()
const gaugeChartRef = ref<HTMLElement>()
let lineChart: echarts.ECharts | null = null
let gaugeChart: echarts.ECharts | null = null
let timeTimer: ReturnType<typeof setInterval> | null = null

/** 初始化时间更新 */
function initTimeUpdate() {
  updateTime()
  timeTimer = setInterval(updateTime, 1000)
}

function updateTime() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString('zh-CN', { hour12: false })
}

/** 初始化折线图 */
function initLineChart() {
  if (!lineChartRef.value) return
  lineChart = echarts.init(lineChartRef.value)

  const isDark = settingStore.isDark
  const textColor = isDark ? '#999' : '#666'

  const days = chartPeriod.value === 'week' ? 7 : 30
  const xAxisData: string[] = []
  const visitData: number[] = []
  const ipData: number[] = []

  for (let i = days - 1; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    xAxisData.push(`${d.getMonth() + 1}/${d.getDate()}`)
    visitData.push(Math.floor(Math.random() * 300 + 100))
    ipData.push(Math.floor(Math.random() * 150 + 50))
  }

  const options: EChartsOption = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: isDark ? '#1f1f1f' : '#fff',
      borderColor: isDark ? '#333' : '#eee',
      textStyle: { color: textColor }
    },
    legend: {
      data: ['访问量', '独立IP'],
      bottom: 0,
      textStyle: { color: textColor }
    },
    grid: { top: 10, right: 20, bottom: 35, left: 50 },
    xAxis: {
      type: 'category',
      data: xAxisData,
      axisLabel: { color: textColor, fontSize: 11 },
      axisLine: { lineStyle: { color: isDark ? '#333' : '#ddd' } }
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: textColor, fontSize: 11 },
      splitLine: { lineStyle: { color: isDark ? '#2a2a2a' : '#f0f0f0' } }
    },
    series: [
      {
        name: '访问量',
        type: 'line',
        smooth: true,
        data: visitData,
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.02)' }
          ])
        },
        lineStyle: { color: '#409EFF', width: 2 },
        itemStyle: { color: '#409EFF' }
      },
      {
        name: '独立IP',
        type: 'line',
        smooth: true,
        data: ipData,
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
            { offset: 1, color: 'rgba(103, 194, 58, 0.02)' }
          ])
        },
        lineStyle: { color: '#67C23A', width: 2 },
        itemStyle: { color: '#67C23A' }
      }
    ]
  }

  lineChart.setOption(options)
}

/** 初始化仪表盘 */
function initGaugeChart() {
  if (!gaugeChartRef.value) return
  gaugeChart = echarts.init(gaugeChartRef.value)

  const isDark = settingStore.isDark
  const textColor = isDark ? '#ccc' : '#666'

  const options: EChartsOption = {
    tooltip: { formatter: '{a} <br/>{b}: {c}%' },
    series: [
      {
        name: 'CPU使用率',
        type: 'gauge',
        center: ['25%', '55%'],
        radius: '75%',
        min: 0,
        max: 100,
        splitNumber: 5,
        axisLine: {
          lineStyle: {
            width: 10,
            color: [
              [0.3, '#67C23A'],
              [0.7, '#E6A23C'],
              [1, '#F56C6C']
            ]
          }
        },
        pointer: { width: 4, length: '60%' },
        axisTick: { distance: -10, length: 6, lineStyle: { color: '#auto' } },
        splitLine: { distance: -12, length: 10, lineStyle: { color: '#auto' } },
        axisLabel: { color: textColor, fontSize: 10, distance: -25 },
        detail: { valueAnimation: true, formatter: '{value}%', color: textColor, fontSize: 14, offsetCenter: [0, '70%'] },
        title: { offsetCenter: [0, '90%'], color: textColor, fontSize: 12 },
        data: [{ value: 35, name: 'CPU' }]
      },
      {
        name: '内存使用率',
        type: 'gauge',
        center: ['75%', '55%'],
        radius: '75%',
        min: 0,
        max: 100,
        splitNumber: 5,
        axisLine: {
          lineStyle: {
            width: 10,
            color: [
              [0.3, '#67C23A'],
              [0.7, '#E6A23C'],
              [1, '#F56C6C']
            ]
          }
        },
        pointer: { width: 4, length: '60%' },
        axisTick: { distance: -10, length: 6, lineStyle: { color: '#auto' } },
        splitLine: { distance: -12, length: 10, lineStyle: { color: '#auto' } },
        axisLabel: { color: textColor, fontSize: 10, distance: -25 },
        detail: { valueAnimation: true, formatter: '{value}%', color: textColor, fontSize: 14, offsetCenter: [0, '70%'] },
        title: { offsetCenter: [0, '90%'], color: textColor, fontSize: 12 },
        data: [{ value: 58, name: '内存' }]
      }
    ]
  }

  gaugeChart.setOption(options)
}

/** 窗口大小变化时重新调整图表 */
function handleResize() {
  lineChart?.resize()
  gaugeChart?.resize()
}

// 监听图表周期切换
watch(chartPeriod, () => {
  initLineChart()
})

// 监听主题变化
watch(
  () => settingStore.isDark,
  () => {
    nextTick(() => {
      initLineChart()
      initGaugeChart()
    })
  }
)

onMounted(() => {
  initTimeUpdate()
  nextTick(() => {
    initLineChart()
    initGaugeChart()
  })
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  if (timeTimer) clearInterval(timeTimer)
  lineChart?.dispose()
  gaugeChart?.dispose()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.workbench-container {
  min-height: 100%;
}

.welcome-banner {
  padding: 20px 24px;
  background: linear-gradient(135deg, var(--el-color-primary-light-9, #ecf5ff) 0%, var(--el-color-primary-light-7, #c6e2ff) 100%);
  border: 1px solid var(--el-color-primary-light-5, #a0cfff);

  .welcome-avatar {
    border: 3px solid #fff;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  }

  .welcome-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    margin: 0;
  }

  .welcome-subtitle {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin-top: 4px;
  }
}

.stat-card {
  padding: 18px 20px;
  transition: all 0.3s ease;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  }

  .stat-icon {
    width: 52px;
    height: 52px;
  }
}

.section-title {
  position: relative;
  padding-left: 10px;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 3px;
    height: 14px;
    border-radius: 2px;
    background: var(--el-color-primary);
  }
}

.quick-action-item {
  transition: all 0.25s ease;

  &:hover {
    background: var(--el-fill-color-light);
    transform: translateY(-2px);
  }
}

.notice-item {
  transition: all 0.2s ease;
}

.empty-notice {
  opacity: 0.6;
}
</style>
