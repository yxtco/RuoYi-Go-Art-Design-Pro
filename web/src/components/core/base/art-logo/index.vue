<!-- 系统logo（支持传入数据库配置的站点 Logo，未配置时使用默认图） -->
<template>
  <div class="flex-cc">
    <img
      :style="logoStyle"
      :src="src || defaultLogo"
      alt="logo"
      class="h-full w-full" />
  </div>
</template>

<script setup lang="ts">
import defaultLogo from '../../../../assets/images/common/logo.webp'

defineOptions({ name: 'ArtLogo' })

interface Props {
  /** logo 大小 */
  size?: number | string
  /** 站点 Logo 图片地址（网站设置写入数据库），为空时用默认 logo */
  src?: string
}

const props = withDefaults(defineProps<Props>(), {
  size: 36,
  src: ''
})

const logoStyle = computed(() => ({ width: `${props.size}px` }))
</script>
