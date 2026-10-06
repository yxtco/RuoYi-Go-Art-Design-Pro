# 后端开发规范（Go 版若依）

## 技术栈

| 项目 | 说明 |
| --- | --- |
| 模块名 | `go-fin-server` |
| Web 框架 | `gin` |
| ORM | `gorm`（支持 `mysql` / `postgres`） |
| 配置 | `spf13/viper`，配置文件 `configs/release.yaml` |
| 日志 | `zap` + `lumberjack`（禁止使用 `fmt.Println`，统一使用 `pkg.Logger`） |
| 缓存 | `go-redis` v9（多实例，默认 `master`） |
| 鉴权 | `golang-jwt/jwt/v5`，Token 存 Redis |
| 接口文档 | `swaggo/swag`，注解写在 Handler 上 |

## 分层架构（单向依赖，禁止跨层）

```
router → middleware → handler → service → model → 数据库
```

各层职责：
- **Model**（`internal/model/`）：GORM 结构体 + 数据访问方法，文件名与表名对应
- **Service**（`internal/service/`）：业务编排，须在 `service/init.go` 的 `Services` 中注册
- **Handler**（`internal/handler/`）：绑定参数 → 调用 Service → 返回响应，禁止直接操作数据库
- **Router**（`internal/router/`）：注册路由 + 挂载权限中间件
- **API**（`api/`）：请求/响应结构体定义，须在 `api/init.go` 的 `Api` 中注册

## Model 层规范

- 字段 Tag 必须包含：`gorm:"column:xxx"` + `json:"xxx"`（驼峰）+ `description:"xxx"`
- 校验用 `binding:"required"` + `err_msg:"自定义消息"`
- 时间字段使用 `pkg/types.LocalTime`
- 每个模型必须实现 `TableName() string`
- 分页查询签名：`SelectXxxList(conditions, sortConditions, pageNum, pageSize) ([]Model, int64, error)`
- 条件由上层传入 `map[string]interface{}`，配合 `query.Where(key, value)` 动态拼接
- 分页使用 `pkg/utils.PageParse(pageNum, pageSize)`
- **布尔字段**：是/否类字段统一用 `bool` 类型（对应 MySQL `BOOLEAN`），禁止用 `uint`

## Handler 层规范

每个方法必须遵循以下范式：
1. `response.SetOperTitle(c, "操作名称")` — 记录操作日志标题
2. 使用 `api.ShouldBindError(c, &req)` 或 `api.ShouldBindUriError(c, &req)` 绑定参数
3. 调用 Service 层方法
4. 成功用 `response.PageData` / `response.Data` / `response.DataMsg`，失败用 `response.Error`
5. 方法上方必须编写 Swagger 注解（`@Summary` / `@Tags` / `@Router` / `@Security BearerAuth`）

## Router 层规范

- 每个模块一个 `SetupXxxRoutes(g *gin.RouterGroup)` 函数，在 `router.go` 中注册
- 必须挂载权限中间件：`middleware.PermissionMiddleware("system:xxx:action")`
- HTTP 方法约定：查询 GET、新增 POST、修改 PUT、删除 DELETE（批量 `DELETE /:ids`）

## 统一响应规范

| 方法 | 适用场景 |
| --- | --- |
| `response.Data(c, data)` | 普通查询/操作 |
| `response.DataMsg(c, data, msg)` | 带自定义消息 |
| `response.PageData(c, data, total)` | **分页列表**（前端 `TableDataInfo`） |
| `response.Error(c, msg)` | 业务错误（code=400） |

业务成功统一 `code=200`，失败 `code=400`，HTTP 状态码始终 200。

## 鉴权

- JWT 中间件注入 `loginUser`，Handler 通过 `c.MustGet("loginUser").(*model.LoginUser)` 获取
- 权限标识与前端菜单权限一致（如 `system:user:list`）
- 数据权限在 Service/Model 层处理

## 新增业务模块 Checklist

1. **Model**：`internal/model/xxx.go`，定义结构体 + `TableName()` + 查询方法
2. **Service**：`internal/service/xxx.go`，在 `service/init.go` 注册
3. **API**：`api/xxx.go`，定义请求结构体，在 `api/init.go` 注册
4. **Handler**：`internal/handler/xxx.go`，含 Swagger 注解 + 权限中间件
5. **Router**：`internal/router/xxx.go`，在 `router.go` 注册
6. **数据库迁移**：涉及表结构/数据变更时，必须在 `server/migrations/` 添加迁移脚本
7. **验证**：`go build ./...` 编译通过
