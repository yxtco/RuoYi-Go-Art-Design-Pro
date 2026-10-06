# 前后端开发操作示例

> 以新增「客户管理」模块为例，演示从零到可用的完整开发过程。

## 总体流程对照

```
后端（server/）                                  前端（web/src/）
─────────────────────────                       ─────────────────────────
1. model/customer.go       ←  数据库表          1. types/api/system/customer.ts
2. service/customer.go     ──────────────────   2. views/system/customer/api.ts
   （并在 service/init.go 注册）                 3. views/system/customer/index.vue
3. handler/customer.go                          4. router/modules/system.ts 注册路由
4. api/customer.go（并在 api/init.go 注册）      ─────────────────────────
5. router/customer.go（并在 router.go 注册）     5. pnpm dev 联调
6. go build 编译通过                             6. pnpm check 类型检查
```

## 后端各层代码模板

### Model 层（`internal/model/customer.go`）

```go
package model

import (
    "go-fin-server/internal/db"
    "go-fin-server/pkg/types"
    "go-fin-server/pkg/utils"
)

type Customer struct {
    Id      uint64 `gorm:"column:id;primary_key" json:"id" description:"客户ID"`
    Name    string `gorm:"column:name" json:"name" description:"客户名称" binding:"required" err_msg:"客户名称 必填"`
    Phone   string `gorm:"column:phone" json:"phone" description:"联系电话"`
    Status  string `gorm:"column:status;default:'0'" json:"status" description:"状态;0:禁用,1:正常"`
    BaseModel
}

func (c Customer) TableName() string { return "customer" }

// 分页查询
func (c Customer) SelectCustomerList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]Customer, int64, error) {
    query := db.DBConnections["master"].Model(&Customer{})
    offset, limit := utils.PageParse(pageNum, pageSize)
    for key, value := range conditions {
        query = query.Where(key, value)
    }
    data := make([]Customer, 0)
    var count int64
    query.Count(&count)
    for _, condition := range sortConditions {
        query = query.Order(condition)
    }
    d := query.Offset(offset).Limit(limit).Find(&data)
    return data, count, d.Error
}

func (c Customer) SelectCustomerById(id uint64) (Customer, error) { /* ... */ }
func (c Customer) InsertCustomer(data *Customer) error { /* ... */ }
func (c Customer) UpdateCustomer(data *Customer) error { /* ... */ }
func (c Customer) DeleteCustomerByIds(ids []uint64) error { /* ... */ }
```

### Service 层（`internal/service/customer.go`）

```go
package system

import "go-fin-server/internal/model"

type CustomerService struct{}

func (c CustomerService) SelectCustomerList(conditions map[string]interface{}, sortConditions []string, pageNum, pageSize int) ([]model.Customer, int64, error) {
    return model.Customer{}.SelectCustomerList(conditions, sortConditions, pageNum, pageSize)
}
// ... 其他方法同理委托给 Model
```

**必须在 `service/init.go` 的 `Services` 中注册。**

### API 定义（`api/system/customer.go`）

```go
package system

type CustomerListRequest struct {
    Name     string `json:"name" form:"name"`
    PageNum  int    `json:"pageNum" form:"pageNum"`
    PageSize int    `json:"pageSize" form:"pageSize"`
}

type CustomerIdUriRequest struct {
    Id *uint64 `uri:"id" binding:"required" err_msg:"客户ID 必填"`
}
```

**必须在 `api/init.go` 的 `Api` 中注册。**

### Handler 层（`internal/handler/system/customer.go`）

```go
type CustomerHandler struct {
    Services service.Services
    Api      api.Api
}

func NewCustomerHandler() *CustomerHandler { return &CustomerHandler{} }

// ListCustomer 查询客户列表
//
//	@Summary	查询客户列表
//	@Tags		客户管理
//	@Produce	json
//	@Param		pageNum		query	int		false	"页码"
//	@Param		pageSize	query	int		false	"每页数量"
//	@Success	200			{object}	map[string]interface{}
//	@Router		/system/customer/list [get]
//	@Security	BearerAuth
func (s *CustomerHandler) ListCustomer(c *gin.Context) {
    response.SetOperTitle(c, "查询客户列表")           // 1. 操作日志标题
    req := s.Api.Customer.CustomerListRequest          // 2. 绑定参数
    if err := api.ShouldBindError(c, &req); err != nil {
        response.Error(c, err.Error())
        return
    }
    conditions := utils.BuildConditions(req)
    sortConditions := utils.BuildSortConditions(req.PageNum)
    data, total, err := s.Services.CustomerService.SelectCustomerList(...)  // 3. 调用 Service
    if err != nil {
        pkg.Logger.Error(err)
        response.Error(c, err.Error())
        return
    }
    response.PageData(c, data, total)                  // 4. 返回响应
}
```

### Router 层（`internal/router/system/customer.go`）

