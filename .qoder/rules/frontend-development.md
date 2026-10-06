# 前端开发规范（art-design-pro）

## 技术栈

| 项目 | 说明 |
| --- | --- |
| 核心框架 | Vue `3.5` + TypeScript `6.0` |
| 构建 | Vite `7.1` |
| UI 组件 | Element Plus `2.11`（按需自动导入，无需手动 import） |
| 状态管理 | Pinia `3.0`（持久化用 `pinia-plugin-persistedstate`） |
| HTTP | axios（统一封装 `@utils/http`） |
| 样式 | Tailwind CSS `4` + Sass |
| 包管理器 | **pnpm**（`>=10`），Node `>=20.19` |

## 常用脚本

```bash
pnpm dev        # 启动开发服务
pnpm build      # 类型检查 + 生产构建
pnpm check      # 仅类型检查（vue-tsc）
pnpm lint       # 代码检查（oxlint）
pnpm fmt        # 代码格式化（oxfmt）
```

## 目录结构约定

- 业务页面：`views/<模块>/<功能>/index.vue` + `api.ts` + `components/`
- 通用组件：`components/core/`（基础）或 `components/business/`（业务通用）
- 接口类型：`types/api/`，与后端响应结构对应
- Store：`store/modules/`
- 路由：`router/modules/`

## 请求层规范（`@utils/http`）

```ts
import response from '@utils/http'

// 列表查询（泛型标注返回类型）
export function listXxx(query: XxxQueryParams) {
  return response<TableDataInfo<Xxx>>({
    url: '/system/xxx/list',
    method: 'get',
    params: query
  })
}

// 增删改
export function addXxx(data: Xxx) {
  return response({ url: '/system/xxx', method: 'post', data })
}
```

- URL 与后端路由一致
- 列表返回 `TableDataInfo<T>`（含 `total` + `rows`），详情返回 `AjaxResult<T>`
- 自动注入 `Authorization: Bearer {token}`
- `code=200` 成功，`code=400` 抛错，`code=401` 自动登出

## 类型定义规范

- 放在 `types/api/<模块>/` 下，字段命名与后端 JSON 一致（驼峰）
- **布尔字段**：是/否类字段类型声明为 `boolean`，与后端 `bool` 直接对应，提交时**无需转 0/1**
- 通用类型在 `types/api/system/common.ts`：`AjaxResult<T>`、`TableDataInfo<T>`、`TreeSelect`

## 组件化架构

CRUD 页面**优先使用** `ArtCrud` 组件：

```vue
<template>
  <ArtCrud :config="crudConfig" />
</template>

<script setup lang="ts">
import ArtCrud from '@/components/business/art-crud/index.vue'
const crudConfig = ref({
  api: { list: listXxx, add: addXxx, update: updateXxx, delete: delXxx },
  columns: [...],
  search: { items: [...] },
  form: { items: [...] }
})
</script>
```

- 通用能力优先复用 `components/core/` 与 `components/business/`
- 独立业务子模块拆到 `views/<功能>/components/`

## 视图开发规范

- 使用 `<script setup lang="ts">` 组合式 API
- Element Plus / `@vueuse/core` / `vue-router` / `pinia` **无需手动 import**（自动注入）
- 业务状态放 `store/modules/`，页面只做展示与交互
- 别名：`@` = `src`、`@views`、`@utils`、`@stores`、`@styles` 等

## 路由与权限

- 路由在 `router/modules/` 对应模块注册
- 权限标识与后端菜单权限一致（如 `system:xxx:list`）
- 权限模式：`VITE_ACCESS_MODE = backend`（后端下发菜单与权限）

## 环境配置

- 开发环境：Vite 代理 `/dev-api` → 后端 `8080`
- 生产环境：后端托管前端静态文件，同源部署，无需额外代理

## 新增业务页面 Checklist

1. **类型**：`types/api/` 定义实体与查询参数接口
2. **接口**：`views/<模块>/xxx/api.ts` 封装 CRUD 接口调用
3. **页面**：`views/<模块>/xxx/index.vue`，优先用 `ArtCrud`
4. **路由**：`router/modules/` 注册路由（含权限）
5. **验证**：`pnpm check` 类型检查通过 + `pnpm dev` 联调
