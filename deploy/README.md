# RuoYi-Go 前后端容器（开发式）Docker 部署

**只创建 backend + web 两个容器**，MySQL / Redis 由宿主机提供（不在此编排内）。

- 后端：`golang` 官方镜像，容器内直接 `go run main.go -c deploy`，**不预编译二进制**
- 前端：`node` 官方镜像，容器内直接 `vite --mode docker` 跑开发服务，**不预构建**
- 源码通过绑定挂载实时同步进容器；更新代码只需 `git pull` + `docker restart`，**无需重建镜像**

```
┌────────────────────────  docker 网络 ruoyi-app-net  ────────────────────────┐
│                                                                             │
│  浏览器 ─▶ 宿主 80 ─▶ [web 容器 vite dev :80]                                 │
│                 │  /dev-api 代理                                               │
│                 └──────────▶ [backend 容器 go run :8080]                      │
│                                   │   │                                       │
│              host.docker.internal│   │（extra_hosts: host-gateway）            │
│                                   ▼   ▼                                       │
│                   宿主机 MySQL(:3306)   宿主机 Redis(:6380)                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

## 目录结构

```
deploy/
├── docker-compose.yml        # 2 容器编排：backend + web（无 mysql/redis）
└── README.md
server/configs/deploy.yaml    # 后端连接宿主机 MySQL/Redis 的配置（-c deploy 读取）
web/.env.docker               # 前端 docker 模式环境变量（vite --mode docker 加载）
```

> 注：`deploy/backend/Dockerfile` 为旧的“单镜像托管前端”方案所用，本编排不再引用，保留备查。

## 前提条件（部署服务器）

1. 已安装 **Docker** 与 **Docker Compose v2 插件**（`docker compose version` 可查）。
2. 宿主机已运行 **MySQL** 与 **Redis**（本编排不创建这两个容器）。
3. 已在服务器克隆本项目代码（`git clone <repo>`），后续更新用 `git pull`。

## 关键配置

### 后端连接：`server/configs/deploy.yaml`

后端容器通过 `host.docker.internal` 访问宿主机 MySQL/Redis（compose 已配
`extra_hosts: host.docker.internal:host-gateway`，Linux 下可用）。

| 项 | 默认值 | 说明 |
| --- | --- | --- |
| MySQL 地址 | `host.docker.internal:3306` | 宿主机 MySQL |
| MySQL 账号/密码 | `root` / `Jerry322` | **请改为宿主机实际密码** |
| MySQL 库名 | `ry-test` | 首次为空库时后端自动导入 `server/data/ry_init_full.sql` |
| Redis 地址 | `host.docker.internal:6380` | 宿主机 Redis |
| Redis 密码 / db | `Jerry322` / `5` | **请改为宿主机实际密码/db** |

改配置后重启后端即可：`docker compose -f deploy/docker-compose.yml restart backend`

### 前端代理：`web/.env.docker`

容器内 `vite --mode docker` 加载，API `/dev-api` 由 vite 代理转发到后端容器
`http://backend:8080`（同网络服务名）。改后重启 web 生效。

## 部署步骤

```bash
# 在仓库根目录执行
docker compose -f deploy/docker-compose.yml up -d

# 查看状态
docker compose -f deploy/docker-compose.yml ps

# 查看日志（首次启动会下载依赖，耐心等待）
docker compose -f deploy/docker-compose.yml logs -f backend
docker compose -f deploy/docker-compose.yml logs -f web
```

访问：

- 前端：`http://<服务器IP>/`
- 后端 API / Swagger：`http://<服务器IP>:8080/`

> 若宿主机 80 端口被占用，把 compose 里 `web` 的 `"80:80"` 改为如 `"8081:80"`。

## 更新代码（git pull + restart，不重建镜像）

源码已绑定挂载，容器内直接读宿主机工作区，因此拉取最新代码后**重启容器**即可生效：

```bash
cd /path/to/repo
git pull                        # 拉取最新代码
docker compose -f deploy/docker-compose.yml restart backend web
```

