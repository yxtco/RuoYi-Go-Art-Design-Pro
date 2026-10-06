# RuoYi-Go Docker 部署

提供三种部署方式，按需选择：

| 方式 | 适用场景 | 说明 |
|------|---------|------|
| **方式一：生产部署** | 正式上线 | 单容器，前后端预编译产物打包，镜像轻量 |
| **方式二：开发部署** | 本地开发/调试 | 双容器，源码挂载，热更新，无需重建镜像 |
| **方式三：远程部署** | 远程服务器部署 | 本地编译 + SSH 传输 + 远程 Docker 构建运行 |

MySQL / Redis 均由宿主机（或独立容器）提供，不在编排内创建。

---

## 方式一：生产部署（Dockerfile）

前后端均在宿主机预编译，Docker 镜像仅包含运行环境（Alpine）+ 产物，无需容器内编译。

```
┌──────────────────────────────────────────────┐
│  单容器（Alpine）                              │
│                                               │
│  浏览器 ─▶ 宿主 8080 ─▶ [server :8080]         │
│                            │  /      → 前端    │
│                            │  /api   → 后端    │
│                            │  /ws    → WS     │
│                            ▼                   │
│              宿主机 MySQL(:3306) / Redis(:6380) │
└──────────────────────────────────────────────┘
```

### 前置步骤

```bash
# 1. 编译后端二进制（在 server/ 目录下）
cd server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../deploy/backend/server main.go

# 2. 编译前端产物（在 web/ 目录下，输出到 server/web-dist）
cd ../web
pnpm build
```

### 构建镜像

```bash
# 在仓库根目录执行
docker build -f deploy/backend/Dockerfile -t ruoyi-go .
```

### 运行容器

```bash
docker run -d \
  --name ruoyi-go \
  --restart unless-stopped \
  -p 8080:8080 \
  --add-host host.docker.internal:host-gateway \
  -v ruoyi-uploads:/app/uploads \
  -v ruoyi-logs:/app/logs \
  ruoyi-go
```

### 关键配置

容器内读取嵌入二进制的 `docker.yaml`（`server/configs/docker.yaml`），按需修改：

| 项 | 默认值 | 说明 |
|---|--------|------|
| MySQL 地址 | `mysql:3306` | 如有独立 MySQL 容器，改服务名；宿主机则改 `host.docker.internal` |
| MySQL 密码 | `ruoyi123` | **改为实际密码** |
| Redis 地址 | `redis:6379` | 同上 |
| Redis 密码 | 空 | **如有密码请填写** |

> 修改配置后需重新编译二进制（配置嵌入到二进制中）。

### 访问

- 前端 + 后端 API：`http://<服务器IP>:8080/`（同源，后端托管前端静态文件）
- Swagger 文档：`http://<服务器IP>:8080/swagger-ui/index.html`

### 更新流程

```bash
# 重新编译 → 重建镜像 → 替换容器
cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../deploy/backend/server main.go
cd ../web && pnpm build
cd .. && docker build -f deploy/backend/Dockerfile -t ruoyi-go .
docker stop ruoyi-go && docker rm ruoyi-go
docker run -d --name ruoyi-go --restart unless-stopped -p 8080:8080 \
  -v ruoyi-uploads:/app/uploads -v ruoyi-logs:/app/logs ruoyi-go
```

---

## 方式二：开发部署（docker-compose.yml）

后端 `go run` + 前端 `vite dev`，源码绑定挂载实时同步，改码重启即生效，无需重建镜像。

```
┌──────────────── docker 网络 ruoyi-app-net ────────────────────┐
│                                                                │
│  浏览器 ─▶ 宿主 80 ─▶ [web 容器 vite dev :80]                   │
│                 │  /dev-api 代理                                │
│                 └──────▶ [backend 容器 go run :8080]            │
│                              │                                 │
│                 host.docker.internal                            │
│                              ▼                                 │
│              宿主机 MySQL(:3306) / Redis(:6380)                 │
└────────────────────────────────────────────────────────────────┘
```

### 前提条件

