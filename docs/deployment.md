# 部署文档

本项目为 Go + Gin + 原生模板的轻量服务，部署极为简单。以下提供多种方式。

## 前置条件
- Go 1.19 及以上
- 服务器具备外网访问能力（用于首次依赖下载）

## 方式一：直接运行
- 上传项目代码到服务器
- 执行：
  - `go mod tidy`
  - `go run main.go`
- 在防火墙或安全组中放行 `8080` 端口
- 使用反向代理（可选）：将 Nginx 反向代理 `http://127.0.0.1:8080/`

## 方式二：编译二进制
- 本地或服务器上编译：
  - `go build -o resume-to-job`
- 运行：
  - `./resume-to-job`
- 将 `templates/` 与 `static/` 一并部署到同一目录（用于模板与静态资源加载）

## 方式三：Docker
- 创建 `Dockerfile` 示例：
```
FROM golang:1.19-alpine AS build
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o resume-to-job

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/resume-to-job /app/
COPY templates /app/templates
COPY static /app/static
EXPOSE 8080
CMD ["/app/resume-to-job"]
```
- 构建并运行：
  - `docker build -t resume-to-job .`
  - `docker run -p 8080:8080 resume-to-job`

## Nginx 反向代理示例
```
server {
    listen 80;
    server_name example.com;

    location / {
        proxy_pass http://127.0.0.1:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 运行参数与环境变量
- 默认监听端口：`8080`
- 如需改为生产模式：`export GIN_MODE=release`

## 备份与升级
- 模板与静态资源：`templates/`、`static/`
- 二进制与配置：`resume-to-job`、启动脚本（如 systemd）

## systemd 示例
```
[Unit]
Description=ResumeToJob Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/resume-to-job
ExecStart=/opt/resume-to-job/resume-to-job
Restart=on-failure
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

## 安全建议
- 启用 `GIN_MODE=release`
- 在反向代理层做速率限制与访问控制
- 不在日志中输出敏感信息
- 部署时避免公开模板编辑目录的写权限
