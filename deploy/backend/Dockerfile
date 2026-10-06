# ============================================================
# 后端镜像（纯运行环境，无需容器内编译）
# 前后端均使用宿主机预编译产物，直接 COPY 进镜像
#
# 前置步骤：
#   1. 编译后端二进制（在 server/ 目录下执行）：
#      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../deploy/backend/server main.go
#   2. 编译前端产物（在 web/ 目录下执行）：
#      pnpm build   （输出到 server/web-dist）
#
# 构建（在仓库根目录执行）：
#   docker compose -f deploy/docker-compose.yml build server
# ============================================================

FROM alpine:3.20
RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Shanghai
WORKDIR /app
# 后端预编译二进制
COPY deploy/backend/server /app/server
RUN chmod +x /app/server
# 前端编译产物，由后端静态托管（server.web_dist=/app/web-dist）
COPY server/web-dist /app/web-dist
# 上传/日志目录
RUN mkdir -p /app/uploads /app/logs
EXPOSE 8080
# 使用嵌入的 docker 配置（MySQL=mysql / Redis=redis）
CMD ["/app/server", "-c", "docker"]
