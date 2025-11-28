# Nginx Example

使用 WatchCow 将 Nginx 容器注册为 fnOS 应用。

## 快速开始

```bash
cd examples/nginx
docker compose up -d
```

访问: http://localhost:8080

## WatchCow Labels

| Label | 说明 | 对应 manifest 字段 |
|-------|------|-------------------|
| `watchcow.enable` | 启用发现 (必需) | - |
| `watchcow.appname` | 应用标识 | `appname` |
| `watchcow.display_name` | 显示名称 | `display_name` |
| `watchcow.desc` | 应用描述 | `desc` |
| `watchcow.version` | 版本号 | `version` |
| `watchcow.maintainer` | 维护者 | `maintainer` |
| `watchcow.service_port` | 服务端口 | `service_port` |
| `watchcow.protocol` | 协议 (http/https) | UI config |
| `watchcow.path` | URL 路径 | UI config |
| `watchcow.ui_type` | UI 类型 (url/iframe) | UI config |
| `watchcow.icon` | 图标 URL | 下载到应用包 |

## 验证

```bash
# 查看容器 labels
docker inspect nginx-demo --format '{{json .Config.Labels}}' | jq .

# WatchCow 会自动:
# 1. 发现带有 watchcow.enable=true 的容器
# 2. 生成 fnOS 应用包
# 3. 使用 appcenter-cli install-local 安装
```