说明：
- 后端依赖缓存在命名卷 `go-mod` / `go-build`，重启不重新下载/编译（增量编译很快）。
- 前端依赖缓存在命名卷 `web-node-modules`，重启不重装。
- 仅当 `package.json` / `go.mod` 发生依赖变更时才需要手动重装：
  ```bash
  docker compose -f deploy/docker-compose.yml exec web pnpm install
  docker compose -f deploy/docker-compose.yml exec backend go mod tidy
  ```

## 数据持久化方案

| 数据 | 存放位置 | 持久化方式 |
| --- | --- | --- |
| MySQL 数据 | 宿主机 MySQL 数据目录 | 由宿主机自身持久（非容器内） |
| Redis 数据 | 宿主机 Redis 数据目录 | 由宿主机自身持久（非容器内） |
| 上传文件 | 宿主机 `server/uploads`（容器 `/app/uploads`） | 源码绑定挂载，随宿主机持久 |
| 后端日志 | 宿主机 `server/logs`（容器 `/app/logs`） | 源码绑定挂载，随宿主机持久 |
| Go 模块缓存 | 命名卷 `go-mod` | Docker 命名卷，容器重建保留 |
| Go 编译缓存 | 命名卷 `go-build` | Docker 命名卷，容器重建保留 |
| 前端 node_modules | 命名卷 `web-node-modules` | Docker 命名卷，容器重建保留 |

> 命名卷在 `docker compose down`（不删卷）后保留；`docker compose down -v` 会清空命名卷。

## 验证

### ① 修改代码 → 重启 → 新代码生效

```bash
# 后端：改一处接口返回，例如 server/main.go 的启动文案
vim server/main.go
docker compose -f deploy/docker-compose.yml restart backend
docker compose -f deploy/docker-compose.yml logs -f backend   # 看到新文案即生效

# 前端：改页面标题
vim web/index.html
docker compose -f deploy/docker-compose.yml restart web
curl -s http://<服务器IP>/ | grep 新标题                          # 命中即生效
```

### ② 容器删除重建 → 持久化数据完整保留

```bash
# 先造数据：登录后上传一个文件到 server/uploads，并往 MySQL 写一条记录

docker compose -f deploy/docker-compose.yml down       # 停止并删除容器（保留卷）
docker compose -f deploy/docker-compose.yml up -d       # 重新创建容器

# 校验
ls server/uploads/                                      # 上传文件仍在
# 登录系统查看之前写入的数据仍在；Redis 缓存由后端启动时从 MySQL 重载
```

### ③ git pull 更新后服务可正常运行

```bash
cd /path/to/repo && git pull
docker compose -f deploy/docker-compose.yml restart backend web
docker compose -f deploy/docker-compose.yml ps         # 两容器均 Up
curl -s -X POST http://<服务器IP>:8080/captchaImage     # 返回验证码 JSON = 后端连通
curl -s http://<服务器IP>/ | head                       # 返回页面 = 前端连通
```

## 常用命令

```bash
docker compose -f deploy/docker-compose.yml up -d        # 启动
docker compose -f deploy/docker-compose.yml restart backend web   # 重启（更新后）
docker compose -f deploy/docker-compose.yml logs -f web  # 前端日志
docker compose -f deploy/docker-compose.yml ps           # 状态
docker compose -f deploy/docker-compose.yml down         # 停止（保留卷）
docker compose -f deploy/docker-compose.yml down -v      # 停止并清空命名卷
```

## 常见问题

- **后端启动报连接 MySQL 失败**：核对 `server/configs/deploy.yaml` 中宿主机 MySQL 的
  地址（`host.docker.internal`）、端口、密码、库名是否正确，以及 MySQL 是否允许远程/宿主机访问。
- **容器内 `host.docker.internal` 不通**：确认宿主机 docker 版本支持
  `extra_hosts: host.docker.internal:host-gateway`（Compose 2.20+ / Docker 20.10+）。
- **前端页面 API 404**：确认 vite 代理正常（`web/.env.docker` 的
  `VITE_API_PROXY_URL` 指向 `http://backend:8080`），且两容器在同一网络。
- **首次启动慢**：首次需下载 Go 模块与 node_modules，属正常；后续重启秒级。
- **宿主机 mysql/redis 需监听 0.0.0.0 或 docker0**：否则容器经 host-gateway 无法回连。
