<template>
  <div class="pagination-container" :class="{ 'is-background': background }">
    <ElPagination
      :total="total"
      :page-size="limit"
      :current-page="page"
      :page-sizes="pageSizes"
      :layout="layout"
      :background="background"
      :disabled="disabled"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange" />
  </div>
</template>

<script setup lang="ts">
import { ElPagination } from 'element-plus'

defineOptions({ name: 'Pagination' })

interface Props {
  /** 总条目数 */
  total: number
  /** 当前页码 */
  page: number
  /** 每页显示条目个数 */
  limit: number
  /** 每页显示个数选择器的选项列表 */
  pageSizes?: number[]
  /** 分页器布局 */
  layout?: string
  /** 是否使用背景色 */
  background?: boolean
  /** 是否禁用 */
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  pageSizes: () => [10, 20, 30, 50],
  layout: 'total, sizes, prev, pager, next, jumper',
  background: true,
  disabled: false
})

const emit = defineEmits<{
  (e: 'update:page', value: number): void
  (e: 'update:limit', value: number): void
  (e: 'pagination', page: number, limit: number): void
}>()

function handleSizeChange(size: number) {
  emit('update:limit', size)
  emit('pagination', props.page, size)
}

function handleCurrentChange(page: number) {
  emit('update:page', page)
  emit('pagination', page, props.limit)
}
</script>

<style lang="scss" scoped>
.pagination-container {
  padding: 12px 0;
  display: flex;
  justify-content: center;

  &.is-background {
    :deep(.el-pagination) {
      --el-pagination-button-bg-color: var(--el-fill-color-lighter);
    }
  }
}
</style>
