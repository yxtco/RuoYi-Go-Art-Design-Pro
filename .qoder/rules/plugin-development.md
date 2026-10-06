# 插件化开发规范

## 概述

项目采用**插件化开发机制**：每个独立功能以插件形式组织，前后端代码分别集中在各自的插件目录中。
核心系统模块（system / monitor / common / tool）仍保留在原有位置，插件机制适用于**新增的独立业务功能**。

**核心特性：自动注册** — 插件通过 Go 的 `init()` + 空白导入机制自动注册路由，
新增插件**无需修改** `service/init.go`、`api/init.go`，只需在 `router.go` 添加一行空白导入。

---

## 自动注册机制原理

```
启动流程：

1. Go 加载包 → 插件的 init() 自动调用 plugins.Register(SetupXxxRoutes)
2. SetupRouter() → setupPluginRoutes(r) → 遍历注册表 → 注册所有插件路由

核心文件（只需了解，无需修改）：
├── server/plugins/registry.go           # 插件注册表（提供 Register 函数）
└── server/internal/router/plugin_loader.go  # 插件加载器（遍历调用已注册的路由函数）
```

---

## 目录结构总览

```
server/                                       web/src/
├── plugins/                                  ├── views/plugins/
│   ├── registry.go          ← 注册表（已有）  │   └── fin/
│   └── fin/                                  │       ├── index.vue
│       ├── fin_model.go                      │       ├── api.ts
│       ├── fin_server.go                     │       └── components/
│       ├── fin_handler.go                    │           └── xxxDialog.vue
│       ├── fin_router.go    ← init() 自动注册 ├── types/api/plugins/
│       └── fin_api.go                        │   └── fin.ts
├── internal/                                 └── router/modules/
│   ├── router/plugin_loader.go  ← 自动加载         └── plugins.ts  ← 插件路由汇总
│   └── router/router.go  ← 仅需添加 1 行空白导入
└── migrations/  ← 插件迁移脚本（统一编号）
```

---

## 后端插件结构（`server/plugins/<name>/`）

以财务插件 `fin` 为例：

```
server/plugins/fin/
├── fin_model.go       # GORM 模型 + 数据访问方法
├── fin_server.go      # Service 业务逻辑层
├── fin_handler.go     # HTTP 处理器（含 Swagger 注解）
├── fin_router.go      # 路由注册 + init() 自动注册
└── fin_api.go         # 请求/响应结构体定义
```

### 文件职责

| 文件 | 职责 | 关键约定 |
| --- | --- | --- |
| `fin_model.go` | GORM 结构体 + 数据访问 | 实现 `TableName()`；表名加插件前缀（如 `fin_xxx`）避免冲突 |
| `fin_server.go` | Service 业务编排 | 值类型接收者 `type FinServer struct{}`；方法委托 Model |
| `fin_handler.go` | HTTP 处理器 | `SetOperTitle → 绑定参数 → 调 Service → 返回响应`；含 Swagger 注解 |
| `fin_router.go` | 路由注册 | **必须**包含 `init()` 调用 `plugins.Register()` 实现自动注册 |
| `fin_api.go` | 请求/响应结构体 | 命名：`FinXxxListRequest` / `FinXxxAddRequest` / `FinXxxIdUriRequest` |

### 后端代码示例

```go
// ===== fin_model.go =====
package fin

import (
    "go-fin-server/internal/db"
    "go-fin-server/internal/model"
    "go-fin-server/pkg/utils"
)

type FinAccount struct {
    Id      uint64  `gorm:"column:id;primary_key" json:"id" description:"账户ID"`
    Name    string  `gorm:"column:name" json:"name" description:"账户名称" binding:"required" err_msg:"账户名称 必填"`
    Balance float64 `gorm:"column:balance;default:0" json:"balance" description:"余额"`
    model.BaseModel
}

func (c FinAccount) TableName() string { return "fin_account" }

func (c FinAccount) SelectFinAccountList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]FinAccount, int64, error) {
    query := db.DBConnections["master"].Model(&FinAccount{})
    offset, limit := utils.PageParse(pageNum, pageSize)
    for key, value := range conditions {
        query = query.Where(key, value)
    }
    data := make([]FinAccount, 0)
    var count int64
    query.Count(&count)
    for _, condition := range sortConditions {
        query = query.Order(condition)
    }
    d := query.Offset(offset).Limit(limit).Find(&data)
    return data, count, d.Error
}

func (c FinAccount) InsertFinAccount(data *FinAccount) error {
    return db.DBConnections["master"].Create(data).Error
}

// UpdateFinAccount / DeleteFinAccountByIds / SelectFinAccountById ...
```

