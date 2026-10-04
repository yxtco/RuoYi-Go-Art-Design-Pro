# ============================================================
# 后端镜像（自包含：同时构建并托管前端编译产物，无需 Nginx）
# 阶段一：Node 编译前端（VITE_API_URL=/ 同源直连后端）
# 阶段二：Go 编译后端二进制
# 阶段三：轻量运行镜像
#
# 构建（在仓库根目录执行）：
#   docker compose -f deploy/docker-compose.yml build server
# ============================================================

# ---------- 阶段一：前端编译 ----------
FROM node:20-alpine AS fe
# 国内镜像源
RUN npm config set registry https://registry.npmmirror.com
WORKDIR /fe
RUN corepack enable && corepack prepare pnpm@10.14.0 --activate
# 先复制依赖清单，利用层缓存
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
# 前端与后端同源部署：API/WebSocket 直接以根路径请求后端
ARG VITE_API_URL=/
ENV VITE_API_URL=$VITE_API_URL
RUN pnpm build

# ---------- 阶段二：后端编译 ----------
FROM golang:1.26-alpine AS be
# 国内代理，加快依赖拉取
RUN go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /be
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
# 编译静态二进制（含嵌入配置 docker.yaml / swagger docs）
RUN CGO_ENABLED=0 go build -o /out/server main.go

# ---------- 阶段三：运行 ----------
FROM alpine:3.20
RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Shanghai
WORKDIR /app
# 后端二进制
COPY --from=be /out/server /app/server
# 前端编译产物，由后端静态托管（server.web_dist=/app/web-dist）
COPY --from=fe /fe/dist /app/web-dist
# 上传/日志目录
RUN mkdir -p /app/uploads /app/logs
EXPOSE 8080
# 使用嵌入的 docker 配置（MySQL=mysql / Redis=redis）
CMD ["/app/server", "-c", "docker"]
