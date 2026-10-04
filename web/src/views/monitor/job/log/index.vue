<template>
  <div>
    <ArtCrud
      ref="crudRef"
      :config="crudConfig"
      v-model:search="queryParams"
      @selection-change="handleSelectionChange">
      <template #toolbar>
        <el-button
          v-auth="'monitor:job:remove'"
          type="danger"
          plain
          :icon="Delete"
          :disabled="multiple"
          @click="handleDelete">
          删除
        </el-button>
        <el-button
          v-auth="'monitor:job:remove'"
          type="danger"
          plain
          :icon="Delete"
          @click="handleClean">
          清空
        </el-button>
        <el-button
          v-auth="'monitor:job:export'"
          type="warning"
          plain
          :icon="Download"
          @click="handleExport">
          导出
        </el-button>
        <el-button type="warning" plain :icon="Close" @click="handleClose">
          关闭
        </el-button>
      </template>

      <template #col-jobGroup="{ row }">
        <dict-tag :options="sys_job_group" :value="row.jobGroup" />
      </template>

      <template #col-status="{ row }">
        <dict-tag :options="sys_common_status" :value="row.status" />
      </template>

      <template #col-createTime="{ row }">
        <span>{{ parseTime(row.createTime) }}</span>
      </template>

      <template #col-operationTemplate="{ row }">
        <ArtButtonTable
          v-auth="'monitor:job:query'"
          type="view"
          @click="handleView(row)" />
      </template>
    </ArtCrud>

    <JobDetail v-model:visible="open" :row="form" type="log" />
  </div>
</template>

<script setup lang="ts">
import type { ArtCrudConfig } from '@/components/business/art-crud/index.vue'
import type { SysJob } from '@/types/api/monitor/job'
import type { JobLogQueryParams, SysJobLog } from '@/types/api/monitor/jobLog'

import { Close, Delete, Download } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

import ArtCrud from '@/components/business/art-crud/index.vue'
import { useDict } from '@/hooks/core/useDict'
import { useWorktabStore } from '@/store/modules/worktab'
import download from '@utils/http/download'
import { addDateRange, parseTime } from '@utils/sys/ruoyi'

import JobDetail from '../components/detail.vue'
import { getJob } from '../index/api'
import { cleanJobLog, delJobLog, listJobLog } from './api'

defineOptions({ name: 'JobLog' })

const route = useRoute()
const router = useRouter()
const worktabStore = useWorktabStore()
const { sys_common_status, sys_job_group } = useDict(
  'sys_common_status',
  'sys_job_group'
)

const crudRef = useTemplateRef<InstanceType<typeof ArtCrud>>('crudRef')
const ids = ref<number[]>([])
const multiple = ref<boolean>(true)
const open = ref<boolean>(false)
const form = ref<SysJobLog>({})
const initialized = ref<boolean>(false)

const queryParams = ref<JobLogQueryParams & { createTime?: string[] }>({
  jobName: undefined,
  jobGroup: undefined,
  status: undefined,
  createTime: []
})

function handleClose() {
  const currentPath = route.path
  worktabStore.removeTab(currentPath)
  const jobTab = worktabStore.getTab('/monitor/job')

  if (jobTab) {
    router.push({ path: jobTab.path, query: jobTab.query })
  }
}

function handleSelectionChange(selection: SysJobLog[]) {
  ids.value = selection.map((item) => item.jobLogId!)
  multiple.value = !selection.length
}

function handleView(row: SysJobLog) {
  form.value = row
  open.value = true
}

function handleDelete() {
  if (!ids.value.length) return
  ElMessageBox.confirm(`是否确认删除调度日志编号为"${ids.value}"的数据项?`)
    .then(() => delJobLog(ids.value))
    .then(() => {
      refreshTable()
      ElMessage.success('删除成功')
    })
    .catch(() => {})
}

function handleClean() {
  ElMessageBox.confirm('是否确认清空所有调度日志数据项?')
    .then(() => cleanJobLog())
    .then(() => {
      refreshTable()
      ElMessage.success('清空成功')
    })
    .catch(() => {})
}

function handleExport() {
  download(
    'monitor/jobLog/export',
    {
      ...queryParams.value
    },
    `job_log_${new Date().getTime()}.xlsx`
  )
}

function refreshTable() {
  crudRef.value?.handleRefresh()
}

async function initByRoute() {
  const jobId = Number(route.params.jobId || 0)
  if (jobId) {
    const job = (await getJob(jobId)) as SysJob
    queryParams.value.jobName = job.jobName
    queryParams.value.jobGroup = job.jobGroup
  }
  initialized.value = true
  await nextTick()
  refreshTable()
}

const crudConfig: any = computed<ArtCrudConfig<SysJobLog>>(() => ({
  api: {
    list: (params) => {
      if (!initialized.value) {
        return Promise.resolve({ rows: [], total: 0 })
      }
      const query = addDateRange(
        {
          ...queryParams.value,
          pageNum: params.pageNum,
          pageSize: params.pageSize
        },
        queryParams.value.createTime as string[]
      )
      delete query.createTime
      return listJobLog(query)
    }
  },
  searchConfig: {
    items: [
      {
        label: '任务名称',
        key: 'jobName',
        type: 'input',
        placeholder: '请输入任务名称'
      },
      {
        label: '任务组名',
        key: 'jobGroup',
        type: 'select',
        placeholder: '请选择任务组名',
        options: sys_job_group.value
      },
      {
        label: '执行状态',
        key: 'status',
        type: 'select',
        placeholder: '请选择执行状态',
        options: sys_common_status.value
      },
      {
        label: '执行时间',
        key: 'createTime',
        type: 'daterange',
        props: {
          type: 'daterange',
          valueFormat: 'YYYY-MM-DD',
          rangeSeparator: '-',
          startPlaceholder: '开始日期',
          endPlaceholder: '结束日期'
        }
      }
    ]
  },
  columns: [
    { type: 'selection', width: 55, align: 'center' },
    { label: '日志编号', prop: 'jobLogId', width: 80, align: 'center' },
    {
      label: '任务名称',
      prop: 'jobName',
      align: 'center',
      showOverflowTooltip: true
    },
    {
      label: '任务组名',
      prop: 'jobGroup',
      align: 'center',
      showOverflowTooltip: true,
      useSlot: true,
      slotName: 'jobGroup'
    },
    {
      label: '调用目标字符串',
      prop: 'invokeTarget',
      align: 'center',
      showOverflowTooltip: true
    },
    {
      label: '日志信息',
      prop: 'jobMessage',
      align: 'center',
      showOverflowTooltip: true
    },
    {
      label: '执行状态',
      prop: 'status',
      align: 'center',
      useSlot: true,
      slotName: 'status'
    },
    {
      label: '执行时间',
      prop: 'createTime',
      width: 180,
      align: 'center',
      useSlot: true,
      slotName: 'createTime'
    },
    {
      label: '操作',
      prop: 'operation',
      align: 'center',
      useSlot: true,
      slotName: 'operationTemplate'
    }
  ],
  btnConfig: {
    showAddOperation: () => false,
    showEditOperation: () => false,
    showDeleteOperation: () => false
  }
}))

onMounted(() => {
  initByRoute()
})
</script>
