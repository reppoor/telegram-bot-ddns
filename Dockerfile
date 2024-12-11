# 使用官方的 Go 镜像作为基础镜像
FROM golang:1.21.4-alpine AS builder

# 设置工作目录
WORKDIR /app

# 将当前目录下的所有文件复制到容器内
COPY . .
# 设置安装依赖的变量环境
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 安装 Go 依赖
RUN go mod tidy
# 切换到 /cmd 目录并构建 Go 应用，将输出文件放到 /app/cmd 目录下
WORKDIR /cmd
RUN go build -o main main.go
# 使用更小的镜像作为运行时镜像
FROM alpine:latest
COPY --from=builder /app/cmd/main /app/cmd/main

# 设置容器启动命令
CMD ["/app/cmd/main"]