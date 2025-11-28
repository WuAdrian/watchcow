# Memos Example

使用 WatchCow 将 Memos 容器注册为 fnOS 应用。

## 关于 Memos

Memos 是一个轻量级的开源笔记工具。

- 官网: https://usememos.com
- GitHub: https://github.com/usememos/memos

## 快速开始

```bash
cd examples/memos
docker compose up -d
```

访问: http://localhost:5230

首次访问需要创建管理员账号。

## 数据持久化

数据存储在 `./data` 目录，备份此目录即可保护数据。

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
docker inspect memos --format '{{json .Config.Labels}}' | jq .
```
