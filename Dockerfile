# 使用官方的 Go 镜像作为基础镜像
FROM golang:1.21.4-alpine AS builder

# 安装必要的工具来下载和解压文件
RUN apk add --no-cache curl unzip

# 设置工作目录
WORKDIR /app

# 从 GitHub 下载仓库并解压
RUN curl -L https://gh.api.99988866.xyz/https://github.com/reppoor/telegram-bot-ddns/archive/refs/tags/1.0.0.zip -o telegram-bot-ddns-1.0.0.zip \
    && unzip telegram-bot-ddns-1.0.0.zip \
    && rm telegram-bot-ddns-1.0.0.zip

# 设置安装依赖的变量环境
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 切换到解压后的文件夹，并安装 Go 依赖
WORKDIR /app/telegram-bot-ddns-1.0.0
RUN go mod tidy

# 构建 Go 应用，将输出文件放到 /app/telegram-bot-ddns-1.0.0/cmd 目录下
RUN go build -o /app/telegram-bot-ddns-1.0.0/cmd/main telegram-bot-ddns-1.0.0/cmd/main.go

# 使用更小的镜像作为运行时镜像
FROM alpine:latest
COPY --from=builder /app/telegram-bot-ddns-1.0.0/cmd/main /app/telegram-bot-ddns-1.0.0/cmd/main

# 设置容器启动命令
CMD ["/app/telegram-bot-ddns-1.0.0/cmd/main"]