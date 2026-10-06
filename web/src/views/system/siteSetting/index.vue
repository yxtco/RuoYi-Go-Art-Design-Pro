<template>
  <div class="site-setting-page">
    <el-card shadow="never" class="mb-4">
      <el-tabs v-model="activeTab" tab-position="right" class="site-setting-tabs">
        <!-- 基础设置 -->
        <el-tab-pane label="基础设置" name="base">
          <el-form :model="form" label-width="110px" label-position="right">
            <el-form-item label="网站名称">
              <el-input
                v-model="form.siteName"
                placeholder="显示在登录页、侧边栏、顶部与浏览器标题"
                maxlength="50" />
            </el-form-item>

            <el-form-item label="网站 Logo">
              <div class="flex-c gap-3">
                <img
                  :src="form.siteLogo || defaultLogo"
                  class="h-10 w-10 rounded border border-[var(--art-card-border)] object-contain" />
                <el-upload
                  :show-file-list="false"
                  :http-request="(opt: any) => handleUpload(opt, 'siteLogo')"
                  accept="image/*">
                  <el-button>选择图片</el-button>
                </el-upload>
                <el-input
                  v-model="form.siteLogo"
                  placeholder="Logo 图片地址，留空用默认 Logo"
                  class="!w-64" />
              </div>
            </el-form-item>

            <el-form-item label="网站图标">
              <el-input v-model="form.siteFavicon" placeholder="favicon 图片地址，可选" />
            </el-form-item>

            <el-form-item label="备案号">
              <el-input v-model="form.siteRecordNo" placeholder="ICP 备案号，如：鄂ICP备2026000000号" />
            </el-form-item>

            <el-form-item label="版权信息">
              <el-input v-model="form.siteCopyright" placeholder="页面底部版权信息，可选" />
            </el-form-item>

            <el-form-item label="网站描述">
              <el-input
                v-model="form.siteDescription"
                type="textarea"
                :rows="3"
                placeholder="网站描述，可选" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 登录设置 -->
        <el-tab-pane label="登录设置" name="login">
          <el-form :model="form" label-width="110px" label-position="right">
            <el-divider content-position="left">登录页内容</el-divider>
            <el-form-item label="登录页标题">
              <el-input v-model="form.loginTitle" placeholder="登录页左侧大标题，留空用默认文案" />
            </el-form-item>

            <el-form-item label="登录页副标题">
              <el-input v-model="form.loginSubtitle" placeholder="登录页左侧副标题，留空用默认文案" />
            </el-form-item>

            <el-form-item label="登录页背景图">
              <div class="flex-c gap-3">
                <img
                  v-if="form.loginBackground"
                  :src="form.loginBackground"
                  class="h-12 w-24 rounded border border-[var(--art-card-border)] object-cover" />
                <el-upload
                  :show-file-list="false"
                  :http-request="(opt: any) => handleUpload(opt, 'loginBackground')"
                  accept="image/*">
                  <el-button>选择图片</el-button>
                </el-upload>
                <el-input
                  v-model="form.loginBackground"
                  placeholder="登录页背景图地址，可选"
                  class="!w-64" />
              </div>
            </el-form-item>

            <el-form-item label="登录页版权">
              <el-input v-model="form.loginCopyright" placeholder="登录页底部版权文字，可选" />
            </el-form-item>

            <el-divider content-position="left">登录安全</el-divider>
            <el-form-item label="启用验证码">
              <el-switch
                :model-value="captchaOn"
                @change="(v: boolean | string | number) => (form.captchaEnabled = v ? 'true' : 'false')" />
              <span class="ml-2 text-g-600">登录时是否显示并校验图形验证码</span>
            </el-form-item>

            <el-form-item label="开放注册">
              <el-switch
                :model-value="registerOn"
                @change="(v: boolean | string | number) => (form.registerUser = v ? 'true' : 'false')" />
              <span class="ml-2 text-g-600">是否允许用户自助注册账号</span>
            </el-form-item>

            <el-form-item label="登录IP黑名单">
              <el-input
                v-model="form.blackIPList"
                type="textarea"
                :rows="3"
                placeholder="多个 IP 用 ; 分隔，支持 * 通配与网段，如：192.168.1.10;192.168.1.*" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 日志设置 -->
        <el-tab-pane label="日志设置" name="log">
          <el-form :model="form" label-width="110px" label-position="right">
            <el-form-item label="日志级别">
              <el-radio-group v-model="form.logLevel">
                <el-radio-button value="quiet">
                  <div class="flex-c flex-col items-center">
                    <span>安静</span>
                    <span class="text-xs text-g-500">仅错误</span>
                  </div>
                </el-radio-button>
                <el-radio-button value="standard">
                  <div class="flex-c flex-col items-center">
                    <span>标准</span>
                    <span class="text-xs text-g-500">常规信息</span>
                  </div>
                </el-radio-button>
                <el-radio-button value="detailed">
                  <div class="flex-c flex-col items-center">
                    <span>详细</span>
                    <span class="text-xs text-g-500">调试模式</span>
                  </div>
                </el-radio-button>
              </el-radio-group>
              <div class="mt-2 text-g-600 text-xs">
                <template v-if="form.logLevel === 'quiet'">安静模式：仅输出 ERROR 及以上级别日志，适合生产环境</template>
                <template v-else-if="form.logLevel === 'standard'">标准模式：输出 INFO/WARN/ERROR 日志，适合日常运行</template>
                <template v-else>详细模式：输出 DEBUG 及以上全部日志，适合排查问题</template>
              </div>
            </el-form-item>

            <el-form-item label="日志格式">
              <el-radio-group v-model="form.logFormat">
                <el-radio-button value="json">JSON</el-radio-button>
                <el-radio-button value="console">控制台文本</el-radio-button>
              </el-radio-group>
              <div class="mt-2 text-g-600 text-xs">
                <template v-if="form.logFormat === 'json'">JSON 格式：结构化输出，便于日志采集与分析</template>
                <template v-else>控制台文本：可读性强，便于开发调试</template>
              </div>
            </el-form-item>

            <el-alert
              title="日志级别调整后即时生效，无需重启服务。日志格式变更需重启服务后生效。"
              type="info"
              :closable="false"
              show-icon />
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <div class="flex-c gap-3">
      <el-button type="primary" :loading="saving" @click="handleSave">保存设置</el-button>
      <el-button @click="handleReset">重置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import type { UploadRequestOptions } from 'element-plus'

