# 数据库变更必须添加迁移脚本

## 规则

当修改涉及数据库结构或数据变更时，**必须**同时添加对应的数据库迁移脚本。

## 触发场景

以下情况必须添加迁移：
- 新增表
- 新增/修改/删除字段
- 新增/修改/删除索引
- 插入必要的系统配置数据（如 `sys_config` 新增键值）
- 数据修复或数据迁移

以下情况不需要迁移：
- 仅修改 Go 代码逻辑，不涉及数据库变更
- 仅修改前端代码

## 迁移文件规范

项目使用 `golang-migrate`，迁移文件位于 `server/migrations/` 目录，通过 `embed.go` 内嵌到二进制中，启动时自动执行。

### 文件命名

```
NNNNNN_描述.up.sql      -- 正向迁移（必须）
NNNNNN_描述.down.sql    -- 回滚迁移（必须）
```

- `NNNNNN` 为 6 位数字序号，在已有最大序号基础上 +1
- 描述使用小写英文 + 下划线，简明概括变更内容
- 例如：`000003_add_user_avatar.up.sql`

### up.sql 编写要求

- 使用 `IF NOT EXISTS` / `INSERT IGNORE` 等幂等语句，确保重复执行不报错
- 每条语句添加中文注释说明用途
- 示例：

```sql
-- 新增用户头像字段
ALTER TABLE sys_user ADD COLUMN IF NOT EXISTS avatar varchar(255) DEFAULT '' COMMENT '头像地址';

-- 新增系统配置（幂等插入）
INSERT IGNORE INTO sys_config (config_name, config_key, config_value, config_type, create_by, create_time, remark)
VALUES ('配置名称', 'config.key', '默认值', 'N', 'admin', NOW(), '配置说明');
```

### down.sql 编写要求

- 提供回滚操作，使迁移可逆
- 示例：

```sql
-- 回滚：删除用户头像字段
ALTER TABLE sys_user DROP COLUMN IF EXISTS avatar;

-- 回滚：删除系统配置
DELETE FROM sys_config WHERE config_key IN ('config.key');
```

## 检查清单

提交涉及数据库变更的代码前，确认：

1. ✅ 已在 `server/migrations/` 创建对应的 `.up.sql` 和 `.down.sql`
2. ✅ 序号正确（在已有最大序号上 +1）
3. ✅ 使用了幂等语句（`IF NOT EXISTS` / `INSERT IGNORE`）
4. ✅ 同步更新了 `server/data/ry_init_full.sql`（全新数据库初始化脚本）
5. ✅ 后端编译通过（`go build ./...`）