1. 已安装 **Docker** 与 **Docker Compose v2**
2. 宿主机已运行 **MySQL** 与 **Redis**
3. 已克隆本项目代码

### 关键配置

**后端**：`server/configs/deploy.yaml`

| 项 | 默认值 | 说明 |
|---|--------|------|
| MySQL 地址 | `host.docker.internal:3306` | 宿主机 MySQL |
| MySQL 密码 | `Jerry322` | **改为宿主机实际密码** |
| Redis 地址 | `host.docker.internal:6380` | 宿主机 Redis |
| Redis 密码 | `Jerry322` | **改为宿主机实际密码** |

**前端**：`web/.env.docker`

API 代理由 vite 转发到 `http://backend:8080`（同网络服务名）。

### 启动

```bash
docker compose -f deploy/docker-compose.yml up -d

# 查看状态 / 日志
docker compose -f deploy/docker-compose.yml ps
docker compose -f deploy/docker-compose.yml logs -f backend
docker compose -f deploy/docker-compose.yml logs -f web
```

### 访问

- 前端：`http://<服务器IP>/`（web 容器 vite dev，端口 80）
- 后端 API：`http://<服务器IP>:8080/`

### 更新代码

```bash
git pull
docker compose -f deploy/docker-compose.yml restart backend web
```

> 源码已绑定挂载，容器内直接读宿主机工作区，重启即可生效。

---

## 方式三：远程部署（本地编译 + SSH + Docker）

本地编译后端二进制和前端静态文件，通过 SSH 传输到远程服务器，在服务器上使用 Docker 构建和运行。

**与方式一的区别**：方式一适合在服务器本地操作；方式三适合从开发机直接部署到远程服务器，无需在服务器上安装 Go/Node.js。

### 架构示意

```
本地开发机（macOS/Linux/Windows）        远程服务器（Linux）
┌──────────────────────────┐            ┌────────────────────────────────────┐
│ ① cd server && go build  │   SCP      │ ④ 接收 server 二进制 + web-dist    │
│ ② cd web && pnpm build   │ ────────▶  │ ⑤ docker build → ruoyi-go 镜像     │
│ ③ scp 产物 + Dockerfile  │            │ ⑥ docker run → ruoyi-go 容器       │
└──────────────────────────┘            │                                    │
                                        │  ┌──────────────────────────────┐  │
                                        │  │  单容器 Alpine（~164MB）      │  │
                                        │  │  :8080 /      → 前端静态      │  │
                                        │  │  :8080 /api   → 后端 API     │  │
                                        │  │  :8080 /ws    → WebSocket    │  │
                                        │  └──────────────────────────────┘  │
                                        │         │ host.docker.internal     │
                                        │         ▼                          │
                                        │  MySQL(:3306)  Redis(:6380)        │
                                        └────────────────────────────────────┘
```

### 前提条件

| 环境 | 要求 |
|------|------|
| **本地开发机** | Go（编译后端）、Node.js + pnpm（编译前端） |
| **远程服务器** | Docker >= 20.10（需要 `host.docker.internal` 支持） |
| **网络** | 本地可 SSH 连接远程服务器 |
| **数据库** | 远程服务器已运行 MySQL 和 Redis，且监听 `0.0.0.0` |

### 步骤一：修改配置文件

编辑 `server/configs/docker.yaml`，将数据库和 Redis 连接改为远程服务器实际值：

```yaml
# server/configs/docker.yaml

server:
  port: 8080
  host: 0.0.0.0                    # 必须为 0.0.0.0，容器内需接受外部连接
  web_dist: '/app/web-dist'        # 前端产物目录（容器内路径）

database_list:
  - name: master
    driver: mysql
    host: host.docker.internal     # 指向宿主机
    port: 3306
    username: root
    password: 'your-mysql-password'  # ← 改为实际密码
    database: ry-go                  # ← 改为实际数据库名
    logLevel: 1

redis_list:
  - name: master
    addr: 'host.docker.internal:6380'  # ← 改为实际地址和端口
    password: 'your-redis-password'    # ← 改为实际密码
    db: 5                              # ← 改为实际 DB 编号
```