import defaultLogo from '@imgs/common/logo.webp'
import { fetchSiteSetting, uploadSiteFile } from '@/api/system/siteSetting'
import { useSiteStore } from '@/store/modules/site'
import type { SiteSetting } from '@/types/api/system/siteSetting'

defineOptions({ name: 'SiteSetting' })

const siteStore = useSiteStore()

const activeTab = ref('base')

const form = reactive<SiteSetting>({
  siteName: '',
  siteLogo: '',
  siteFavicon: '',
  siteRecordNo: '',
  siteCopyright: '',
  siteDescription: '',
  loginTitle: '',
  loginSubtitle: '',
  loginBackground: '',
  loginCopyright: '',
  captchaEnabled: 'true',
  registerUser: 'false',
  blackIPList: '',
  logLevel: 'standard',
  logFormat: 'json'
})

// 开关绑定（form 存字符串 "true"/"false"，switch 需要布尔）
const captchaOn = computed(() => form.captchaEnabled === 'true')
const registerOn = computed(() => form.registerUser === 'true')

const saving = ref(false)

async function load() {
  try {
    const data = await fetchSiteSetting()
    if (data) {
      Object.assign(form, data)
    } else {
      Object.assign(form, {
        siteName: siteStore.name,
        siteLogo: siteStore.logo,
        siteFavicon: siteStore.favicon,
        siteRecordNo: siteStore.recordNo,
        siteCopyright: siteStore.copyright,
        siteDescription: siteStore.description,
        loginTitle: siteStore.loginTitle,
        loginSubtitle: siteStore.loginSubtitle,
        loginBackground: siteStore.loginBackground,
        loginCopyright: siteStore.loginCopyright,
        captchaEnabled: siteStore.captchaEnabled,
        registerUser: siteStore.registerUser,
        blackIPList: siteStore.blackIPList,
        logLevel: siteStore.logLevel,
        logFormat: siteStore.logFormat
      })
    }
  } catch {
    ElMessage.error('获取网站设置失败')
  }
}

async function handleUpload(opt: UploadRequestOptions, field: 'siteLogo' | 'loginBackground') {
  try {
    const res = await uploadSiteFile(opt.file)
    if (res?.url) {
      form[field] = res.url
      ElMessage.success('图片上传成功')
    } else {
      ElMessage.error('图片上传失败')
    }
  } catch {
    ElMessage.error('图片上传失败')
  }
}

async function handleSave() {
  saving.value = true
  try {
    await siteStore.saveSite({ ...form })
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

function handleReset() {
  load()
}

onMounted(load)
</script>

<style lang="scss" scoped>
.site-setting-page {
  .site-setting-tabs {
    min-height: 320px;

    :deep(.el-tabs__content) {
      padding-right: 8px;
    }
  }

  .el-form-item {
    margin-bottom: 18px;
  }
}
</style>
