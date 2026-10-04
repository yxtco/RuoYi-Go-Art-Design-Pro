<h1 align="center">RuoYi-Go Art Design Pro</h1>

<p align="center">
  <strong>基于 <a href="https://github.com/nils-golothm/ruoyi-go">ruoyi-go</a> × <a href="https://github.com/Daymychen/art-design-pro">Art Design Pro</a> 的全栈中后台管理系统</strong>
</p>

<p align="center">
  <a href="https://github.com/chen496817468/ruoyi-go-art-design-pro" target="_blank">GitHub</a> ·
  <a href="#快速开始">快速开始</a> ·
  <a href="#功能模块">功能模块</a> ·
  <a href="#致谢">致谢</a>
</p>

<p align="center">
  <a href="https://github.com/chen496817468/ruoyi-go-art-design-pro">
    <img src="https://img.shields.io/badge/GitHub-ruoyi--go--art--design--pro-181717?logo=github" alt="GitHub" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/Gin-1.12-00ADD8" alt="Gin" />
  <img src="https://img.shields.io/badge/GORM-1.31-4A90D9" alt="GORM" />
  <img src="https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vue.js" alt="Vue" />
  <img src="https://img.shields.io/badge/TypeScript-6.0-3178C6?logo=typescript" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Element_Plus-2.11-409EFF?logo=element" alt="Element Plus" />
  <img src="https://img.shields.io/badge/Tailwind_CSS-4-06B6D4?logo=tailwindcss" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/Vite-7-646CFF?logo=vite" alt="Vite" />
  <img src="https://img.shields.io/badge/Redis-9.x-DC382D?logo=redis" alt="Redis" />
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql" alt="MySQL" />
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker" alt="Docker" />
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License" />
</p>

---

## 简介

