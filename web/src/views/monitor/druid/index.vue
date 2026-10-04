<template>
  <div class="app-container druid-sql-monitor">
    <!-- 统计概览卡片 -->
    <el-row :gutter="16">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">SQL 总执行次数</div>
            <div class="stat-value">{{ stats.totalCount }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">总执行时间(ms)</div>
            <div class="stat-value">{{ stats.totalDuration }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-item">
            <div class="stat-label">平均执行时间(ms)</div>
            <div class="stat-value">{{ stats.avgDuration }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card stat-card--warning">
          <div class="stat-item">
            <div class="stat-label">慢查询次数(≥100ms)</div>
            <div class="stat-value text-danger">{{ stats.slowCount }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- SQL 类型分布 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span><el-icon><PieChart /></el-icon> SQL 类型执行次数分布</span>
          </template>
          <el-table :data="typeTableData" border size="small" style="width: 100%">
            <el-table-column label="SQL 类型" width="120">
              <template #default="{ row }">
                <el-tag :type="row.tagType" size="small">{{ row.name }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="执行次数" prop="count" />
            <el-table-column label="累计耗时(ms)" prop="duration" />
            <el-table-column label="平均耗时(ms)">
              <template #default="{ row }">
                {{ calcAvg(row.duration, row.count) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span><el-icon><DataLine /></el-icon> 执行次数占比</span>
            <el-button style="float: right; padding: 3px 0" text :icon="Refresh" @click="getList">
              刷新
            </el-button>
          </template>
          <div class="type-bar-container">
            <div v-for="item in typeDistribution" :key="item.name" class="type-bar-item">
              <span class="type-bar-label">{{ item.name }}</span>
              <el-progress
                :percentage="item.percentage"
                :color="item.color"
                :stroke-width="18"
                :text-inside="true"
                style="flex: 1; margin: 0 12px"
              />
              <span class="type-bar-count">{{ item.count }} 次</span>
            </div>
            <el-empty v-if="typeDistribution.length === 0" description="暂无数据" :image-size="60" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近 SQL 执行记录 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span><el-icon><Document /></el-icon> 最近 SQL 执行记录</span>
            <el-button style="float: right; padding: 3px 0" text :icon="Delete" @click="handleClear">
              清空统计
            </el-button>
          </template>
          <el-table :data="stats.recent || []" border size="small" style="width: 100%"
                    :row-class-name="tableRowClassName">
            <el-table-column label="执行时间" prop="time" width="160" />
            <el-table-column label="类型" prop="type" width="90">
              <template #default="{ row }">
                <el-tag :type="getTagType(row.type)" size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="耗时(ms)" prop="duration" width="100">
              <template #default="{ row }">
                <span :class="{ 'text-danger': row.duration >= 100 }">{{ row.duration }}</span>
              </template>
            </el-table-column>
            <el-table-column label="影响行数" prop="rows" width="90" />
            <el-table-column label="SQL 语句" prop="sql" show-overflow-tooltip>
              <template #default="{ row }">
                <code class="sql-code">{{ row.sql }}</code>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 慢查询记录 -->
    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span><el-icon><WarningFilled /></el-icon> 慢查询记录（≥100ms）</span>
            <span style="float: right; color: #909399; font-size: 13px">共 {{ (stats.slow || []).length }} 条</span>
          </template>
          <el-table :data="stats.slow || []" border size="small" style="width: 100%"
                    row-class-name="slow-row">
            <el-table-column label="执行时间" prop="time" width="160" />
            <el-table-column label="类型" prop="type" width="90">
              <template #default="{ row }">
                <el-tag :type="getTagType(row.type)" size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="耗时(ms)" prop="duration" width="100">
              <template #default="{ row }">
                <span class="text-danger" style="font-weight: bold">{{ row.duration }}</span>
              </template>
            </el-table-column>
            <el-table-column label="影响行数" prop="rows" width="90" />
            <el-table-column label="SQL 语句" prop="sql" show-overflow-tooltip>
              <template #default="{ row }">
                <code class="sql-code sql-code--danger">{{ row.sql }}</code>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="(stats.slow || []).length === 0" description="暂无慢查询记录" :image-size="60" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { PieChart, DataLine, Document, Delete, Refresh, WarningFilled } from '@element-plus/icons-vue'

import { getSQLStats, clearSQLStats, type SqlStats, type SqlRecord } from './api'

defineOptions({ name: 'Druid' })

const stats = ref<SqlStats>({
  totalCount: 0,
  totalDuration: 0,
  avgDuration: 0,
  slowCount: 0,
  selectCount: 0,
  insertCount: 0,
  updateCount: 0,
  deleteCount: 0,
  selectDuration: 0,
  insertDuration: 0,
  updateDuration: 0,
  deleteDuration: 0,
  recent: [],
  slow: []
})

// 自动刷新定时器
let refreshTimer: ReturnType<typeof setInterval> | null = null

/** SQL 类型分布数据（表格用） */
const typeTableData = computed(() => [
  { name: 'SELECT', count: stats.value.selectCount, duration: stats.value.selectDuration, tagType: 'success' },
  { name: 'INSERT', count: stats.value.insertCount, duration: stats.value.insertDuration, tagType: 'info' },
  { name: 'UPDATE', count: stats.value.updateCount, duration: stats.value.updateDuration, tagType: 'warning' },
  { name: 'DELETE', count: stats.value.deleteCount, duration: stats.value.deleteDuration, tagType: 'danger' }
])

/** SQL 类型执行次数分布（进度条用） */
const typeDistribution = computed(() => {
  const total = stats.value.selectCount + stats.value.insertCount + stats.value.updateCount + stats.value.deleteCount
  if (total === 0) return []
  const items = [
    { name: 'SELECT', count: stats.value.selectCount, color: '#67C23A' },
    { name: 'INSERT', count: stats.value.insertCount, color: '#909399' },
    { name: 'UPDATE', count: stats.value.updateCount, color: '#E6A23C' },
    { name: 'DELETE', count: stats.value.deleteCount, color: '#F56C6C' }
  ]
  return items
    .filter(item => item.count > 0)
    .map(item => ({
      ...item,
      percentage: Math.round((item.count / total) * 100)
    }))
})

/** 查询 SQL 监控数据 */
async function getList() {
  try {
    const res = await getSQLStats()
    stats.value = res
  } catch (error) {
    console.error('[Druid] 获取 SQL 监控数据失败:', error)
  }
}

/** 清空统计数据 */
function handleClear() {
  ElMessageBox.confirm('确认清空所有 SQL 监控统计数据？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    await clearSQLStats()
    getList()
    ElMessage.success('清空成功')
  }).catch(() => {})
}

/** 根据 SQL 类型返回 Tag 颜色 */
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger'
function getTagType(type: string): TagType {
  const map: Record<string, TagType> = { SELECT: 'success', INSERT: 'info', UPDATE: 'warning', DELETE: 'danger' }
  return map[type] || 'info'
}

/** 表格行样式 */
function tableRowClassName({ row }: { row: SqlRecord }) {
  if (row.duration >= 100) return 'warning-row'
  return ''
}

/** 计算平均耗时 */
function calcAvg(duration: number, count: number) {
  if (!count || count === 0) return '0'
  return (duration / count).toFixed(2)
}

onMounted(() => {
  getList()
  // 每 10 秒自动刷新
  refreshTimer = setInterval(() => {
    getList()
  }, 10000)
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style lang="scss" scoped>
.stat-card {
  text-align: center;
  .stat-item {
    padding: 8px 0;
  }
  .stat-label {
    font-size: 13px;
    color: #909399;
    margin-bottom: 8px;
  }
  .stat-value {
    font-size: 28px;
    font-weight: bold;
    color: #303133;
  }
  &--warning .stat-value {
    color: #f56c6c;
  }
}

.type-bar-container {
  padding: 8px 0;
}
.type-bar-item {
  display: flex;
  align-items: center;
  margin-bottom: 14px;
}
.type-bar-label {
  width: 60px;
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}
.type-bar-count {
  width: 60px;
  text-align: right;
  font-size: 13px;
  color: #909399;
}

.text-danger {
  color: #f56c6c;
}

.sql-code {
  font-size: 12px;
  color: #606266;
  &--danger {
    color: #f56c6c;
  }
}
</style>

<style lang="scss">
.druid-sql-monitor {
  .el-table .warning-row {
    background: #fdf6ec;
  }
  .el-table .slow-row {
    background: #fef0f0;
  }
}
</style>
