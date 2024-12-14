# 使用官方的 Go 镜像作为基础镜像
FROM golang:1.21.4-alpine AS builder

# 安装必要的工具来下载和解压文件
RUN apk add --no-cache curl unzip

# 设置工作目录
WORKDIR /app

# 从 GitHub 下载仓库并解压
RUN curl -L https://gh.api.99988866.xyz/https://github.com/reppoor/telegram-bot-ddns/archive/refs/heads/master.zip -o telegram-bot-ddns-master.zip \
    && unzip telegram-bot-ddns-master.zip \
    && rm telegram-bot-ddns-master.zip

# 重命名解压后的文件夹（相对路径）
RUN mv telegram-bot-ddns-master telegrambot

# 设置安装依赖的变量环境
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 切换到解压后的文件夹，并安装 Go 依赖
WORKDIR /app/telegrambot
RUN go mod tidy
RUN ls /app/telegrambot
# 构建 Go 应用
RUN go build -o cmd/main cmd/main.go
# 使用轻量级的基础镜像来运行应用
FROM alpine:latest
# 创建工作目录
WORKDIR /app

# 从构建阶段复制构建产物
COPY --from=builder /app/telegrambot/cmd/main /app/
# 从构建阶段复制构建产物
COPY --from=builder /app/telegrambot/go.mod /app/
# 复制配置文件
COPY --from=builder /app/telegrambot/conf.yaml /app/


# 设置容器启动命令
CMD ["./main"]