**RuoYi-Go Art Design Pro** 是基于 [ruoyi-go](https://github.com/nils-golothm/ruoyi-go)（by nils-golothm）后端核心，前端采用 [Art Design Pro](https://github.com/Daymychen/art-design-pro) 设计体系进行全面重写的全栈中后台管理系统。

- **后端**：沿用 [ruoyi-go](https://github.com/nils-golothm/ruoyi-go) 的 Gin + GORM + Redis 架构，完整保留若依核心功能模块
- **前端**：基于 [Art Design Pro](https://github.com/Daymychen/art-design-pro) 设计体系，采用 Vue 3 + TypeScript + Element Plus + Tailwind CSS 4 全面重写，对接若依后端 RBAC 权限体系

项目融合了 ruoyi-go 成熟的后端能力与 Art Design Pro 现代化的前端 UI 设计，提供开箱即用的企业级开发体验。

---

## 特性

### 后端（基于 ruoyi-go）

- 🚀 **Gin + GORM** — 高性能 Web 框架与 ORM
- 🔐 **JWT + Redis 鉴权** — Bearer Token 认证，支持单点登录踢人
- 🛡️ **RBAC 权限体系** — 用户-角色-菜单三级权限控制
- 📝 **Zap 日志系统** — 三级日志分离（系统/操作/错误），支持文件轮转
- 📊 **SQL 监控** — 慢 SQL 统计与分析
- 🔔 **WebSocket 实时通信** — 在线用户管理与消息推送
- ⏰ **定时任务调度** — 基于 cron 的任务管理与执行日志
- 🔑 **RSA 加解密** — 登录密码传输安全加密
- 🧩 **字典缓存** — MySQL 字典数据自动加载至 Redis
- 📦 **数据库迁移** — golang-migrate 内嵌迁移，幂等安全
- 🐳 **Docker 一键部署** — 三容器编排（Server + MySQL + Redis）
- 📄 **Swagger API 文档** — 自动生成接口文档

### 前端（Art Design Pro 设计体系）

- 🎨 **Art Design Pro 设计体系** — 现代化 UI 风格，区别于传统若依前端
- 🧩 **Vue 3 + TypeScript** — 完整类型推导与 IDE 支持
- ⚡ **Vite 7** — 极速开发体验，毫秒级 HMR
- 🎭 **Element Plus** — 企业级 UI 组件库
- 🎨 **Tailwind CSS 4** — 原子化 CSS，快速构建自定义样式
- 🌓 **深色模式** — 完善的明暗主题切换
- 🌍 **国际化** — 基于 vue-i18n 的多语言支持
- 📦 **Pinia** — 类型安全的状态管理
- 📊 **ECharts 6** — 强大的数据可视化能力
- 🔧 **Oxlint + Oxfmt** — Rust 高性能代码检查与格式化
- 🔒 **JSEncrypt** — 前端 RSA 加密传输

---

## 功能模块

| 模块 | 说明 | 状态 |
| :--- | :--- | :---: |
| 用户管理 | 用户 CRUD、个人中心、头像上传 | ✅ |
| 角色管理 | 角色 CRUD、菜单权限分配、数据权限 | ✅ |
| 菜单管理 | 三级菜单配置、按钮权限标识 | ✅ |
| 部门管理 | 树形部门结构 | ✅ |
| 岗位管理 | 岗位信息维护 | ✅ |
| 字典管理 | 字典类型 + 字典数据，Redis 缓存 | ✅ |
| 参数设置 | 系统参数配置 | ✅ |
| 通知公告 | 公告发布、未读数量统计 | ✅ |
| 操作日志 | 操作审计 + 本地文件记录 | ✅ |
| 登录日志 | 登录历史、IP 归属地 | ✅ |
| 在线用户 | 实时在线用户管理、强制踢人 | ✅ |
| 服务监控 | CPU / 内存 / 磁盘 / Go 运行时 | ✅ |
| 缓存监控 | Redis 缓存信息查看 | ✅ |
| 定时任务 | Cron 任务管理、执行日志 | ✅ |
| 文件上传 | 公共文件上传接口 | ✅ |
| WebSocket | 实时消息推送、在线会话 | ✅ |
| 图形验证码 | Redis 存储、动态开关 | ✅ |
| 接口限流 | 基于 juju/ratelimit 的请求限流 | ✅ |
| 签名校验 | 请求 Sign 防篡改 | ✅ |
| 数据导出 | Excel 导出 | ✅ |
| Swagger 文档 | API 在线文档 | ✅ |

---

## 技术栈

### 后端

| 技术 | 版本 | 用途 |
| :--- | :---: | :--- |
| Go | 1.26 | 开发语言 |
| Gin | 1.12 | Web 框架 |
| GORM | 1.31 | ORM 框架 |
| Redis (go-redis) | 9.x | 缓存 / 会话 / 验证码 |
| Zap | 1.28 | 结构化日志 |
| golang-jwt | 5.x | JWT 认证 |
| gorilla/websocket | 1.5 | WebSocket 通信 |
| swaggo | 1.16 | Swagger 文档生成 |
| golang-migrate | 4.20 | 数据库迁移 |
| base64Captcha | 1.3 | 图形验证码 |
| robfig/cron | 3.0 | 定时任务调度 |
| gopsutil | 4.x | 系统监控信息采集 |
| Viper | 1.21 | 配置管理 |
| lumberjack | 2.2 | 日志文件轮转 |

### 前端

| 技术 | 版本 | 用途 |
| :--- | :---: | :--- |
| Vue | 3.5 | 前端框架 |
| TypeScript | 6.0 | 类型系统 |
| Vite | 7.x | 构建工具 |
| Element Plus | 2.11 | UI 组件库 |
| Tailwind CSS | 4.x | 原子化 CSS |
| Pinia | 3.x | 状态管理 |
| Vue Router | 4.x | 路由管理 |
| vue-i18n | 9.x | 国际化 |
| ECharts | 6.x | 数据可视化 |
| Axios | 1.x | HTTP 请求 |
| JSEncrypt | 3.x | RSA 加密 |
| wangEditor | 5.x | 富文本编辑器 |
| xgplayer | 3.x | 视频播放 |
| oxlint / oxfmt | latest | 代码检查与格式化 |

---

## 项目结构

```
ruoyi-go-art-design-pro/
├── server/                     # 后端（Go，基于 ruoyi-go）
│   ├── api/                    # API 接口层（控制器）
│   │   ├── common/             #   通用接口
│   │   ├── monitor/            #   监控模块
│   │   └── system/             #   系统管理
│   ├── configs/                # 配置文件（test / release / docker）
│   ├── data/                   # 数据库初始化 SQL
│   ├── internal/               # 内部模块
│   │   ├── config/             #   配置加载
│   │   ├── constant/           #   常量定义
│   │   ├── db/                 #   数据库 & Redis 初始化
│   │   ├── handler/            #   请求处理器
│   │   ├── initdb/             #   数据库自动初始化
│   │   ├── jobscheduler/       #   定时任务调度器
│   │   ├── middleware/         #   中间件（JWT / CORS / 日志 / 限流 / 权限）
│   │   ├── migrate/            #   数据库迁移
│   │   ├── model/              #   数据模型
│   │   ├── response/           #   统一响应封装
│   │   ├── router/             #   路由注册
│   │   └── service/            #   业务逻辑层
│   ├── migrations/             # golang-migrate 迁移文件
│   ├── pkg/                    # 通用工具包
│   │   ├── captcha/            #   验证码
│   │   ├── exceltool/          #   Excel 操作
│   │   ├── fileuploadtool/     #   文件上传
│   │   ├── jwttool/            #   JWT 工具
│   │   ├── redistool/          #   Redis 工具
│   │   ├── rsatool/            #   RSA 加解密
│   │   ├── websocket/          #   WebSocket Hub
│   │   └── utils/              #   实用工具
│   ├── scripts/                # 构建 & 部署脚本
│   ├── main.go                 # 入口文件
│   ├── go.mod                  # Go 模块定义
│   └── Dockerfile              # 后端 Docker 构建
├── web/                        # 前端（Vue 3 + Art Design Pro 设计体系）
│   ├── src/
│   │   ├── api/                #   API 请求模块
│   │   ├── assets/             #   静态资源
│   │   ├── components/         #   公共组件
│   │   ├── config/             #   应用配置
│   │   ├── directives/         #   自定义指令
│   │   ├── hooks/              #   可组合函数
│   │   ├── locales/            #   国际化语言包
│   │   ├── router/             #   路由配置
│   │   ├── store/              #   Pinia 状态管理
│   │   ├── types/              #   TypeScript 类型定义
│   │   ├── utils/              #   工具函数
│   │   └── views/              #   页面视图
│   ├── vite.config.ts          # Vite 配置
│   ├── tsconfig.json           # TypeScript 配置
│   └── package.json            # 前端依赖
├── deploy/                     # Docker 部署
│   ├── backend/Dockerfile      # 后端镜像
│   └── docker-compose.yml      # 三容器编排
└── docs/                       # 开发文档
```

---

## 快速开始

### 环境要求

| 环境 | 版本要求 |
| :--- | :--- |
| Go | >= 1.22 |
| Node.js | >= 20.19.0 |
| pnpm | >= 8.8.0 |
| MySQL | >= 8.0 |
| Redis | >= 6.0 |

### 后端启动

```bash
# 克隆项目
git clone https://github.com/chen496817468/ruoyi-go-art-design-pro.git
cd ruoyi-go-art-design-pro

# 进入后端目录
cd server

# 安装依赖
go mod tidy

# 修改配置文件（数据库、Redis 连接信息）
# 配置文件位于 configs/test.yaml（开发）或 configs/release.yaml（生产）

# 启动服务（开发模式）
go run main.go -c test

# 启动服务（生产模式）
go run main.go -c release
```

> 首次启动时，程序会自动检测数据库是否已初始化，若为空库则自动导入 `data/ry_init_full.sql` 完成建表和种子数据。

### 前端启动

```bash
# 进入前端目录
cd web

# 安装依赖
pnpm install

# 启动开发服务
pnpm dev

# 构建生产环境
pnpm build

# 代码检查
pnpm lint

# 代码格式化
pnpm fmt
```

### Docker 一键部署

```bash
# 进入部署目录
cd deploy

# 启动所有服务（构建 + 运行）
docker compose up -d --build

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f server

# 访问地址
# http://localhost:8080/
```

---

## 默认账号

| 角色 | 账号 | 密码 |
| :--- | :--- | :--- |
| 超级管理员 | admin | admin123 |

---

## 优势

- **全栈 Go + Vue** — 后端高性能、前端现代化
- **ruoyi-go 后端基座** — 继承 nils-golothm 成熟的 Go 版若依实现，架构清晰
- **Art Design Pro 设计体系** — 区别于传统若依 UI，视觉体验全面升级
- **层级清晰** — Router → Handler → Service → Model 四层架构，职责分明
- **大量便捷封装** — JWT、Redis、验证码、文件上传、Excel 等开箱即用
- **Docker 原生支持** — 三容器编排，一键部署到任何支持 Docker 的环境
- **数据库自动初始化** — 首次启动自动建表，零配置即可运行
- **若依生态兼容** — 功能模块与若依 Java 版高度一致，学习成本低

---

## 致谢

- [ruoyi-go](https://github.com/nils-golothm/ruoyi-go) — Go 版若依后端核心，本项目的后端基座（by nils-golothm）
- [若依 RuoYi](http://ruoyi.vip) — 优秀的后台管理框架
- [Art Design Pro](https://github.com/Daymychen/art-design-pro) — 现代化前端设计体系，本项目前端基于此重写
- [Gin](https://github.com/gin-gonic/gin) — 高性能 Web 框架
- [GORM](https://gorm.io) — Go ORM 框架
- [Vue.js](https://vuejs.org) — 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org) — 企业级 UI 组件库

---

## License

[MIT](LICENSE)
# RuoYi-Go-Art-Design-Pro
