# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装基本依赖
RUN apk add --no-cache git ca-certificates build-base

# 设置工作目录
WORKDIR /app

# 首先复制依赖相关文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制所有源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/apiserver ./cmd/apiserver

# 运行阶段
FROM alpine:3.19

# 安装 CA 证书和基本工具
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN adduser -D -g '' appuser

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/apiserver /usr/local/bin/

# 创建配置目录
RUN mkdir -p /etc/go-protoc

# 复制配置文件
COPY --from=builder /app/configs/config.yaml /etc/go-protoc/config.yaml

# 设置工作目录
WORKDIR /home/appuser

# 切换到非 root 用户
USER appuser

# 设置环境变量
ENV CONFIG_PATH=/etc/go-protoc/config.yaml

# 暴露 HTTP 和 gRPC 端口
EXPOSE 8090 8091

# 启动应用
ENTRYPOINT ["apiserver"]