```go
// ===== fin_server.go =====
package fin

// FinServer 插件业务逻辑（直接委托 Model，无需注册到核心 Services）
type FinServer struct{}

func (s FinServer) SelectFinAccountList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]FinAccount, int64, error) {
    return FinAccount{}.SelectFinAccountList(conditions, sortConditions, pageNum, pageSize)
}

// 其他方法同理委托给 Model
```

```go
// ===== fin_handler.go =====
package fin

import (
    "go-fin-server/internal/response"
    "go-fin-server/pkg"
    "go-fin-server/pkg/utils"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"
)

// FinHandler 插件处理器（插件自包含，不依赖核心 Services/Api 注册）
type FinHandler struct {
    server FinServer  // 直接使用本插件的 Server
}

func NewFinHandler() *FinHandler { return &FinHandler{server: FinServer{}} }

// ListAccount 查询账户列表
//
//	@Summary	查询账户列表
//	@Tags		财务管理
//	@Produce	json
//	@Param		name		query	string	false	"账户名称"
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}
//	@Router		/fin/account/list [get]
//	@Security	BearerAuth
func (h *FinHandler) ListAccount(c *gin.Context) {
    response.SetOperTitle(c, "查询账户列表")
    // 绑定参数（插件可直接使用 api.ShouldBindError）
    // ...
    conditions := utils.BuildConditions(req)
    sortConditions := utils.BuildSortConditions(req.PageNum)
    data, total, err := h.server.SelectFinAccountList(conditions, sortConditions, req.PageNum, req.PageSize)
    if err != nil {
        pkg.Logger.Error(err)
        response.Error(c, err.Error())
        return
    }
    response.PageData(c, data, total)
}

// AddAccount / UpdateAccount / DelAccount ...
```

```go
// ===== fin_router.go =====
package fin

import (
    "go-fin-server/internal/middleware"
    "go-fin-server/plugins"

    "github.com/gin-gonic/gin"
)

// init 自动注册插件路由（Go 加载包时自动调用，无需手动注册）
func init() {
    plugins.Register(SetupFinRoutes)
}

// SetupFinRoutes 插件路由注册函数（由 plugin_loader 统一调用）
func SetupFinRoutes(r *gin.Engine) {
    h := NewFinHandler()
    group := r.Group("/fin")
    {
        group.GET("account/list", middleware.PermissionMiddleware("fin:account:list"), h.ListAccount)
        group.GET("account/:id", middleware.PermissionMiddleware("fin:account:query"), h.GetAccount)
        group.POST("account", middleware.PermissionMiddleware("fin:account:add"), h.AddAccount)
        group.PUT("account", middleware.PermissionMiddleware("fin:account:edit"), h.UpdateAccount)
        group.DELETE("account/:ids", middleware.PermissionMiddleware("fin:account:remove"), h.DelAccount)
    }
}
```

```go
// ===== fin_api.go =====
package fin

type FinAccountListRequest struct {
    Name     string `json:"name" form:"name"`
    PageNum  int    `json:"pageNum" form:"pageNum"`
    PageSize int    `json:"pageSize" form:"pageSize"`
}

type FinAccountAddRequest struct {
    Name  string  `json:"name" binding:"required" err_msg:"账户名称 必填"`
    Balance float64 `json:"balance"`
}

type FinAccountIdUriRequest struct {
    Id *uint64 `uri:"id" binding:"required" err_msg:"账户ID 必填"`
}
```

### 接入主项目（只需 1 行）

在 `server/internal/router/router.go` 的**插件空白导入区**添加一行：

```go
import (
    // ... 已有导入

    // ====== 插件空白导入区 ======
    _ "go-fin-server/plugins/fin"    // ← 新增这一行即可
    // ============================
)
```

> **无需修改** `service/init.go` 和 `api/init.go`。插件自包含 Handler + Server，不依赖核心注册。

---

## 前端插件结构（`web/src/views/plugins/<name>/`）

参照现有功能模块（如 `views/system/user/`）的组织方式：

```
web/src/views/plugins/fin/
├── index.vue              # 主页面（优先使用 ArtCrud 组件）
├── api.ts                 # 本插件接口调用
└── components/            # 插件内子组件（弹窗、详情等，可选）
    └── accountDialog.vue
```

### 配套文件

| 文件位置 | 职责 |
| --- | --- |
| `views/plugins/fin/index.vue` | 主页面，组合 ArtCrud 或自定义布局 |
| `views/plugins/fin/api.ts` | 封装接口调用（`listFinAccount` / `addFinAccount` 等） |
| `views/plugins/fin/components/` | 子组件（弹窗、详情面板等） |
| `types/api/plugins/fin.ts` | TypeScript 类型定义（与后端 JSON 字段对应） |
| `router/modules/plugins.ts` | 插件路由汇总（所有插件路由统一在此注册） |

### 前端代码示例

