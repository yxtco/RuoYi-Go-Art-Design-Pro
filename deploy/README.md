# RuoYi-Go 三容器 Docker 部署

**后端 Go 直接托管前端编译产物**，无需 Nginx，共 3 个容器：

```
┌────────────────────────────  docker 网络 app-net  ────────────────────────────┐
│                                                                               │
│  ┌──────────┐        ┌──────────┐        ┌──────────┐        ┌─────────────┐  │
│  │  用户浏览器 │  ───▶ │  server  │  ───▶  │  mysql   │        │   redis     │  │
│  │ :8080     │  HTTP │ 后端+前端  │  GORM  │  :3306   │  缓存   │   :6379      │  │
│  └──────────┘        │ :8080     │        └──────────┘        └─────────────┘  │
│                      │ API+SPA   │                                              │
│                      │ +WebSocket│                                              │
│                      └──────────┘                                              │
└───────────────────────────────────────────────────────────────────────────────┘
```

## 目录结构

```
deploy/
├── docker-compose.yml        # 3 容器编排：server + mysql + redis
├── backend/
│   └── Dockerfile            # 多阶段：编译前端 → 编译 Go → 运行（后端托管前端）
└── README.md
```

## 工作原理

- **server** 镜像自包含：阶段一用 Node 编译前端（`VITE_API_URL=/` 同源直连），
  阶段二用 Go 编译后端，阶段三把前端 `dist` 拷贝到 `/app/web-dist`，
  由后端 Gin 的 SPA 托管（`internal/router/spa.go`）直接返回页面与 history 回退。
- **无需 Nginx / 无需跨域**：页面、API（`/login`、`/system/**`）、WebSocket（`/ws/chat`）
  全部同源在 `:8080`。
- **mysql**：首次启动自动执行 `server/data/ry_20230706.sql` 建表与种子数据。
- **配置**：后端使用嵌入的 `server/configs/docker.yaml`（`-c docker`），
  数据库/Redis 指向同网络服务名 `mysql` / `redis`。

## 部署步骤

```bash
cd deploy
docker compose up -d --build
```

访问 `http://localhost:8080/`。首次启动需等待 mysql 初始化（约 20~60 秒），
`docker compose ps` 显示 server 为 healthy 后即可登录。

### 常用命令

```bash
docker compose ps                     # 查看状态
docker compose logs -f server         # 后端日志
docker compose logs -f mysql          # 数据库初始化日志
docker compose down                   # 停止（保留数据卷）
docker compose down -v                # 停止并删除数据卷（会清空数据库）
```

## 关键配置对照

| 项 | 位置 | 说明 |
| --- | --- | --- |
| MySQL 密码 | compose `MYSQL_ROOT_PASSWORD` 与 `server/configs/docker.yaml` | 需一致 |
| MySQL 库名 | `MYSQL_DATABASE: ry-test` 与 docker.yaml `database: ry-test` | 需一致 |
| 数据库初始化 SQL | compose 挂载 `../server/data/ry_20230706.sql` | 仅首次建库时执行 |
| 前端产物目录 | docker.yaml `server.web_dist: /app/web-dist` | 后端托管路径 |
| 上传存储 | compose 卷 `uploads-data` 挂到 `/app/uploads` | 持久化上传文件 |
| 对外端口 | compose `8080:8080` | 前端+后端入口 |

## 前端加载优化

- **代码拆分**：`web/vite.config.ts` 的 `build.rollupOptions.output.manualChunks`
  把 Vue 生态、echarts、xlsx、wangeditor、xgplayer、axios 等大库拆成独立 chunk，
  减小单个 JS 体积、便于并行加载与浏览器缓存（element-plus 保持按需导入）。
- **构建期压缩**：编译同时生成 gzip(`.gz`) 与 brotli(`.br`) 压缩文件。
- **后端按编码返回**：`server/internal/router/spa.go` 的 `serveStaticCompressed`
  根据浏览器 `Accept-Encoding` 优先返回 `.br`（再 `.gz`），显著减小传输体积。

## 验证

1. `curl http://localhost:8080/` 返回前端 `index.html`。
2. `curl -X POST http://localhost:8080/captchaImage` 返回验证码 JSON（后端连通）。
3. 浏览器打开 `http://localhost:8080/`，登录后进入聊天，显示「已连接」并可看到在线用户。

## 资源限制

三个容器均设置 CPU / 内存上下限（limits + reservations）。
如需带宽限速（如 2mb），Docker Compose 无直接字段，可用宿主机 `tc` 对容器 veth 接口限速。