> **重要**：配置通过 `//go:embed` 嵌入到二进制中，修改后必须重新编译。

### 步骤二：本地编译

```bash
# 进入项目根目录
cd /path/to/ruoyi-go

# 1. 编译后端二进制（交叉编译 Linux amd64）
cd server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../deploy/backend/server main.go
cd ..

# 2. 编译前端产物（输出到 server/web-dist）
cd web
pnpm build
cd ..

# 确认产物
ls -lh deploy/backend/server     # 后端二进制（约 70MB）
ls -lh server/web-dist/          # 前端静态文件（约 10MB）
```

### 步骤三：传输产物到服务器

```bash
# ===== 变量定义（替换为实际值）=====
SERVER_USER="Jerry"
SERVER_HOST="192.168.3.123"
SERVER_PATH="/vol1/1000/ruoyi-go"

# 清理远程旧的前端产物并创建目录
ssh ${SERVER_USER}@${SERVER_HOST} \
  "rm -rf ${SERVER_PATH}/server/web-dist && mkdir -p ${SERVER_PATH}/server/web-dist ${SERVER_PATH}/deploy/backend"

# 传输后端二进制
scp deploy/backend/server ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/deploy/backend/

# 传输前端产物
scp -r server/web-dist/* ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/server/web-dist/

# 传输 Dockerfile
scp deploy/backend/Dockerfile ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/deploy/backend/
```

### 步骤四：SSH 到服务器构建镜像并运行

```bash
ssh ${SERVER_USER}@${SERVER_HOST}

# ---- 以下在服务器上执行 ----

# 构建镜像
cd ${SERVER_PATH}
docker build -f deploy/backend/Dockerfile -t ruoyi-go .

# 停止并删除旧容器（不存在也不报错）
docker stop ruoyi-go 2>/dev/null && docker rm ruoyi-go 2>/dev/null || true

# 启动新容器
docker run -d \
  --name ruoyi-go \
  --restart unless-stopped \
  -p 8080:8080 \
  --add-host host.docker.internal:host-gateway \
  -v ruoyi-uploads:/app/uploads \
  -v ruoyi-logs:/app/logs \
  ruoyi-go

# 验证启动
docker ps | grep ruoyi-go
docker logs ruoyi-go
```

启动成功的日志应包含：

```
上传目录初始化完成: /app/uploads
数据库迁移完成
字典缓存加载完成
定时任务调度器已启动
WebSocket Hub 已启动
前端静态托管已启用: distDir=/app/web-dist
========================================
  服务启动成功!
========================================
```

### 步骤五：验证部署

```bash
# 在服务器上验证
curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/        # 前端页面，应返回 200
curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/api/    # 后端接口，应返回 200
```

浏览器访问：
- 前端页面：`http://<服务器IP>:8080/`
- Swagger 文档：`http://<服务器IP>:8080/swagger-ui/index.html`

### 一键部署脚本

在项目根目录创建 `deploy-remote.sh`：

```bash
#!/bin/bash
set -e

# ========== 配置区（按需修改）==========
SERVER_USER="Jerry"
SERVER_HOST="192.168.3.123"
SERVER_PATH="/vol1/1000/ruoyi-go"
# =======================================

echo "🔨 编译后端二进制..."
cd server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../deploy/backend/server main.go
cd ..

echo "🎨 编译前端产物..."
cd web
pnpm build
cd ..

echo "📤 传输产物到服务器..."
ssh ${SERVER_USER}@${SERVER_HOST} \
  "rm -rf ${SERVER_PATH}/server/web-dist && mkdir -p ${SERVER_PATH}/server/web-dist ${SERVER_PATH}/deploy/backend"
scp deploy/backend/server ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/deploy/backend/
scp -r server/web-dist/* ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/server/web-dist/
scp deploy/backend/Dockerfile ${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/deploy/backend/

echo "🐳 远程构建和部署..."
ssh ${SERVER_USER}@${SERVER_HOST} << 'ENDSSH'
  cd /vol1/1000/ruoyi-go
  docker build -f deploy/backend/Dockerfile -t ruoyi-go .
  docker stop ruoyi-go 2>/dev/null && docker rm ruoyi-go 2>/dev/null || true
  docker run -d \
    --name ruoyi-go \
    --restart unless-stopped \
    -p 8080:8080 \
    --add-host host.docker.internal:host-gateway \
    -v ruoyi-uploads:/app/uploads \
    -v ruoyi-logs:/app/logs \
    ruoyi-go
  echo "--- 容器状态 ---"
  docker ps | grep ruoyi-go
  echo "--- 启动日志 ---"
  docker logs --tail 10 ruoyi-go
ENDSSH

echo ""
echo "🎉 部署完成！"
echo "   访问地址: http://${SERVER_HOST}:8080/"
echo "   Swagger:  http://${SERVER_HOST}:8080/swagger-ui/index.html"
```