```go
func SetupCustomerRoutes(g *gin.RouterGroup) {
    h := system.NewCustomerHandler()
    group := g.Group("/customer")
    {
        group.GET("list", middleware.PermissionMiddleware("system:customer:list"), h.ListCustomer)
        group.GET(":id", middleware.PermissionMiddleware("system:customer:query"), h.GetCustomer)
        group.POST("", middleware.PermissionMiddleware("system:customer:add"), h.AddCustomer)
        group.PUT("", middleware.PermissionMiddleware("system:customer:edit"), h.UpdateCustomer)
        group.DELETE("/:ids", middleware.PermissionMiddleware("system:customer:remove"), h.DelCustomer)
    }
}
```

**必须在 `router.go` 中注册调用 `SetupCustomerRoutes`。**

## 前端各层代码模板

### 类型定义（`types/api/system/customer.ts`）

```ts
export interface Customer {
  id?: number
  name?: string
  phone?: string
  status?: string
  createTime?: string
}

export interface CustomerQueryParams {
  name?: string
  status?: string
  pageNum?: number
  pageSize?: number
}
```

### 接口调用（`views/system/customer/api.ts`）

```ts
import response from '@utils/http'
import { Customer, CustomerQueryParams } from '@/types/api/system/customer'
import { TableDataInfo } from '@/types/api/system/common'

export function listCustomer(query: CustomerQueryParams) {
  return response<TableDataInfo<Customer>>({ url: '/system/customer/list', method: 'get', params: query })
}
export function getCustomer(id?: number) {
  return response<Customer>({ url: '/system/customer/' + (id ?? ''), method: 'get' })
}
export function addCustomer(data: Customer) {
  return response({ url: '/system/customer', method: 'post', data })
}
export function updateCustomer(data: Customer) {
  return response({ url: '/system/customer', method: 'put', data })
}
export function delCustomer(id: number | number[]) {
  return response({ url: '/system/customer/' + id, method: 'delete' })
}
```

### 主页面（`views/system/customer/index.vue`，使用 ArtCrud）

```vue
<template>
  <ArtCrud :config="crudConfig" />
</template>

<script setup lang="ts">
import ArtCrud from '@/components/business/art-crud/index.vue'
import { ref } from 'vue'
import { listCustomer, addCustomer, updateCustomer, delCustomer } from './api'

const crudConfig = ref({
  api: { list: listCustomer, add: addCustomer, update: updateCustomer, delete: delCustomer },
  columns: [
    { type: 'selection' as const },
    { type: 'index' as const, label: '序号', width: 60 },
    { prop: 'name', label: '客户名称', minWidth: 120 },
    { prop: 'phone', label: '联系电话', minWidth: 140 },
    { prop: 'status', label: '状态', minWidth: 100 }
  ],
  search: {
    items: [
      { key: 'name', label: '客户名称', type: 'input', placeholder: '请输入客户名称' }
    ]
  },
  form: {
    items: [
      { key: 'name', label: '客户名称', type: 'input', required: true },
      { key: 'phone', label: '联系电话', type: 'input' }
    ]
  }
})
</script>
```

### 路由注册（`router/modules/system.ts`）

```ts
{
  path: '/system/customer',
  name: 'Customer',
  component: 'layout/index',
  permissions: ['system:customer:list'],
  meta: { title: '客户管理', icon: 'ri:user-line' },
  children: [
    { path: '', component: 'system/customer/index', name: 'CustomerIndex', meta: { title: '客户列表' } }
  ]
}
```

## 布尔字段（checkbox 开关）约定

是/否类字段三层统一为布尔类型，**全程无需转 0/1**：

| 层 | 类型 | 示例 |
| --- | --- | --- |
| 数据库 | `BOOLEAN`（即 `tinyint(1)`） | `menu_check_strictly BOOLEAN DEFAULT 1` |
| 后端 Go | `bool` | `MenuCheckStrictly bool \`gorm:"column:menu_check_strictly"\`` |
| 前端 TS | `boolean` | `menuCheckStrictly?: boolean` |

**禁止**：后端用 `uint` 接布尔值（gin 绑定会报错 `Invalid type for field xxx, expected uint`）。
**禁止**：前端把布尔转成 `0/1` 数字提交（后端 `bool` 无法解析数字 `1`）。

## 开发自检清单

- [ ] 后端：Model / Service（已注册）/ Handler / API（已注册）/ Router（已注册）五层齐全
- [ ] 后端：每个接口有 Swagger 注解与权限中间件
- [ ] 后端：分页用 `response.PageData`，普通用 `response.Data/DataMsg`
- [ ] 前端：类型定义 → api.ts → index.vue → 路由 四步齐全
- [ ] 前端：CRUD 页面优先复用 `ArtCrud`
- [ ] 前后端字段命名、URL、权限标识一致
- [ ] `go build ./...` 与 `pnpm check` 均通过
