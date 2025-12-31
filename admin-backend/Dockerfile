FROM golang:1.23-alpine AS builder

WORKDIR /app

# 设置GOPROXY和更长的超时时间
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
ENV GOSUMDB=off
ENV GOTIMEOUT=120s

# 安装依赖
COPY go.mod go.sum ./
# 添加重试逻辑
RUN apk add --no-cache ca-certificates git && \
    for i in $(seq 1 3); do go mod download && break || sleep 5; done

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o yflow-backend ./cmd/server

# 创建最终的轻量级镜像
FROM alpine:latest

WORKDIR /app

# 安装必要的CA证书 (可能需要HTTPS请求)
RUN apk --no-cache add ca-certificates

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/yflow-backend .

# 创建.env文件存放位置
RUN mkdir -p /app/config
VOLUME /app/config

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./yflow-backend"]