使用方式：

```bash
chmod +x deploy-remote.sh
./deploy-remote.sh
```

### 更新流程

```bash
# 方式一：使用脚本（推荐）
./deploy-remote.sh

# 方式二：手动更新
cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../deploy/backend/server main.go
cd ../web && pnpm build
cd ..
scp deploy/backend/server Jerry@server:/vol1/1000/ruoyi-go/deploy/backend/
scp -r server/web-dist/* Jerry@server:/vol1/1000/ruoyi-go/server/web-dist/
ssh Jerry@server "cd /vol1/1000/ruoyi-go && docker build -f deploy/backend/Dockerfile -t ruoyi-go . && docker stop ruoyi-go && docker rm ruoyi-go && docker run -d --name ruoyi-go --restart unless-stopped -p 8080:8080 --add-host host.docker.internal:host-gateway -v ruoyi-uploads:/app/uploads -v ruoyi-logs:/app/logs ruoyi-go"
```

### 服务器 Docker 清理

部署完成后，可清理不再使用的旧镜像释放磁盘空间：

```bash
# 在服务器上执行
# 删除所有悬空镜像（未被任何容器引用的镜像）
docker image prune -f

# 删除所有未被运行中容器引用的镜像（更彻底）
docker image prune -a -f

# 清理构建缓存
docker builder prune -f

# 查看清理后的镜像
docker images
```

---

## 目录结构

```
deploy/
├── backend/
│   └── Dockerfile           # 方式一/三：生产部署（单容器，预编译产物）
├── docker-compose.yml       # 方式二：开发部署（双容器，go run + vite dev）
└── README.md                # 本文档

server/configs/
├── docker.yaml              # 方式一/三配置（嵌入二进制，-c docker）
└── deploy.yaml              # 方式二配置（-c deploy，连接宿主机 MySQL/Redis）

web/
└── .env.docker              # 方式二前端环境变量（vite --mode docker）
```

## 数据持久化

| 数据 | 方式一（生产） | 方式二（开发） | 方式三（远程） |
|------|--------------|--------------|---------------|
| MySQL / Redis | 宿主机自身持久 | 宿主机自身持久 | 宿主机自身持久 |
| 上传文件 | Docker 命名卷 `ruoyi-uploads` | 源码绑定挂载 `server/uploads` | Docker 命名卷 `ruoyi-uploads` |
| 后端日志 | Docker 命名卷 `ruoyi-logs` | 源码绑定挂载 `server/logs` | Docker 命名卷 `ruoyi-logs` |

## 常见问题

- **后端启动报 MySQL 连接失败**：核对配置文件中地址、端口、密码是否正确，以及 MySQL 是否允许远程访问。
- **容器内 `host.docker.internal` 不通**：确认 Docker 版本 >= 20.10，Compose 已配 `extra_hosts: host.docker.internal:host-gateway`。
- **宿主机 MySQL/Redis 需监听 `0.0.0.0`**：否则容器经 host-gateway 无法回连。
- **首次启动慢（方式二）**：首次需下载 Go 模块与 node_modules，后续重启秒级。