```ts
// ===== views/plugins/fin/api.ts =====
import response from '@utils/http'
import { FinAccount, FinAccountQueryParams } from '@/types/api/plugins/fin'
import { TableDataInfo } from '@/types/api/system/common'

export function listFinAccount(query: FinAccountQueryParams) {
  return response<TableDataInfo<FinAccount>>({
    url: '/fin/account/list', method: 'get', params: query
  })
}
export function addFinAccount(data: FinAccount) {
  return response({ url: '/fin/account', method: 'post', data })
}
export function updateFinAccount(data: FinAccount) {
  return response({ url: '/fin/account', method: 'put', data })
}
export function delFinAccount(id: number | number[]) {
  return response({ url: '/fin/account/' + id, method: 'delete' })
}
```

```vue
<!-- ===== views/plugins/fin/index.vue ===== -->
<template>
  <ArtCrud :config="crudConfig" />
</template>

<script setup lang="ts">
import ArtCrud from '@/components/business/art-crud/index.vue'
import { ref } from 'vue'
import { listFinAccount, addFinAccount, updateFinAccount, delFinAccount } from './api'

const crudConfig = ref({
  api: { list: listFinAccount, add: addFinAccount, update: updateFinAccount, delete: delFinAccount },
  columns: [
    { type: 'selection' as const },
    { type: 'index' as const, label: '序号', width: 60 },
    { prop: 'name', label: '账户名称', minWidth: 120 },
    { prop: 'balance', label: '余额', minWidth: 100 }
  ],
  search: {
    items: [{ key: 'name', label: '账户名称', type: 'input', placeholder: '请输入账户名称' }]
  },
  form: {
    items: [{ key: 'name', label: '账户名称', type: 'input', required: true }]
  }
})
</script>
```

```ts
// ===== types/api/plugins/fin.ts =====
export interface FinAccount {
  id?: number
  name?: string
  balance?: number
  createTime?: string
}

export interface FinAccountQueryParams {
  name?: string
  pageNum?: number
  pageSize?: number
}
```

```ts
// ===== router/modules/plugins.ts =====
import type { RuoYiAppRouteRecord } from '@/types/router'

export const pluginRoutes: RuoYiAppRouteRecord[] = [
  {
    path: '/fin',
    name: 'Fin',
    component: 'layout/index',
    permissions: ['fin:account:list'],
    meta: { title: '财务管理', icon: 'ri:money-cny-circle-line' },
    children: [
      { path: 'account', component: 'plugins/fin/index', name: 'FinAccount', meta: { title: '账户管理' } }
    ]
  }
]
```

---

## 数据库迁移

插件涉及数据库变更时：

1. 在 `server/migrations/` 中添加迁移文件（统一编号，与其他迁移一起管理）
2. 同步更新 `server/data/ry_init_full.sql`（全新数据库初始化脚本）
3. 表名建议加插件前缀（如 `fin_account`、`fin_record`）避免与核心表冲突

### 菜单迁移脚本规范

**强制规则：插件涉及前端菜单（目录 / 菜单 / 按钮）时，必须创建对应的迁移脚本。**

菜单数据写入 `sys_menu` 表，迁移脚本需同时包含 **up（插入菜单）** 和 **down（删除菜单）** 两个文件。

#### 菜单类型说明

| menu_type | 含义 | 说明 |
| --- | --- | --- |
| `M` | 目录 | 一级菜单目录，对应侧边栏分组 |
| `C` | 菜单 | 具体页面，对应前端路由 |
| `F` | 按钮 | 页面内操作按钮，对应权限标识 |

#### 字段约定

| 字段 | 说明 | 插件注意事项 |
| --- | --- | --- |
| `id` | 主键 | 使用较大的数字避免冲突（建议从 2000 起按序分配） |
| `menu_name` | 菜单名称 | 如「财务管理」「账户管理」 |
| `parent_id` | 父菜单ID | 一级目录为 `0`；子菜单/按钮的 parent_id 指向父菜单的 id |
| `order_num` | 显示顺序 | 从 `0` 开始递增 |
| `path` | 路由地址 | 目录：`fin`；菜单：`account`；按钮：留空 `''` |
| `component` | 组件路径 | 目录：`NULL`；菜单：`plugins/fin/index`；按钮：`NULL` |
| `menu_type` | 菜单类型 | `M` / `C` / `F` |
| `perms` | 权限标识 | 按钮必填，格式 `插件名:资源:动作`，如 `fin:account:list` |
| `icon` | 菜单图标 | 仅目录需要，如 `ri:money-cny-circle-line`；菜单/按钮填 `'#'` |
| `is_frame` | 是否外链 | 固定 `'1'`（否） |
| `is_cache` | 是否缓存 | `'0'`（缓存）或 `'1'`（不缓存） |
| `visible` | 显示状态 | `'0'`（显示） |
| `status` | 菜单状态 | `'0'`（正常） |

#### 迁移脚本示例

以财务插件 `fin` 为例，需要创建：目录（M）、菜单（C）、4 个按钮（F）。

**up 脚本**（`server/migrations/000003_add_fin_menu.up.sql`）：

```sql
-- 新增财务管理插件菜单
-- 使用 INSERT IGNORE：兼容已有数据库（避免重复插入报错）

-- 1. 一级目录：财务管理（menu_type = M）
INSERT IGNORE INTO sys_menu (id, menu_name, parent_id, order_num, path, component, query, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES (2000, '财务管理', 0, 10, 'fin', NULL, '', '1', '0', 'M', '0', '0', '', 'ri:money-cny-circle-line', 'admin', NOW(), '', NULL, '财务管理插件目录');

-- 2. 二级菜单：账户管理（menu_type = C，parent_id 指向目录 id）
INSERT IGNORE INTO sys_menu (id, menu_name, parent_id, order_num, path, component, query, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES (2001, '账户管理', 2000, 0, 'account', 'plugins/fin/index', '', '1', '0', 'C', '0', '0', '', '#', 'admin', NOW(), '', NULL, '账户管理菜单');

-- 3. 按钮权限（menu_type = F，parent_id 指向菜单 id）
INSERT IGNORE INTO sys_menu (id, menu_name, parent_id, order_num, path, component, query, is_frame, is_cache, menu_type, visible, status, perms, icon, create_by, create_time, update_by, update_time, remark)
VALUES
    (2002, '账户查询', 2001, 0, '', NULL, '', '1', '0', 'F', '0', '0', 'fin:account:query', '#', 'admin', NOW(), '', NULL, ''),
    (2003, '账户新增', 2001, 1, '', NULL, '', '1', '0', 'F', '0', '0', 'fin:account:add', '#', 'admin', NOW(), '', NULL, ''),
    (2004, '账户修改', 2001, 2, '', NULL, '', '1', '0', 'F', '0', '0', 'fin:account:edit', '#', 'admin', NOW(), '', NULL, ''),
    (2005, '账户删除', 2001, 3, '', NULL, '', '1', '0', 'F', '0', '0', 'fin:account:remove', '#', 'admin', NOW(), '', NULL, '');
```

**down 脚本**（`server/migrations/000003_add_fin_menu.down.sql`）：

```sql
-- 回滚：删除财务管理插件菜单（按 id 删除，子菜单和按钮一并清除）
DELETE FROM sys_menu WHERE id IN (2000, 2001, 2002, 2003, 2004, 2005);
```

#### ID 分配建议

- 每个插件分配一个 **独立的 ID 段**，避免多插件菜单 ID 冲突
- 建议按插件序号分配：第一个插件从 `2000` 起，第二个从 `3000` 起，以此类推
- 同一插件内：目录占 1 个 ID，菜单占 1 个 ID，按钮依次递增
- down 脚本中使用 `IN (...)` 列出所有 ID，确保回滚干净

---

## 命名约定

| 项目 | 约定 | 示例 |
| --- | --- | --- |
| 插件目录名 | 小写英文，简短有意义 | `fin`、`asset`、`crm` |
| 后端文件名 | `<插件名>_层名.go` | `fin_model.go`、`fin_handler.go` |
| 数据库表名 | `<插件名>_业务名` | `fin_account`、`fin_record` |
| 路由路径 | `/插件名/资源` | `/fin/account/list` |
| 权限标识 | `插件名:资源:动作` | `fin:account:list` |

---

## 新增插件完整 Checklist

### 后端（3 步）

1. **创建插件目录**：`server/plugins/<name>/` 下创建 5 个文件
   - `<name>_model.go` — GORM 模型 + 数据访问
   - `<name>_server.go` — Service 业务逻辑
   - `<name>_handler.go` — HTTP 处理器（含 Swagger 注解）
   - `<name>_router.go` — **必须包含 `init()` + `plugins.Register()`**
   - `<name>_api.go` — 请求/响应结构体
2. **空白导入**：在 `router.go` 的插件导入区添加 `_ "go-fin-server/plugins/<name>"`
3. **编译验证**：`go build ./...` 通过

### 前端（3 步）

4. **创建前端目录**：`views/plugins/<name>/` 下创建 index.vue + api.ts
5. **类型定义**：`types/api/plugins/<name>.ts`
6. **路由注册**：`router/modules/plugins.ts` 中添加路由记录

### 数据库（必选）

7. **迁移脚本**：`server/migrations/` 添加 up/down SQL（含业务表 + 菜单数据）
8. **菜单迁移**：涉及前端菜单时，必须创建菜单迁移脚本（目录 M + 菜单 C + 按钮 F）
9. **更新初始化**：同步更新 `server/data/ry_init_full.sql`

### 验证

10. `go build ./...` 编译通过
11. `pnpm check` 类型检查通过
12. 启动服务，访问页面联调接口
